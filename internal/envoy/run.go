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
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/tetratelabs/func-e/internal/moreos"
)

// Run execs the binary at the path with the args passed. It is a blocking function that can be shutdown via SIGINT.
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

	// Print the process line to the console for user knowledge and parsing convenience
	moreos.Fprintf(r.Out, "starting: %s\n", strings.Join(r.cmd.Args, " ")) //nolint
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to start Envoy process: %w", err)
	}

	// Warn, but don't fail if we can't write the pid file for some reason
	r.maybeWarn(os.WriteFile(r.pidPath, []byte(strconv.Itoa(cmd.Process.Pid)), 0o600))

	waitCtx, waitCancel := context.WithCancel(ctx)
	defer waitCancel()

	sigCtx, sigCancel := signal.NotifyContext(waitCtx, syscall.SIGINT, syscall.SIGTERM)
	defer sigCancel()
	r.FakeInterrupt = sigCancel

	// Wait in a goroutine. We may need to kill the process if a signal occurs first.
	go func() {
		defer waitCancel()
		_ = r.cmd.Wait() // Envoy logs like "caught SIGINT" or "caught ENVOY_SIGTERM", so we don't repeat logging here.
	}()

	awaitAdminAddress(sigCtx, r)

	// Block until we receive SIGINT or are canceled because Envoy has died
	<-sigCtx.Done()

	// The process could have exited due to incorrect arguments or otherwise.
	// If it is still running, run shutdown hooks and propagate the interrupt.
	if cmd.ProcessState == nil {
		r.handleShutdown(ctx)
	}

	// At this point, shutdown hooks have run and Envoy is interrupted.
	// Block until it exits to ensure file descriptors are closed prior to archival.
	// Allow up to 5 seconds for a clean stop, killing if it can't for any reason.
	select {
	case <-waitCtx.Done(): // cmd.Wait goroutine finished
	case <-time.After(5 * time.Second):
		_ = moreos.EnsureProcessDone(r.cmd.Process)
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
			moreos.Fprintf(r.Out, "discovered admin address: %s\n", adminAddress) //nolint
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
}
