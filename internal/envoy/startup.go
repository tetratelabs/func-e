// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// safeStartupHook wraps a StartupHook to provide panic recovery and timeout handling.
type safeStartupHook struct {
	delegate StartupHook
	logf     LogFunc
	timeout  time.Duration
}

// Ensure safeStartupHook implements StartupHook interface
var _ StartupHook = (*safeStartupHook)(nil).Hook

// Hook implements the StartupHook interface with panic recovery and timeout.
func (s *safeStartupHook) Hook(ctx context.Context, runDir, adminAddress string) error {
	defer func() {
		if p := recover(); p != nil {
			s.logf("startup hook panicked: %v", p)
		}
	}()

	if s.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.timeout)
		defer cancel()
	}

	if err := s.delegate(ctx, runDir, adminAddress); err != nil {
		s.logf(err.Error())
	}
	return nil
}

// collectConfigDump fetches config_dump from the Envoy admin API.
// It uses the ?include_eds parameter to include endpoint discovery service (EDS)
// data, which covers nearly all xDS configurations including:
// - Listeners (LDS)
// - Routes (RDS)
// - Clusters (CDS)
// - Endpoints (EDS)
// - Secrets (SDS)
// This provides a comprehensive snapshot of Envoy's dynamic configuration state.
func collectConfigDump(ctx context.Context, client *http.Client, runDir, adminAddress string) error {
	url := fmt.Sprintf("http://%s/config_dump?include_eds", adminAddress)
	file := filepath.Join(runDir, "config_dump.json")
	return copyURLToFile(ctx, client, url, file)
}

func copyURLToFile(ctx context.Context, client *http.Client, url, fullPath string) error {
	// #nosec -> runDir is allowed to be anywhere
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
	res, err := client.Do(req)
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
