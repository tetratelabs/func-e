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
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/tetratelabs/getenvoy/internal/globals"
)

const (
	// Don't wait forever. This has hung on macOS before
	shutdownTimeout = 5 * time.Second
	// Match envoy's log format field
	dateFormat = "[2006-01-02 15:04:05.999]"
)

// NewRuntime creates a new Runtime that runs envoy in globals.RunOpts RunDir
// opts allows a user running envoy to control the working directory by ID or path, allowing explicit cleanup.
func NewRuntime(opts *globals.RunOpts) *Runtime {
	return &Runtime{opts: opts}
}

// Runtime manages an Envoy lifecycle
type Runtime struct {
	opts *globals.RunOpts

	cmd *exec.Cmd
	Out io.Writer
	Err io.Writer

	adminAddress, adminAddressPath string

	// FakeInterrupt is exposed for unit tests to pretend "getenvoy run" received an interrupt or a ctrl-c.
	// End-to-end tests should kill the getenvoy process to achieve the same.
	FakeInterrupt context.CancelFunc

	shutdownHooks []func(context.Context) error
}

// GetRunDir returns the run-specific directory files can be written to.
func (r *Runtime) GetRunDir() string {
	return r.opts.RunDir
}

// ensureAdminAddressPath sets the "--admin-address-path" flag so that it can be used in /ready checks. If a value
// already exists, it will be returned. Otherwise, the flag will be set to the file "admin-address.txt" in the
// run directory. We don't use the working directory as sometimes that is a source directory.
//
// Notably, this allows ephemeral admin ports via bootstrap configuration admin/port_value=0 (minimum Envoy 1.12 for macOS support)
func (r *Runtime) ensureAdminAddressPath() error {
	args := r.cmd.Args
	flag := `--admin-address-path`
	for i, a := range args {
		if a == flag {
			if i+1 == len(args) || args[i+1] == "" {
				return fmt.Errorf(`missing value to argument %q`, flag)
			}
			r.adminAddressPath = args[i+1]
			return nil
		}
	}
	// Envoy's run directory is mutable, so it is fine to write the admin address there.
	r.adminAddressPath = filepath.Join(r.opts.RunDir, "admin-address.txt")
	r.cmd.Args = append(r.cmd.Args, flag, r.adminAddressPath)
	return nil
}

// GetAdminAddress returns the current admin address in host:port format, or empty if not yet available.
// Exported for shutdown.enableEnvoyAdminDataCollection, which is always on due to shutdown.EnableHooks.
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
