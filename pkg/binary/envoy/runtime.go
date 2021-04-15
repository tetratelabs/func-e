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
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/tetratelabs/log"

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
		RootDir:         local,
		fetcher:         fetcher{local},
		TmplDir:         filepath.Join(local, "templates"),
		preStart:        make([]func(binary.Runner) error, 0),
		preTermination:  make([]func(binary.Runner) error, 0),
		postTermination: make([]func(binary.Runner) error, 0),
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
	binary.FetchRunner
	fetcher

	RootDir  string
	debugDir string
	TmplDir  string

	WorkingDir string
	IO         ioutil.StdStreams

	cmd                            *exec.Cmd
	adminAddress, adminAddressPath string
	fakeInterrupt                  context.CancelFunc

	preStart, preTermination, postTermination []func(binary.Runner) error

	isReady bool
}

// GetAdminAddress returns the current admin address in host:port format, or empty if not yet available.
// Exported for debug.EnableEnvoyAdminDataCollection, which is always on due to debug.EnableAll.
func (r *Runtime) GetAdminAddress() (string, error) {
	if r.adminAddress != "" { // We don't expect the admin address to change once written, so cache it.
		return r.adminAddress, nil
	}
	adminAddress, err := os.ReadFile(r.adminAddressPath) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("unable to read %s: %w", r.adminAddressPath, err)
	}
	if _, _, err := net.SplitHostPort(string(adminAddress)); err != nil {
		return "", fmt.Errorf("invalid admin address in %s: %w", r.adminAddressPath, err)
	}
	r.adminAddress = string(adminAddress)
	return r.adminAddress, nil
}

// Status indicates the state of the child process
func (r *Runtime) Status() int {
	switch {
	case r.cmd == nil, r.cmd.Process == nil:
		return binary.StatusStarting
	case r.cmd.ProcessState == nil: // the process started, but it hasn't completed, yet
		// TODO: envoyReady() can succeed when there's a port collision even when the process managed by this
		// runtime dies. We should consider checking for conflict on admin port_value when non-zero
		if status, err := r.envoyReady(); err == nil {
			return status
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

// envoyReady reads the HTTP status from the /ready endpoint and returns binary.StatusReady on 200 or
// binary.StatusInitializing on 503.
func (r *Runtime) envoyReady() (int, error) {
	adminAddress, err := r.GetAdminAddress()
	if err != nil { // don't yet have an admin address
		log.Debugf("%v", err) // don't fill logs
		return binary.StatusInitializing, nil
	}

	// Once we have seen its ready once stop spamming the ready endpoint.
	// If we expand the interface to support ready <-> not ready then
	// this approach will be wrong but as the states are monotonic this is good enough for now
	if r.isReady {
		return binary.StatusReady, nil
	}
	resp, err := http.Get(fmt.Sprintf("http://%s/ready", adminAddress))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close() //nolint
	switch resp.StatusCode {
	case http.StatusOK:
		r.isReady = true
		return binary.StatusReady, nil
	case http.StatusServiceUnavailable:
		return binary.StatusInitializing, nil
	default:
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

// FakeInterrupt is exposed for unit tests to pretend "getenvoy run" received an interrupt or a ctrl-c. End-to-end
// tests should kill the getenvoy process to achieve the same.
func (r *Runtime) FakeInterrupt() {
	fakeInterrupt := r.fakeInterrupt
	if fakeInterrupt != nil {
		fakeInterrupt()
	}
}
