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

	"github.com/mholt/archiver"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
	"github.com/tetratelabs/log"
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

	ctx, cancel := context.WithCancel(context.Background())
	r.ctx = ctx

	// #nosec -> users can run whatever binary they like!
	r.cmd = exec.Command(path, args...)
	r.cmd.Dir = r.WorkingDir
	r.cmd.Stdout = r.IO.Out
	r.cmd.Stderr = r.IO.Err
	r.cmd.SysProcAttr = sysProcAttr()

	r.handlePreStart()

	go r.runEnvoy(cancel)

	r.waitForTerminationSignals()
	r.handleTermination()
	cancel() // this should never actually do anything but lets avoid any future context leaks

	// Block until the Envoy process and termination handler are finished cleaning up
	r.wg.Wait()

	// Tar up the debug data and clean up
	if err := archiver.Archive([]string{r.DebugStore()}, r.DebugStore()+".tar.gz"); err != nil {
		return fmt.Errorf("unable to archive debug store directory %v: %v", r.DebugStore(), err)
	}
	return os.RemoveAll(r.DebugStore())
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

// RegisterWait informs the runtime it needs to wait for you to complete
// It is a wrapper around sync.WaitGroup.Add()
func (r *Runtime) RegisterWait(delta int) {
	r.wg.Add(delta)
}

// RegisterDone informs the runtime that you have completed
// It is a wrapper around sync.WaitGroup.Done()
func (r *Runtime) RegisterDone() {
	r.wg.Done()
}

// AppendArgs appends the passed args to the child process' args
func (r *Runtime) AppendArgs(args []string) {
	r.cmd.Args = append(r.cmd.Args, args...)
}

func (r *Runtime) waitForTerminationSignals() {
	signal.Notify(r.signals, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive SIGINT or are canceled because Envoy has died
	select {
	case <-r.ctx.Done():
		log.Infof("No Envoy processes remaining, terminating GetEnvoy process (PID=%d)", os.Getpid())
		return
	case <-r.signals:
		log.Infof("GetEnvoy process (PID=%d) received SIGINT", os.Getpid())
		return
	}
}

func (r *Runtime) runEnvoy(cancel context.CancelFunc) {
	if r.cmd.Stdout == nil {
		r.cmd.Stdout = os.Stdout
	}
	if r.cmd.Stderr == nil {
		r.cmd.Stderr = os.Stderr
	}

	r.wg.Add(1)
	defer r.wg.Done()
	defer cancel()

	log.Infof("Envoy command: %v", r.cmd.Args)
	if err := r.cmd.Start(); err != nil {
		log.Errorf("Unable to start Envoy process: %v", err)
		return
	}

	if err := r.cmd.Wait(); err != nil {
		if r.cmd.ProcessState.ExitCode() == -1 {
			log.Infof("Envoy process (PID=%d) terminated via %v", r.cmd.Process.Pid, err)
		} else {
			log.Infof("Envoy process (PID=%d) terminated with an error: %v", r.cmd.Process.Pid, err)
		}
	}
}

func (r *Runtime) initializeDebugStore() error {
	r.debugDir = filepath.Join(r.store, "debug", strconv.FormatInt(time.Now().UnixNano(), 10))
	return os.MkdirAll(r.debugDir, 0750)
}
