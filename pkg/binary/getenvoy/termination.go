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

package getenvoy

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"syscall"

	"github.com/hashicorp/go-multierror"
	"github.com/tetratelabs/log"
)

func (r *Runtime) handleTermination() {
	if r.cmd.ProcessState != nil {
		log.Infof("Envoy process (PID=%d) terminated prematurely", r.cmd.Process.Pid)
		return
	}

	// Execute all registered preTermination functions
	for _, f := range r.preTermination {
		if err := f(); err != nil {
			log.Error(err.Error())
		}
	}

	// Forward on the SIGINT to Envoy
	log.Infof("Sending Envoy process (PID=%d) SIGINT", r.cmd.Process.Pid)
	_ = r.cmd.Process.Signal(syscall.SIGKILL)

	// TODO: tar it all up! (Liam)
}

func (r *Runtime) registerPreTermination(f ...preTerminationFunc) {
	r.preTermination = append(r.preTermination, f...)
}

var adminAPIPaths = map[string]string{
	"certs":             "certs.json",
	"clusters":          "clusters.txt",
	"config_dump":       "config_dump.json",
	"contention":        "contention.txt",
	"listeners":         "listeners.txt",
	"memory":            "memory.json",
	"server_info":       "server_info.json",
	"stats?format=json": "stats.json",
	"runtime":           "runtime.json",
}

// EnvoyAdminDataCollection registers collection of Envoy Admin API information
// TODO: Test this (Liam)
func (r *Runtime) EnvoyAdminDataCollection(enable bool) {
	if enable {
		r.registerPreTermination(func() error {
			var multiErr *multierror.Error
			for path, file := range adminAPIPaths {
				resp, err := http.Get(fmt.Sprintf("http://0.0.0.0:15001/%v", path))
				if err != nil {
					multiErr = multierror.Append(multiErr, err)
				}
				f, err := os.OpenFile(filepath.Join(r.debugDir, file), os.O_CREATE|os.O_WRONLY, 0600)
				if err != nil {
					multiErr = multierror.Append(multiErr, err)
				}
				defer func() { _ = f.Close() }()
				defer func() { _ = resp.Body.Close() }()
				if _, err := io.Copy(f, resp.Body); err != nil {
					multiErr = multierror.Append(multiErr, err)
				}
			}
			return multiErr.ErrorOrNil()
		})
	}
}
