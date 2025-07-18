// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package shutdown

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

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

// enableAdminDataCollection is a preset option that registers collection of Envoy Admin API information
func enableAdminDataCollection(r *envoy.Runtime) error {
	e := adminDataCollection{r.GetAdminAddress, r.GetRunDir()}
	r.RegisterShutdownHook(e.retrieveAdminAPIData)
	return nil
}

type adminDataCollection struct {
	getAdminAddress func() (string, error)
	workingDir      string
}

func (e *adminDataCollection) retrieveAdminAPIData(ctx context.Context) error {
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
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			return copyURLToFile(ctx, url, file)
		})
	}
	return g.Wait() // first error
}

func copyURLToFile(ctx context.Context, url, fullPath string) error {
	// #nosec -> e.workingDir is allowed to be anywhere
	f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("could not open %q: %w", fullPath, err)
	}
	defer f.Close() //nolint:errcheck

	// #nosec -> adminAddress is written by Envoy and the paths are hard-coded
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("could not create request %v: %w", url, err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not read %v: %w", url, err)
	}
	defer res.Body.Close() //nolint:errcheck

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("received %v from %v", res.StatusCode, url)
	}
	if _, err := io.Copy(f, res.Body); err != nil {
		return fmt.Errorf("could not write response body of %v: %w", url, err)
	}
	return nil
}
