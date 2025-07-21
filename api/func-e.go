// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

// Package api allows Go projects to use func-e as a library, decoupled from how
// the func-e binary reads environment variables or CLI args.
package api

import (
	"context"
	"io"
	"os"

	"github.com/tetratelabs/func-e/internal/api"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/version"
)

// HomeDir is an absolute path which most importantly contains "versions"
// installed from EnvoyVersionsURL. Defaults to "${HOME}/.func-e"
func HomeDir(homeDir string) RunOption {
	return func(o *runOpts) {
		o.homeDir = homeDir
	}
}

// EnvoyVersionsURL is the path to the envoy-versions.json.
// Defaults to "https://archive.tetratelabs.io/envoy/envoy-versions.json"
func EnvoyVersionsURL(envoyVersionsURL string) RunOption {
	return func(o *runOpts) {
		o.envoyVersionsURL = envoyVersionsURL
	}
}

// EnvoyVersion overrides the version of Envoy to run. Defaults to the
// contents of "$HomeDir/versions/version".
//
// When that file is missing, it is generated from ".latestVersion" from the
// EnvoyVersionsURL. Its value can be in full version major.minor.patch format,
// e.g. 1.18.1 or without patch component, major.minor, e.g. 1.18.
func EnvoyVersion(envoyVersion string) RunOption {
	return func(o *runOpts) {
		o.envoyVersion = envoyVersion
	}
}

// Out is where status messages are written. Defaults to os.Stdout
func Out(out io.Writer) RunOption {
	return func(o *runOpts) {
		o.out = out
	}
}

// EnvoyOut sets the writer for Envoy stdout
func EnvoyOut(w io.Writer) RunOption {
	return func(o *runOpts) {
		o.envoyOut = w
	}
}

// EnvoyErr sets the writer for Envoy stderr
func EnvoyErr(w io.Writer) RunOption {
	return func(o *runOpts) {
		o.envoyErr = w
	}
}

// envoyPath overrides the path to the Envoy binary. Used for testing with a fake binary.
func envoyPath(envoyPath string) RunOption {
	return func(o *runOpts) {
		o.envoyPath = envoyPath
	}
}

// RunOption is configuration for Run.
//
// Note: None of these default to values read from OS environment variables.
// If you wish to introduce such behavior, populate them in calling code.
type RunOption func(*runOpts)

type runOpts struct {
	homeDir          string
	envoyVersion     string
	envoyVersionsURL string
	out              io.Writer
	envoyOut         io.Writer
	envoyErr         io.Writer
	envoyPath        string // optional: path to the Envoy binary (for tests)
}

// Run downloads Envoy and runs it as a process with the arguments
// passed to it. Use RunOption for configuration options.
//
// On success, this blocks and returns nil when either `ctx` is done, or the
// process exits with status zero.
func Run(ctx context.Context, args []string, options ...RunOption) error {
	// TODO: we need a real API and it being an interface in this package, initialized in the root
	// directory like wazero does. That this stitches the impl makes it not an API package and causes
	// package import complexity we need to remove.
	o, err := initOpts(ctx, options...)
	if err != nil {
		return err
	}
	return api.Run(ctx, o, args)
}

// initOpts is a placeholder to adapt E2E tests until we have a real API for func-e.
func initOpts(ctx context.Context, options ...RunOption) (*globals.GlobalOpts, error) {
	ro := &runOpts{
		out:      os.Stdout,
		envoyOut: os.Stdout,
		envoyErr: os.Stderr,
	}
	for _, option := range options {
		option(ro)
	}

	o := &globals.GlobalOpts{
		EnvoyVersion: version.PatchVersion(ro.envoyVersion),
		Out:          ro.out,
		RunOpts: globals.RunOpts{
			EnvoyPath: ro.envoyPath,
			EnvoyOut:  ro.envoyOut,
			EnvoyErr:  ro.envoyErr,
		},
	}
	if err := api.InitializeGlobalOpts(o, ro.envoyVersionsURL, ro.homeDir, ""); err != nil {
		return nil, err
	}

	return o, api.EnsureEnvoyVersion(ctx, o)
}
