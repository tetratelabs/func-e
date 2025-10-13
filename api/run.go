// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

// Package api allows Go projects to use func-e as a library, decoupled from how
// the func-e binary reads environment variables or CLI args.
package api

import (
	"context"
	"io"

	"github.com/tetratelabs/func-e/internal/api"
)

// Deprecated: Use ConfigHome, DataHome, StateHome or RuntimeDir instead.
// This function will be removed in a future version.
func HomeDir(homeDir string) RunOption {
	return func(o *api.RunOpts) {
		o.ConfigHome = homeDir
		o.DataHome = homeDir
		o.StateHome = homeDir
		o.RuntimeDir = homeDir
	}
}

// ConfigHome is the directory containing configuration files.
// Defaults to "~/.config/func-e"
//
// Files stored here:
// - envoy-version (selected version preference)
func ConfigHome(configHome string) RunOption {
	return func(o *api.RunOpts) {
		o.ConfigHome = configHome
	}
}

// DataHome is the directory containing downloaded Envoy binaries.
// Defaults to "~/.local/share/func-e"
//
// Files stored here:
// - envoy-versions/{version}/bin/envoy (downloaded Envoy binaries)
func DataHome(dataHome string) RunOption {
	return func(o *api.RunOpts) {
		o.DataHome = dataHome
	}
}

// StateHome is the directory containing persistent state like run logs.
// Defaults to "~/.local/state/func-e"
//
// Files stored here:
// - envoy-runs/{runID}/stdout.log,stderr.log (per-run logs)
// - envoy-runs/{runID}/config_dump.json (Envoy configuration snapshot)
func StateHome(stateHome string) RunOption {
	return func(o *api.RunOpts) {
		o.StateHome = stateHome
	}
}

// RuntimeDir is the directory containing ephemeral runtime files.
// Defaults to "/tmp/func-e-${UID}"
//
// Files stored here:
// - {runID}/admin-address.txt (Envoy admin API endpoint)
//
// Note: Runtime files are ephemeral and may be cleaned up on system restart.
func RuntimeDir(runtimeDir string) RunOption {
	return func(o *api.RunOpts) {
		o.RuntimeDir = runtimeDir
	}
}

// RunID sets a custom run identifier used in StateDir and RuntimeDir paths.
// By default, a timestamp-based runID is auto-generated (e.g., "20250115_123456_789").
//
// Use this to:
// - Create predictable directories for Docker/K8s (e.g., RunID("0"))
// - Implement custom naming schemes
//
// Validation: runID cannot contain path separators (/ or \)
func RunID(runID string) RunOption {
	return func(o *api.RunOpts) {
		o.RunID = runID
	}
}

// EnvoyVersionsURL is the path to the envoy-versions.json.
// Defaults to "https://archive.tetratelabs.io/envoy/envoy-versions.json"
func EnvoyVersionsURL(envoyVersionsURL string) RunOption {
	return func(o *api.RunOpts) {
		o.EnvoyVersionsURL = envoyVersionsURL
	}
}

// EnvoyVersion overrides the version of Envoy to run. Defaults to the
// contents of "$ConfigHome/envoy-version".
//
// When that file is missing, it is generated from ".latestVersion" from the
// EnvoyVersionsURL. Its value can be in full version major.minor.patch format,
// e.g. 1.18.1 or without patch component, major.minor, e.g. 1.18.
func EnvoyVersion(envoyVersion string) RunOption {
	return func(o *api.RunOpts) {
		o.EnvoyVersion = envoyVersion
	}
}

// Out is where status messages are written. Defaults to os.Stdout
func Out(out io.Writer) RunOption {
	return func(o *api.RunOpts) {
		o.Out = out
	}
}

// EnvoyOut sets the writer for Envoy stdout
func EnvoyOut(w io.Writer) RunOption {
	return func(o *api.RunOpts) {
		o.EnvoyOut = w
	}
}

// EnvoyErr sets the writer for Envoy stderr
func EnvoyErr(w io.Writer) RunOption {
	return func(o *api.RunOpts) {
		o.EnvoyErr = w
	}
}

// RunOption is a configuration for RunFunc.
//
// Note: None of these default to values read from OS environment variables.
// If you wish to introduce such behavior, populate them in calling code.
type RunOption func(*api.RunOpts)

// RunFunc downloads Envoy and runs it as a process with the arguments
// passed to it. Use api.RunOption for configuration options.
//
// On success, this blocks and returns nil when either `ctx` is done, or the
// process exits with status zero.
//
// The default implementation of RunFunc is func_e.Run.
type RunFunc func(ctx context.Context, args []string, options ...RunOption) error
