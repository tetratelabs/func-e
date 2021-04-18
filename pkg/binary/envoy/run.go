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
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/tetratelabs/log"

	"github.com/tetratelabs/getenvoy/pkg/manifest"
)

// Run execs the binary defined by the key with the args passed
// It is a blocking function that can only be terminated via SIGINT
func (r *Runtime) Run(key *manifest.Key, args []string) error {
	path := filepath.Join(r.platformDirectory(key), envoyLocation)
	return r.RunPath(path, args)
}

// RunPath execs the binary at the path with the args passed
// It is a blocking function that can only be terminated via SIGINT
func (r *Runtime) RunPath(path string, args []string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("unable to stat %q: %v", path, err)
	}

	// We can't use CommandContext even if that seems correct here. The reason is that we need to invoke preTerminate
	// handlers, and they expect the process to still be running. For example, this allows admin API hooks.
	cmd := exec.Command(path, args...) // #nosec -> users can run whatever binary they like!
	cmd.Dir = r.WorkingDir
	cmd.Stdout = r.IO.Out
	cmd.Stderr = r.IO.Err
	cmd.SysProcAttr = sysProcAttr()
	r.cmd = cmd

	err := r.handlePreStart()
	if err != nil {
		return err
	}

	if cmd.Stdout == nil {
		cmd.Stdout = os.Stdout
	}
	if cmd.Stderr == nil {
		cmd.Stderr = os.Stderr
	}

	log.Infof("Envoy command: %v", cmd.Args)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to start Envoy process: %w", err)
	}

	waitCtx, waitCancel := context.WithCancel(context.Background())
	sigCtx, sigCancel := signal.NotifyContext(waitCtx, syscall.SIGINT, syscall.SIGTERM)
	defer waitCancel()
	r.fakeInterrupt = sigCancel

	go waitForExit(cmd, waitCancel) // waits in a goroutine. We may need to kill the process if a signal occurs first.

	// Block until we receive SIGINT or are canceled because Envoy has died
	<-sigCtx.Done()

	if cmd.ProcessState != nil {
		log.Infof("Envoy process (PID=%d) terminated prematurely", cmd.Process.Pid)
		return r.handlePostTermination()
	}

	r.handleTermination()

	// Block until the process is complete. This ensures file descriptors are closed.
	<-waitCtx.Done()

	return r.handlePostTermination()
}

// DebugStore returns the location at which the runtime instance persists debug data for this given instance
// Getters typically aren't idiomatic Go, however, this one is deliberately part of the runner interface
func (r *Runtime) DebugStore() string {
	return r.debugDir
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

func (r *Runtime) initializeDebugStore() error {
	r.debugDir = filepath.Join(r.store, "debug", strconv.FormatInt(time.Now().UnixNano(), 10))
	return os.MkdirAll(r.debugDir, 0750)
}
