// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package globals

import (
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/tetratelabs/func-e/internal/version"
)

// RunOpts support invocations of "func-e run"
type RunOpts struct {
	// EnvoyPath is the exec.Cmd path to "envoy". Defaults to "$HomeDir/versions/$version/bin/envoy"
	EnvoyPath string
	// EnvoyOut is where to write Envoy's stdout.
	EnvoyOut io.Writer
	// EnvoyErr is where to write Envoy's stdout.
	EnvoyErr io.Writer
	// RunDir is the location any generated files are written.
	// This is not Envoy's working directory, which remains the same as the $PWD of func-e.
	// Defaults to "$HomeDir/runs/$epochtime"
	RunDir string
}

// GlobalOpts represents options that affect more than one func-e commands.
//
// Fields representing non-hidden flags have values set according to the following rules:
//  1. value that precedes flag parsing, used in tests
//  2. to a value of the command line argument, e.g. `--home-dir`
//  3. optional mapping to an environment variable, e.g. `FUNC_E_HOME` (not all flags are mapped to ENV)
//  4. otherwise, to the default value, e.g. DefaultHomeDir
type GlobalOpts struct {
	// RunOpts are inlined to allow tests to override parameters without changing ENV variables or flags
	RunOpts
	// Version is the version of the CLI, used in help statements and HTTP requests via "User-Agent".
	// Override this via "-X main.version=XXX"
	Version string
	// EnvoyVersionsURL is the path to the envoy-versions.json. Defaults to DefaultEnvoyVersionsURL
	EnvoyVersionsURL string
	// EnvoyVersion is the default version of Envoy to run. Defaults to the contents of "$HomeDir/versions/version".
	// When that file is missing, it is generated from ".latestVersion" from the EnvoyVersionsURL. Its
	// value can be in full version major.minor.patch format, e.g. 1.18.1 or without patch component,
	// major.minor, e.g. 1.18.
	EnvoyVersion version.PatchVersion
	// HomeDir is an absolute path which most importantly contains "versions" installed from EnvoyVersionsURL. Defaults to DefaultHomeDir
	HomeDir string
	// Quiet means don't Logf to Out
	Quiet bool
	// Out is where status messages are written. Defaults to os.Stdout
	Out io.Writer
	// The platform to target for the Envoy install.
	Platform version.Platform
	// GetEnvoyVersions returns Envoy release versions from EnvoyVersionsURL.
	GetEnvoyVersions version.GetReleaseVersions
}

// Logf is used for shared functions that log conditionally on GlobalOpts.Quiet
func (o *GlobalOpts) Logf(format string, a ...interface{}) {
	if o.Quiet { // TODO: we may want to do scoped logging via a Context property, if this becomes common.
		return
	}
	// Always add a newline to ensure consistent formatting
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(o.Out, format, a...) //nolint:errcheck
}

const (
	// DefaultEnvoyVersionsURL is the default value for GlobalOpts.EnvoyVersionsURL
	DefaultEnvoyVersionsURL = "https://archive.tetratelabs.io/envoy/envoy-versions.json"
	// DefaultEnvoyVersionsSchemaURL is the JSON schema used to validate GlobalOpts.EnvoyVersionsURL
	DefaultEnvoyVersionsSchemaURL = "https://archive.tetratelabs.io/release-versions-schema.json"
	// DefaultPlatform is the current platform of the host machine
	DefaultPlatform = version.Platform(runtime.GOOS + "/" + runtime.GOARCH)
)

// DefaultHomeDir is the default value for GlobalOpts.HomeDir
var DefaultHomeDir = "${HOME}/.func-e"
