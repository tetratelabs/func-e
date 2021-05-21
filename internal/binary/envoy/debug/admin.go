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

	"github.com/tetratelabs/getenvoy/internal/binary/envoy"
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

// enableEnvoyAdminDataCollection is a preset option that registers collection of Envoy Admin API information
func enableEnvoyAdminDataCollection(r *envoy.Runtime) error {
	e := envoyAdminDataCollection{r.GetAdminAddress, r.GetWorkingDir()}
	r.RegisterPreTermination(e.retrieveAdminAPIData)
	return nil
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
	for p, f := range adminAPIPaths {
		url := fmt.Sprintf("http://%s/%v", adminAddress, p)
		file := filepath.Join(e.workingDir, f)
		if e := copyURLToFile(url, file); e != nil {
			return e
		}
	}
	return nil
}

func copyURLToFile(url, fullPath string) error {
	// #nosec -> e.workingDir is allowed to be anywhere
	f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("could not open %q: %w", fullPath, err)
	}
	defer f.Close() //nolint

	// #nosec -> adminAddress is written by Envoy and the paths are hard-coded
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("could not read %v: %w", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received %v from %v", resp.StatusCode, url)
	}
	defer resp.Body.Close() //nolint

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("could not write response body of %v: %w", url, err)
	}
	return nil
}
