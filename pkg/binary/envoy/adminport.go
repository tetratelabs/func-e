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
	"io/ioutil"
	"net"
	"path/filepath"

	"github.com/tetratelabs/log"

	"github.com/tetratelabs/getenvoy/pkg/binary"
)

// EnableAdminAddressDetection sets the "--admin-address-path" flag and reads it back before attempting ready checks.
//
// Notably, this allows ephemeral admin ports via bootstrap configuration admin/port_value=0
func EnableAdminAddressDetection(r *Runtime) {
	r.RegisterPreStart(appendArg)
}

func appendArg(r binary.Runner) error {
	e, ok := r.(*Runtime)
	if !ok {
		return fmt.Errorf("unable to use admin address detection")
	}
	r.AppendArgs([]string{"--admin-address-path", filepath.Join(r.DebugStore(), "admin-address.txt")})

	// NOTE: this is in the envoy package to allow sneaky access to overwrite
	e.getAdminAddress = readAdminAddressFile
	return nil
}

func readAdminAddressFile(r *Runtime) string {
	adminAddressFile := filepath.Join(r.DebugStore(), "admin-address.txt")
	adminAddress, err := ioutil.ReadFile(adminAddressFile) //nolint:gosec
	if err != nil {
		log.Debugf("unable to read %s: %v", adminAddressFile, err)
		return ""
	}
	if _, _, err := net.SplitHostPort(string(adminAddress)); err != nil {
		log.Debugf("invalid admin address in %s: %v", adminAddressFile, err)
		return ""
	}
	return string(adminAddress)
}
