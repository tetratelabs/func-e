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

package shutdown

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"

	"github.com/tetratelabs/func-e/internal/envoy"
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
	e := envoyAdminDataCollection{r.GetAdminAddress, r.GetRunDir()}
	r.RegisterShutdownHook(e.retrieveAdminAPIData)
	return nil
}

type envoyAdminDataCollection struct {
	getAdminAddress func() (string, error)
	workingDir      string
}

func (e *envoyAdminDataCollection) retrieveAdminAPIData(ctx context.Context) error {
	adminAddress, err := e.getAdminAddress()
	if err != nil {
		return fmt.Errorf("unable to capture Envoy configuration and metrics: %w", err)
	}

	// Save each admin API path to a file in parallel returning on first error
	// Execute all admin fetches in parallel
	g, ctx := errgroup.WithContext(ctx)
	for p, f := range adminAPIPaths {
		url := fmt.Sprintf("http://%s/%v", adminAddress, p)
		file := filepath.Join(e.workingDir, f)

		g.Go(func() error {
			return copyURLToFile(ctx, url, file)
		})
	}
	return g.Wait() // first error
}

func copyURLToFile(ctx context.Context, url, fullPath string) error {
	// #nosec -> e.workingDir is allowed to be anywhere
	f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("could not open %q: %w", fullPath, err)
	}
	defer f.Close() //nolint

	// #nosec -> adminAddress is written by Envoy and the paths are hard-coded
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("could not create request %v: %w", url, err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not read %v: %w", url, err)
	}
	defer res.Body.Close() //nolint

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("received %v from %v", res.StatusCode, url)
	}
	if _, err := io.Copy(f, res.Body); err != nil {
		return fmt.Errorf("could not write response body of %v: %w", url, err)
	}
	return nil
}
