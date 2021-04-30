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
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/tetratelabs/getenvoy/pkg/globals"
	ioutil "github.com/tetratelabs/getenvoy/pkg/util/io"
)

// NewRuntime creates a new Runtime that runs envoy in globals.RunOpts WorkingDir
// opts allows a user running envoy to control the working directory by ID or path, allowing explicit cleanup.
func NewRuntime(opts *globals.RunOpts) *Runtime {
	return &Runtime{opts: opts}
}

// Runtime manages an Envoy lifecycle including fetching (if necessary) and running
type Runtime struct {
	opts *globals.RunOpts

	cmd *exec.Cmd
	IO  ioutil.StdStreams

	adminAddress, adminAddressPath string

	// FakeInterrupt is exposed for unit tests to pretend "getenvoy run" received an interrupt or a ctrl-c.
	// End-to-end tests should kill the getenvoy process to achieve the same.
	FakeInterrupt context.CancelFunc

	preStart, preTermination, postTermination []func() error
}

// GetWorkingDir returns the run-specific directory files can be written to.
func (r *Runtime) GetWorkingDir() string {
	return r.opts.WorkingDir
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

// GetPid returns the pid of the child process
func (r *Runtime) GetPid() (int, error) {
	if r.cmd == nil || r.cmd.Process == nil {
		return 0, errors.New("envoy process not yet started")
	}
	return r.cmd.Process.Pid, nil
}
