// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Run execs the Envoy binary at the path with the args passed.
//
// On success, this blocks and returns nil when either `ctx` is done, or the
// process exits with status zero.
func (r *Runtime) Run(ctx context.Context, args []string) error {
	// We require the admin server, so ensure it exists, and we can read its listener via a file path.
	var err error
	r.adminAddressPath, args, err = ensureAdminAddress(r.logf, r.o.RunDir, args)
	if err != nil {
		return err
	}

	// Append the run directory to args for an easy lookup of where pid files etc are stored.
	// Why? MacOS SIP restricts cross-process env var access: we need a solution that works with both Linux and MacOS.
	args = append(args, "--", "--func-e-run-dir", r.o.RunDir)

	cmd := exec.CommandContext(ctx, r.o.EnvoyPath, args...) // #nosec -> users can run whatever binary they like!
	cmd.Stdout = r.Out
	cmd.SysProcAttr = processGroupAttr()

	// Create a pipe to capture stderr and forward to r.Err
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("unable to create stderr pipe: %w", err)
	}

	r.cmd = cmd

	// Print the binary and run directory to the user for debugging purposes.
	r.logf("starting: %s in run directory %s", r.o.EnvoyPath, r.o.RunDir)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to start Envoy process: %w", err)
	}

	// Warn, but don't fail if we can't write the pid file for some reason
	r.maybeWarn(os.WriteFile(filepath.Join(r.o.RunDir, "envoy.pid"), []byte(strconv.Itoa(cmd.Process.Pid)), 0o600))

	hookErrCh := make(chan error, 1)

	// Process stderr in a goroutine
	go r.processStderr(ctx, stderrPipe, hookErrCh)

	// Wait for the process, and any stderr processing, to complete
	exitErr := cmd.Wait()
	hookErr := <-hookErrCh

	// First, check for startup hook errors
	if hookErr != nil {
		return hookErr
	}

	// Next, handle process exit errors
	if exitErr != nil {
		// Only ignore exit errors on cancellation if there was no stderr error
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil
		}
		return exitErr
	}

	return nil // Clean exit
}

// processStderr scans stderr output and triggers the startup hook when Envoy is ready.
func (r *Runtime) processStderr(ctx context.Context, stderrPipe io.Reader, hookErrCh chan<- error) {
	var hookErr error
	defer func() {
		if p := recover(); p != nil {
			if hookErr == nil {
				hookErr = fmt.Errorf("processStderr panicked: %v", p)
			}
			r.logf("processStderr panicked: %v", p)
		}
		hookErrCh <- hookErr
	}()

	scanner := bufio.NewScanner(stderrPipe)
	hookTriggered := false

	for scanner.Scan() {
		line := scanner.Text()
		// Copy stderr line to the output writer
		fmt.Fprintln(r.Err, line) //nolint:errcheck

		// Trigger startup hook when admin is ready
		if !hookTriggered && strings.Contains(line, "starting main dispatch loop") {
			hookTriggered = true
			adminAddrBytes, err := os.ReadFile(r.adminAddressPath)
			if err != nil {
				hookErr = fmt.Errorf("failed to read admin address from %s: %w", r.adminAddressPath, err)
				r.logf(hookErr.Error())
				break
			}
			adminAddress := strings.TrimSpace(string(adminAddrBytes))
			r.adminAddress = adminAddress

			// Call startup hook
			if err := r.startupHook(ctx, r.o.RunDir, adminAddress); err != nil {
				hookErr = err
				r.logf(err.Error())
				break
			}
		}
	}
	// ignore scanner errors as we are only concerned in hook errors
}
