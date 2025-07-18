// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// Run execs the binary at the path with the args passed. It is a blocking function that can be shutdown via ctx.
//
// This will exit either `ctx` is done, or the process exits.
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

	// We can't use CommandContext even if that seems correct here. The reason is that we need to invoke shutdown hooks,
	// and they expect the process to still be running. For example, this allows admin API hooks.
	cmd := exec.Command(r.o.EnvoyPath, args...) // #nosec -> users can run whatever binary they like!
	cmd.Stdout = r.Out
	cmd.Stderr = r.Err
	cmd.SysProcAttr = processGroupAttr()
	r.cmd = cmd

	// Print the binary path to the user for debugging purposes.
	r.logf("starting: %s with --admin-address-path %s\n", r.o.EnvoyPath, r.adminAddressPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to start Envoy process: %w", err)
	}

	// Warn, but don't fail if we can't write the pid file for some reason
	r.maybeWarn(os.WriteFile(filepath.Join(r.o.RunDir, "envoy.pid"), []byte(strconv.Itoa(cmd.Process.Pid)), 0o600))

	// Wait in a goroutine. We may need to kill the process if a signal occurs first.
	//
	// Note: do not wrap the original context, otherwise "<-cmdExitWait.Done()" won't block until the process exits
	// if the original context is done.
	cmdExitWait, cmdExit := context.WithCancel(context.Background())
	defer cmdExit()
	go func() {
		defer cmdExit()
		_ = r.cmd.Wait()
	}()

	awaitAdminAddress(ctx, r)

	// Block until the process exits or the original context is done.
	select {
	case <-ctx.Done():
		// When original context is done, we need to shut down the process by ourselves.
		// Run the shutdown hooks and wait for them to complete.
		r.handleShutdown()
		// Then wait for the process to exit.
		<-cmdExitWait.Done()
	case <-cmdExitWait.Done():
		// Process exited naturally
	}

	// Warn, but don't fail on error archiving the run directory
	if !r.o.DontArchiveRunDir {
		r.maybeWarn(r.archiveRunDir())
	}

	if cmd.ProcessState.ExitCode() > 0 {
		return fmt.Errorf("envoy exited with status: %d", cmd.ProcessState.ExitCode())
	}
	return nil
}

// awaitAdminAddress waits up to 2 seconds for the admin address to be available and logs it.
// See https://github.com/envoyproxy/envoy/issues/16050 for moving this logging upstream.
func awaitAdminAddress(sigCtx context.Context, r *Runtime) {
	for i := 0; i < 10 && sigCtx.Err() == nil; i++ {
		adminAddress, adminErr := r.GetAdminAddress()
		if adminErr == nil {
			fmt.Fprintf(r.Out, "discovered admin address: %s\n", adminAddress) //nolint:errcheck
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
}
