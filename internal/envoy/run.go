// Copyright 2019 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package envoy

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/tetratelabs/func-e/internal/moreos"
)

// Run execs the binary at the path with the args passed. It is a blocking function that can be shutdown via ctx.
//
// This will exit either `ctx` is done, or the process exits.
func (r *Runtime) Run(ctx context.Context, args []string) error {
	// We can't use CommandContext even if that seems correct here. The reason is that we need to invoke shutdown hooks,
	// and they expect the process to still be running. For example, this allows admin API hooks.
	cmd := exec.Command(r.opts.EnvoyPath, args...) // #nosec -> users can run whatever binary they like!
	cmd.Stdout = r.Out
	cmd.Stderr = r.Err
	cmd.SysProcAttr = moreos.ProcessGroupAttr()
	r.cmd = cmd

	if err := r.ensureAdminAddressPath(); err != nil {
		return err
	}

	// Print the binary path to the user for debugging purposes.
	moreos.Fprintf(r.Out, "starting: %s with --admin-address-path %s\n", r.opts.EnvoyPath, r.adminAddressPath) //nolint
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to start Envoy process: %w", err)
	}

	// Warn, but don't fail if we can't write the pid file for some reason
	r.maybeWarn(os.WriteFile(r.pidPath, []byte(strconv.Itoa(cmd.Process.Pid)), 0o600))

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
		// When original context is done, we need to shutdown the process by ourselves.
		// Run the shutdown hooks and wait for them to complete.
		r.handleShutdown()
		// Then wait for the process to exit.
		<-cmdExitWait.Done()
	case <-cmdExitWait.Done():
	}

	// Warn, but don't fail on error archiving the run directory
	if !r.opts.DontArchiveRunDir {
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
			moreos.Fprintf(r.Out, "discovered admin address: %s\n", adminAddress)
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
}
