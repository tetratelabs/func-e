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
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/tetratelabs/getenvoy/pkg/binary"
	"github.com/tetratelabs/getenvoy/pkg/common"

	ioutil "github.com/tetratelabs/getenvoy/pkg/util/io"
)

// RuntimeOption represents a configuration option to NewRuntime.
type RuntimeOption func(*Runtime)

// And returns a combined list of configuration options.
func (o RuntimeOption) And(opts ...RuntimeOption) RuntimeOptions {
	return o.AndAll(opts)
}

// AndAll returns a combined list of configuration options.
func (o RuntimeOption) AndAll(opts RuntimeOptions) RuntimeOptions {
	return RuntimeOptions{o}.And(opts...)
}

// RuntimeOptions represents a list of configuration options to NewRuntime.
type RuntimeOptions []RuntimeOption

// And returns a combined list of configuration options.
func (o RuntimeOptions) And(opts ...RuntimeOption) RuntimeOptions {
	return o.AndAll(opts)
}

// AndAll returns a combined list of configuration options.
func (o RuntimeOptions) AndAll(opts RuntimeOptions) RuntimeOptions {
	return append(o, opts...)
}

// NewRuntime creates a new Runtime with the local file storage set to the home directory
func NewRuntime(options ...RuntimeOption) (binary.FetchRunner, error) {
	local := common.HomeDir
	runtime := &Runtime{
		Config:         NewConfig(),
		fetcher:        fetcher{local},
		TmplDir:        filepath.Join(local, "templates"),
		wg:             &sync.WaitGroup{},
		signals:        make(chan os.Signal),
		preStart:       make([]func(binary.Runner) error, 0),
		preTermination: make([]func(binary.Runner) error, 0),
	}

	if debugErr := runtime.initializeDebugStore(); debugErr != nil {
		return nil, fmt.Errorf("unable to create directory to store debug information: %v", debugErr)
	}

	for _, option := range options {
		option(runtime)
	}
	return runtime, nil
}

type fetcher struct {
	store string
}

// Runtime manages an Envoy lifecycle including fetching (if necessary) and running
type Runtime struct {
	fetcher

	debugDir string
	TmplDir  string
	Config   *Config

	WorkingDir string
	IO         ioutil.StdStreams

	cmd *exec.Cmd
	ctx context.Context
	wg  *sync.WaitGroup

	signals chan os.Signal

	preStart       []func(binary.Runner) error
	preTermination []func(binary.Runner) error

	isReady bool
}

// Status indicates the state of the child process
func (r *Runtime) Status() int {
	switch {
	case r.cmd == nil, r.cmd.Process == nil:
		return binary.StatusStarting
	case r.cmd.ProcessState == nil:
		if r.envoyReady() {
			return binary.StatusReady
		}
		return binary.StatusStarted
	default:
		return binary.StatusTerminated
	}
}

// GetPid returns the pid of the child process
func (r *Runtime) GetPid() (int, error) {
	if r.cmd == nil || r.cmd.Process == nil {
		return 0, fmt.Errorf("envoy process not yet started")
	}
	return r.cmd.Process.Pid, nil
}

func (r *Runtime) envoyReady() bool {
	// Once we have seen its ready once stop spamming the ready endpoint.
	// If we expand the interface to support ready <-> not ready then
	// this approach will be wrong but as the states are monotonic this is good enough for now
	if r.isReady {
		return true
	}
	resp, err := http.Get(fmt.Sprintf("http://%s/ready", r.Config.GetAdminAddress()))
	if err != nil {
		return false
	}
	defer resp.Body.Close() //nolint
	if resp.StatusCode == http.StatusOK {
		r.isReady = true
		return r.isReady
	}
	return false
}

// Wait blocks until the child process reaches the state passed
// Note: It does not guarantee that it is in the specified state just that it has reached it
func (r *Runtime) Wait(state int) {
	for r.Status() < state {
		// This is a call to a function to allow the goroutine to be preempted for garbage collection
		// The sleep duration is somewhat arbitrary
		func() { time.Sleep(time.Millisecond * 100) }()
	}
}

// WaitWithContext blocks until the child process reaches the state passed or the context is canceled
// Note: It does not guarantee that it is in the specified state just that it has reached it
func (r *Runtime) WaitWithContext(ctx context.Context, state int) {
	done := make(chan struct{})
	go func() {
		r.Wait(state)
		close(done)
	}()
	select {
	case <-done:
		return
	case <-ctx.Done():
		return

	}
}

// SendSignal sends a signal to the parent process
func (r *Runtime) SendSignal(s os.Signal) {
	r.signals <- s
}
