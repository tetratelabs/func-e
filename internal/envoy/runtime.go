// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	internalapi "github.com/tetratelabs/func-e/internal/api"
	"github.com/tetratelabs/func-e/internal/envoy/config"
	"github.com/tetratelabs/func-e/internal/globals"
)

type LogFunc func(format string, a ...any)

const (
	configYamlFlag       = `--config-yaml`
	adminEphemeralConfig = "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}"
	adminAddressPathFlag = `--admin-address-path`
)

// NewRuntime creates a new Runtime that runs envoy with the given options.
// opts allows a user running envoy to control directories and hooks.
func NewRuntime(opts *globals.RunOpts, logf LogFunc) *Runtime {
	// Use user-provided hook if set, otherwise use default
	var hook internalapi.StartupHook
	if opts.StartupHook != nil {
		hook = opts.StartupHook
	} else {
		// Capture runDir in closure for config_dump collection
		runDir := opts.RunDir
		safeHook := &safeStartupHook{
			delegate: func(ctx context.Context, adminClient internalapi.AdminClient, runID string) error {
				return collectConfigDump(ctx, http.DefaultClient, adminClient, runDir)
			},
			logf:    logf,
			timeout: 3 * time.Second,
		}
		hook = safeHook.Hook
	}
	return &Runtime{o: opts, logf: logf, startupHook: hook}
}

// Runtime manages an Envoy lifecycle
type Runtime struct {
	o *globals.RunOpts

	cmd              *exec.Cmd
	Out, Err         io.Writer
	OutFile, ErrFile *os.File

	logf LogFunc

	startupHook internalapi.StartupHook
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

// ensureAdminAddress ensures there is an admin server in the args adds
// configYamlFlag of adminEphemeralConfig if there is none. Next, we add
// adminAddressPathFlag if not already set. This allows reading back the admin
// address later regardless of whether the admin server is ephemeral or not.
//
// Note: If adminAddressPathFlag is backfilled, it will be to the
// globals.RunOpts RunDir, which is mutable.
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
	if adminAddress, err := config.FindAdminAddressFromArgs(args); err != nil {
		logf("failed to find admin address: %s", err)
	} else if adminAddress == "" {
		logf("configuring ephemeral admin server")
		args = append(args, configYamlFlag, adminEphemeralConfig)
	}

	if adminAddressPath == "" {
		// Envoy's run directory is mutable, so it is fine to write the admin address there.
		adminAddressPath = filepath.Join(runDir, "admin-address.txt")
		args = append(args, adminAddressPathFlag, adminAddressPath)
	}
	return adminAddressPath, args, nil
}
