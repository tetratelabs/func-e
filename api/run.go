// Copyright 2022 Tetrate
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

// Package api allows go projects to use func-e as a library.
package api

import (
	"context"
	"io"
	"os"
	"runtime"

	"github.com/tetratelabs/func-e/internal/cmd"
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

// ErrOut is where error messages are written. Defaults to os.Stderr
func ErrOut(err io.Writer) RunOption {
	return func(o *runOpts) {
		o.errOut = err
	}
}

// RunOption is configuration for Run.
type RunOption func(*runOpts)

type runOpts struct {
	homeDir          string
	envoyVersion     string
	envoyVersionsURL string
	out, errOut      io.Writer
}

// Run downloads Envoy and runs it as a process with the arguments
// passed to it. Use RunOption for configuration options.
func Run(ctx context.Context, args []string, options ...RunOption) error {
	ro := &runOpts{
		homeDir:          globals.DefaultHomeDir,
		envoyVersion:     "", // default to lookup
		envoyVersionsURL: globals.DefaultEnvoyVersionsURL,
		out:              os.Stdout,
	}
	for _, option := range options {
		option(ro)
	}

	o := globals.GlobalOpts{
		HomeDir:          ro.homeDir,
		EnvoyVersion:     version.PatchVersion(ro.envoyVersion),
		EnvoyVersionsURL: ro.envoyVersion,
		Out:              ro.out,
	}

	funcECmd := cmd.NewApp(&o)

	funcERunArgs := []string{"func-e", "--platform", runtime.GOOS + "/" + runtime.GOARCH, "run"}
	funcERunArgs = append(funcERunArgs, args...)

	errChan := make(chan error)
	go func() {
		errChan <- funcECmd.RunContext(ctx, funcERunArgs)
	}()

	// Wait for run to exit or an explicit stop.
	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		return err
	}
}
