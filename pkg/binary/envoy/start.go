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
	"fmt"
	"path/filepath"
)

func (r *Runtime) handlePreStart() error {
	// Execute all registered preStart functions
	for _, f := range r.preStart {
		if err := f(); err != nil {
			return err
		}
	}
	// Add admin path after we know the debug store location and other functions apply.
	// NOTE: getenvoy isn't supported as a library, so we don't really need to worry about preStart hooks conflicting.
	// If we are able to know the debug path sooner, we could refactor this later to be more direct.
	return r.ensureAdminAddressPath()
}

// RegisterPreStart registers the passed functions to be run before Envoy has started
func (r *Runtime) RegisterPreStart(f ...func() error) {
	r.preStart = append(r.preStart, f...)
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
	r.cmd.Args = append(r.cmd.Args, flag, "admin-address.txt")
	r.adminAddressPath = filepath.Join(r.opts.WorkingDir, "admin-address.txt")
	return nil
}
