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
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/tetratelabs/getenvoy/pkg/manifest"
	"github.com/tetratelabs/log"
)

// Run execs the binary defined by the key with the args passed
// It is a blocking function that can only be terminated via SIGINT
func (r *Runtime) Run(key *manifest.Key, args []string) error {
	path := filepath.Join(r.binaryPath(key), "envoy")
	return r.RunPath(path, args)
}

// RunPath execs the binary at the path with the args passed
// It is a blocking function that can only be terminated via SIGINT
func (r *Runtime) RunPath(path string, args []string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("unable to stat %q: %v", path, err)
	}

	// Debug store should always be created prior to any prestart functions
	if err := r.initializeDebugStore(); err != nil {
		return fmt.Errorf("unable to create directory to store debug information: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	r.handlePreStart()

	go r.runEnvoy(path, args, cancel)

	r.waitForTerminationSignals(ctx)
	r.handleTermination()
	cancel() // this should never actually do anything but lets avoid any future context leaks

	// Block until the Envoy process and termination handler are finished cleaning up
	r.wg.Wait()
	return nil
}

// DebugStore returns the location at which the runtime instance persists debug data for this given instance
// Getters typically aren't idiomatic Go, however, this one is deliberately part of the runner interface
func (r *Runtime) DebugStore() string {
	return r.debugDir
}

func (r *Runtime) waitForTerminationSignals(ctx context.Context) {
	signal.Notify(r.signals, syscall.SIGINT)

	// Block until we receive SIGINT or are canceled because Envoy has died
	select {
	case <-ctx.Done():
		log.Infof("No Envoy processes remaining, terminating GetEnvoy process (PID=%d)", os.Getpid())
		return
	case <-r.signals:
		log.Infof("GetEnvoy process (PID=%d) received SIGINT", os.Getpid())
		return
	}
}

func (r *Runtime) runEnvoy(path string, args []string, cancel context.CancelFunc) {
	// #nosec -> users can run whatever binary they like!
	r.cmd = exec.Command(path, args...)
	r.cmd.SysProcAttr = sysProcAttr()
	r.cmd.Stdout = os.Stdout
	r.cmd.Stderr = os.Stderr

	r.wg.Add(1)
	defer r.wg.Done()
	defer cancel()

	if err := r.cmd.Run(); err != nil {
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
