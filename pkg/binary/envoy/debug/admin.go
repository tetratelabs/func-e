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

package debug

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/tetratelabs/multierror"

	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
)

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

// EnableEnvoyAdminDataCollection is a preset option that registers collection of Envoy Admin API information
func EnableEnvoyAdminDataCollection(r *envoy.Runtime) {
	e := envoyAdminDataCollection{r.GetAdminAddress, r.GetWorkingDir()}
	r.RegisterPreTermination(e.retrieveAdminAPIData)
}

type envoyAdminDataCollection struct {
	getAdminAddress func() (string, error)
	workingDir      string
}

func (e *envoyAdminDataCollection) retrieveAdminAPIData() error {
	adminAddress, err := e.getAdminAddress()
	if err != nil {
		return fmt.Errorf("unable to capture Envoy configuration and metrics: %w", err)
	}
	multiErr := &multierror.Error{}
	for path, file := range adminAPIPaths {
		resp, err := http.Get(fmt.Sprintf("http://%s/%v", adminAddress, path))
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			multiErr = multierror.Append(multiErr, fmt.Errorf("received %v from /%v ", resp.StatusCode, path))
			continue
		}
		// #nosec -> r.GetWorkingDir() is allowed to be anywhere
		f, err := os.OpenFile(filepath.Join(e.workingDir, file), os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
			continue
		}
		defer f.Close()         //nolint
		defer resp.Body.Close() //nolint
		if _, err := io.Copy(f, resp.Body); err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return multiErr.ErrorOrNil()
}
