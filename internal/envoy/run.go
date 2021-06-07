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
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// Run execs the binary at the path with the args passed. It is a blocking function that can be shutdown via SIGINT.
func (r *Runtime) Run(ctx context.Context, args []string) (err error) {
	// We can't use CommandContext even if that seems correct here. The reason is that we need to invoke shutdown hooks,
	// and they expect the process to still be running. For example, this allows admin API hooks.
	cmd := exec.Command(r.opts.EnvoyPath, args...) // #nosec -> users can run whatever binary they like!
	cmd.Stdout = r.Out
	cmd.Stderr = r.Err
	cmd.SysProcAttr = sysProcAttr()
	r.cmd = cmd

	// Suppress any error and replace it with the envoy exit status when > 1
	defer func() {
		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() > 0 {
			if err != nil {
				fmt.Fprintln(r.Out, "warning:", err) //nolint
			}
			err = fmt.Errorf("envoy exited with status: %d", cmd.ProcessState.ExitCode())
		}
	}()

	if err := r.ensureAdminAddressPath(); err != nil {
		return err
	}

	// Print the process line to the console for user knowledge and parsing convenience
	fmt.Fprintln(r.Out, "starting:", strings.Join(r.cmd.Args, " ")) //nolint
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to start Envoy process: %w", err)
	}

	waitCtx, waitCancel := context.WithCancel(ctx)
	sigCtx, sigCancel := signal.NotifyContext(waitCtx, syscall.SIGINT, syscall.SIGTERM)
	defer waitCancel()
	r.FakeInterrupt = sigCancel

	// Wait in a goroutine. We may need to kill the process if a signal occurs first.
	go func() {
		defer waitCancel()
		_ = r.cmd.Wait() // Envoy logs like "caught SIGINT" or "caught ENVOY_SIGTERM", so we don't repeat logging here.
	}()

	awaitAdminAddress(sigCtx, r)

	// Block until we receive SIGINT or are canceled because Envoy has died
	<-sigCtx.Done()
	if cmd.ProcessState != nil && !r.opts.DontArchiveRunDir {
		return r.archiveRunDir()
	}

	r.handleShutdown(ctx)

	// Block until the process is complete. This ensures file descriptors are closed.
	<-waitCtx.Done()

	return r.archiveRunDir()
}

// awaitAdminAddress waits up to 2 seconds for the admin address to be available and logs it.
// See https://github.com/envoyproxy/envoy/issues/16050 for moving this logging upstream.
func awaitAdminAddress(sigCtx context.Context, r *Runtime) {
	for i := 0; i < 10 && sigCtx.Err() == nil; i++ {
		adminAddress, adminErr := r.GetAdminAddress()
		if adminErr == nil {
			fmt.Fprintln(r.Out, "discovered admin address:", adminAddress) //nolint
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
}
