// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

// Package api allows Go projects to use func-e as a library, decoupled from how
// the func-e binary reads environment variables or CLI args.
package api

import (
	"context"
	"io"

	"github.com/tetratelabs/func-e/internal/opts"
)

// HomeDir is an absolute path which most importantly contains "versions"
// installed from EnvoyVersionsURL. Defaults to "${HOME}/.func-e"
func HomeDir(homeDir string) RunOption {
	return func(o *opts.RunOpts) {
		o.HomeDir = homeDir
	}
}

// EnvoyVersionsURL is the path to the envoy-versions.json.
// Defaults to "https://archive.tetratelabs.io/envoy/envoy-versions.json"
func EnvoyVersionsURL(envoyVersionsURL string) RunOption {
	return func(o *opts.RunOpts) {
		o.EnvoyVersionsURL = envoyVersionsURL
	}
}

// EnvoyVersion overrides the version of Envoy to run. Defaults to the
// contents of "$HomeDir/versions/version".
//
// When that file is missing, it is generated from ".latestVersion" from the
// EnvoyVersionsURL. Its value can be in full version major.minor.patch format,
// e.g. 1.18.1 or without patch component, major.minor, e.g. 1.18.
func EnvoyVersion(envoyVersion string) RunOption {
	return func(o *opts.RunOpts) {
		o.EnvoyVersion = envoyVersion
	}
}

// Out is where status messages are written. Defaults to os.Stdout
func Out(out io.Writer) RunOption {
	return func(o *opts.RunOpts) {
		o.Out = out
	}
}

// EnvoyOut sets the writer for Envoy stdout
func EnvoyOut(w io.Writer) RunOption {
	return func(o *opts.RunOpts) {
		o.EnvoyOut = w
	}
}

// EnvoyErr sets the writer for Envoy stderr
func EnvoyErr(w io.Writer) RunOption {
	return func(o *opts.RunOpts) {
		o.EnvoyErr = w
	}
}

// RunOption is a configuration for RunFunc.
//
// Note: None of these default to values read from OS environment variables.
// If you wish to introduce such behavior, populate them in calling code.
type RunOption func(*opts.RunOpts)

// RunFunc downloads Envoy and runs it as a process with the arguments
// passed to it. Use api.RunOption for configuration options.
//
// On success, this blocks and returns nil when either `ctx` is done, or the
// process exits with status zero.
type RunFunc func(ctx context.Context, args []string, options ...RunOption) error
