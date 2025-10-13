// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package globals

import (
	"fmt"
	"path/filepath"
	"strconv"
	"time"
)

// EnvoyVersionsDir returns the directory containing Envoy binaries.
// Legacy: "$dataHome/versions"
// Default: "$dataHome/envoy-versions"
func (o *GlobalOpts) EnvoyVersionsDir() string {
	if o.HomeDir != "" {
		return filepath.Join(o.DataHome, "versions")
	}
	return filepath.Join(o.DataHome, "envoy-versions")
}

// EnvoyVersionFile returns the path to the selected version file.
// Legacy: "$homeDir/version"
// Default: "$configHome/envoy-version"
func (o *GlobalOpts) EnvoyVersionFile() string {
	if o.HomeDir != "" {
		return filepath.Join(o.DataHome, "version")
	}
	return filepath.Join(o.ConfigHome, "envoy-version")
}

// EnvoyVersionFileSource returns the display string for the version file source.
// This is used in help text and command output.
// Legacy: "$FUNC_E_HOME/version"
// Default: "$FUNC_E_CONFIG_HOME/envoy-version"
func (o *GlobalOpts) EnvoyVersionFileSource() string {
	if o.HomeDir != "" {
		return "$FUNC_E_HOME/version"
	}
	return "$FUNC_E_CONFIG_HOME/envoy-version"
}

// EnvoyRunDir returns the directory for a specific run (logs, config_dump.json, etc.).
// Legacy: "$stateHome/runs/{runID}"
// Default: "$stateHome/envoy-runs/{runID}"
func (o *GlobalOpts) EnvoyRunDir(runID string) string {
	if o.HomeDir != "" {
		return filepath.Join(o.StateHome, "runs", runID)
	}
	return filepath.Join(o.StateHome, "envoy-runs", runID)
}

// EnvoyRuntimeDir returns the directory for temporary files of a specific run.
// Legacy: "$runtimeDir/runs/{runID}"
// Default: "$runtimeDir/{runID}"
func (o *GlobalOpts) EnvoyRuntimeDir(runID string) string {
	if o.HomeDir != "" {
		return filepath.Join(o.RuntimeDir, "runs", runID)
	}
	return filepath.Join(o.RuntimeDir, runID)
}

// GenerateRunID creates a unique run identifier.
// Legacy: epoch nanoseconds "1760229853109700000"
// Default: "20251012_143053_700" (YYYYMMDD_HHMMSS_UUU)
func (o *GlobalOpts) GenerateRunID(t time.Time) string {
	if o.HomeDir != "" {
		return strconv.FormatInt(t.UnixNano(), 10)
	}
	// YYYYMMDD_HHMMSS_UUU format
	// Last 3 digits of microseconds to allow concurrent runs
	micro := t.Nanosecond() / 1000 % 1000
	return fmt.Sprintf("%s_%03d", t.Format("20060102_150405"), micro)
}
