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
	"io"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/tetratelabs/log"
)

// Run execs the binary at the path with the args passed. It is a blocking function that can be terminated via SIGINT.
func (r *Runtime) Run(ctx context.Context, args []string) error {
	// We can't use CommandContext even if that seems correct here. The reason is that we need to invoke preTerminate
	// handlers, and they expect the process to still be running. For example, this allows admin API hooks.
	cmd := exec.Command(r.opts.EnvoyPath, args...) // #nosec -> users can run whatever binary they like!
	cmd.Dir = r.opts.WorkingDir
	cmd.Stdout = r.Out
	cmd.Stderr = r.Err
	cmd.SysProcAttr = sysProcAttr()
	r.cmd = cmd

	err := r.handlePreStart()
	if err != nil {
		return err
	}

	// Log for the information of users and also for us to have a reliable parsing in e2e tests.
	log.Infof("cd %s\n", cmd.Dir)
	log.Infof("%s %s\n", r.opts.EnvoyPath, strings.Join(args, " "))
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to start Envoy process: %w", err)
	}

	waitCtx, waitCancel := context.WithCancel(ctx)
	sigCtx, sigCancel := signal.NotifyContext(waitCtx, syscall.SIGINT, syscall.SIGTERM)
	defer waitCancel()
	r.FakeInterrupt = sigCancel

	go waitForExit(cmd, waitCancel) // waits in a goroutine. We may need to kill the process if a signal occurs first.

	awaitAdminAddress(sigCtx, r)

	// Block until we receive SIGINT or are canceled because Envoy has died
	<-sigCtx.Done()

	if cmd.ProcessState != nil {
		return r.handlePostTermination()
	}

	r.handleTermination()

	// Block until the process is complete. This ensures file descriptors are closed.
	<-waitCtx.Done()

	return r.handlePostTermination()
}

// awaitAdminAddress waits up to 2 seconds for the admin address to be available and logs it.
// See https://github.com/envoyproxy/envoy/issues/16050 for moving this logging upstream.
func awaitAdminAddress(sigCtx context.Context, r *Runtime) {
	for i := 0; i < 10 && sigCtx.Err() == nil; i++ {
		adminAddress, adminErr := r.GetAdminAddress()
		if adminErr == nil {
			log.Infof("discovered admin address: %v", adminAddress)
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// SetStdout writes the stdout of Envoy to the passed writer
func (r *Runtime) SetStdout(fn func(io.Writer) io.Writer) {
	r.cmd.Stdout = fn(r.cmd.Stdout)
}

// SetStderr writes the stderr of Envoy to the passed writer
func (r *Runtime) SetStderr(fn func(io.Writer) io.Writer) {
	r.cmd.Stderr = fn(r.cmd.Stderr)
}

// AppendArgs appends the passed args to the child process' args
func (r *Runtime) AppendArgs(args []string) {
	r.cmd.Args = append(r.cmd.Args, args...)
}

func waitForExit(cmd *exec.Cmd, cancel context.CancelFunc) {
	defer cancel()
	if err := cmd.Wait(); err != nil {
		if cmd.ProcessState.ExitCode() == -1 {
			log.Infof("Envoy process (PID=%d) terminated via %v", cmd.Process.Pid, err)
		} else {
			log.Infof("Envoy process (PID=%d) terminated with an error: %v", cmd.Process.Pid, err)
		}
	}
}
