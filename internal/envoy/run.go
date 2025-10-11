// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/tetratelabs/func-e/internal/admin"
)

// Run execs the Envoy binary at the path with the args passed.
//
// On success, this blocks and returns nil when either `ctx` is done, or the
// process exits with status zero.
func (r *Runtime) Run(ctx context.Context, args []string) error {
	// We require the admin server, so ensure it exists, and we can read its listener via a file path.
	var err error
	adminAddressPath, args, err := ensureAdminAddress(r.logf, r.o.RunDir, args)
	if err != nil {
		return err
	}

	// Append the run directory to args for an easy lookup of where pid files etc are stored.
	// Why? MacOS SIP restricts cross-process env var access: we need a solution that works with both Linux and MacOS.
	args = append(args, "--", "--func-e-run-dir", r.o.RunDir)

	cmd := exec.CommandContext(ctx, r.o.EnvoyPath, args...) // #nosec -> users can run whatever binary they like!
	cmd.Stdout = r.Out
	cmd.Stderr = r.Err
	cmd.SysProcAttr = processGroupAttr()

	r.cmd = cmd

	// Print the binary and run directory to the user for debugging purposes.
	r.logf("starting: %s in run directory %s", r.o.EnvoyPath, r.o.RunDir)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to start Envoy process: %w", err)
	}

	// Warn, but don't fail if we can't write the pid file for some reason
	r.maybeWarn(os.WriteFile(filepath.Join(r.o.RunDir, "envoy.pid"), []byte(strconv.Itoa(cmd.Process.Pid)), 0o600))

	hookErrCh := make(chan error, 1)

	// Create a context that's cancelled when Envoy process exits
	monitorCtx, cancelMonitor := context.WithCancel(ctx)
	defer cancelMonitor()

	// Monitor admin readiness and trigger startup hook in a goroutine
	go func() {
		defer func() {
			if p := recover(); p != nil {
				hookErrCh <- fmt.Errorf("startup hook panicked: %v", p)
				r.logf("startup hook panicked: %v", p)
			}
		}()

		var err error
		adminClient, err := admin.NewAdminClient(monitorCtx, r.o.RunDir, adminAddressPath)
		if err != nil {
			// If we can't create the admin client, it likely means Envoy failed to start
			// Don't log or return error here - let cmd.Wait() handle the exit error
			hookErrCh <- nil
			return
		}

		// StartupHook's precondition is the admin server being ready.
		if err = adminClient.AwaitReady(monitorCtx, 100*time.Millisecond); err == nil {
			err = r.startupHook(monitorCtx, adminClient)
		}

		// Report real errors; ignore context cancellation (clean shutdown)
		if err != nil && !errors.Is(err, context.Canceled) {
			r.logf(err.Error())
			hookErrCh <- err
		} else {
			hookErrCh <- nil
		}
	}()

	// Wait for the process, and any admin monitoring, to complete
	exitErr := cmd.Wait()
	cancelMonitor() // Stop monitoring immediately when process exits
	hookErr := <-hookErrCh

	// Prioritize hook errors - if the hook ran and failed, that's the most relevant error
	if hookErr != nil {
		return hookErr
	}

	// Ignore exit errors on clean cancellation (user Ctrl-C, etc.)
	if exitErr == nil || errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}
	return exitErr
}
