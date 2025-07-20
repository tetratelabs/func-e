// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/tetratelabs/func-e/internal/envoy/config"
	"github.com/tetratelabs/func-e/internal/globals"
)

type LogFunc func(format string, a ...any)

// StartupHook runs just after Envoy logs "starting main dispatch loop".
//
// This is useful for callers who need access to two non-deterministic values:
//
// 1. The run directory (where stdout, stderr and the pid file are written)
// 2. The admin address (which is possibly ephemeral)
//
// ## Implementation Notes
//
// Startup hooks are considered mandatory and will stop the run with error if
// failed. If your hook is optional, rescue panics and log your own errors.
//
// Startup hooks run on the goroutine that consumes Envoy's STDERR. Envoy
// doesn't write a lot to stderr, so short tasks won't fill up the pipe and
// cause Envoy to block. However, if your hook is long-running, it must be
// run in a goroutine.
type StartupHook func(ctx context.Context, runDir, adminAddress string) error

const (
	configYamlFlag       = `--config-yaml`
	adminEphemeralConfig = "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}"
	adminAddressPathFlag = `--admin-address-path`
)

// NewRuntime creates a new Runtime that runs envoy in globals.RunOpts RunDir
// opts allows a user running envoy to control the working directory by ID or path, allowing explicit cleanup.
func NewRuntime(opts *globals.RunOpts, logf LogFunc) *Runtime {
	safeHook := &safeStartupHook{
		delegate: collectConfigDump,
		logf:     logf,
		timeout:  3 * time.Second,
	}
	return &Runtime{o: opts, logf: logf, startupHook: safeHook.Hook}
}

// Runtime manages an Envoy lifecycle
type Runtime struct {
	o *globals.RunOpts

	cmd              *exec.Cmd
	Out, Err         io.Writer
	OutFile, ErrFile *os.File

	logf LogFunc

	adminAddress, adminAddressPath string
	startupHook                    StartupHook
}

// String is only used in tests. It is slow, but helps when debugging CI failures
func (r *Runtime) String() string {
	exitStatus := -1
	if r.cmd != nil && r.cmd.ProcessState != nil {
		if ws, ok := r.cmd.ProcessState.Sys().(syscall.WaitStatus); ok {
			exitStatus = ws.ExitStatus()
		}
	}

	return fmt.Sprintf("{exitStatus: %d}", exitStatus)
}

// GetRunDir returns the run-specific directory files can be written to.
func (r *Runtime) GetRunDir() string {
	return r.o.RunDir
}

// maybeWarn writes a warning message to Runtime.Out when the error isn't nil
func (r *Runtime) maybeWarn(err error) {
	if err != nil {
		r.logf("warning: %s", err)
	}
}

// ensureAdminAddress ensures there is an admin server in the args adds configYamlFlag of adminEphemeralConfig if there
// is none. Next, we add adminAddressPathFlag if not already set. This allows reading back the admin address later
// regardless of whether the admin server is ephemeral or not.
//
// Note: If adminAddressPathFlag is backfilled, it will be to the globals.RunOpts RunDir, which is mutable.
func ensureAdminAddress(logf LogFunc, runDir string, argsIn []string) (string, []string, error) {
	args := argsIn
	var hasConfig bool
	var adminAddressPath string
ARGS:
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-c", "--config-path", configYamlFlag:
			i++
			if i < len(args) {
				if args[i] != "" {
					hasConfig = true
					continue
				}
			}
			break ARGS
		case adminAddressPathFlag:
			i++
			if i >= len(args) || args[i] == "" {
				return "", args, fmt.Errorf("missing value to argument %q", adminAddressPathFlag)
			}
			adminAddressPath = args[i]
			continue
		}
	}
	if !hasConfig {
		return "", args, nil // allow envoy to fail
	}

	// We backfill an ephemeral admin server only when we can verify for sure there is none.
	if adminAddress, err := config.FindAdminAddress(args); err != nil {
		logf("failed to find admin address: %s", err)
	} else if adminAddress == "" {
		logf("configuring ephemeral admin server")
		args = append(args, configYamlFlag, adminEphemeralConfig)
	}

	// TODO: remove admin address path requirement for non-ephemeral configs
	if adminAddressPath == "" {
		// Envoy's run directory is mutable, so it is fine to write the admin address there.
		adminAddressPath = filepath.Join(runDir, "admin-address.txt")
		args = append(args, adminAddressPathFlag, adminAddressPath)
	}
	return adminAddressPath, args, nil
}

// GetAdminAddress returns the current admin address in host:port format, or empty if not yet available.
// Exported for admin data collection functionality.
func (r *Runtime) GetAdminAddress() (string, error) {
	if r.adminAddress != "" { // We don't expect the admin address to change once written, so cache it.
		return r.adminAddress, nil
	}
	adminAddress, err := os.ReadFile(r.adminAddressPath)
	if err != nil {
		return "", fmt.Errorf("unable to read %s: %w", r.adminAddressPath, err)
	}
	if _, _, err := net.SplitHostPort(string(adminAddress)); err != nil {
		return "", fmt.Errorf("invalid admin address in %s: %w", r.adminAddressPath, err)
	}
	r.adminAddress = string(adminAddress)
	return r.adminAddress, nil
}
