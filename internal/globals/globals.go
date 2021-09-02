// Copyright 2021 Tetrate
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

package globals

import (
	"io"
	"regexp"
	"runtime"

	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/version"
)

// RunOpts support invocations of "func-e run"
type RunOpts struct {
	// EnvoyPath is the exec.Cmd path to "envoy". Defaults to "$HomeDir/versions/$version/bin/envoy"
	EnvoyPath string
	// RunDir is the location any generated files are written.
	// This is not Envoy's working directory, which remains the same as the $PWD of func-e.
	//
	// Upon shutdown, this directory is archived as "../$(basename $RunDir).tar.gz"
	// Defaults to "$HomeDir/runs/$epochtime"
	RunDir string
	// DontArchiveRunDir is used in testing and prevents archiving the RunDir
	DontArchiveRunDir bool
}

// GlobalOpts represents options that affect more than one func-e commands.
//
// Fields representing non-hidden flags have values set according to the following rules:
//  1) value that precedes flag parsing, used in tests
//  2) to a value of the command line argument, e.g. `--home-dir`
//  3) optional mapping to an environment variable, e.g. `FUNC_E_HOME` (not all flags are mapped to ENV)
//  4) otherwise, to the default value, e.g. DefaultHomeDir
type GlobalOpts struct {
	// RunOpts are inlined to allow tests to override parameters without changing ENV variables or flags
	RunOpts
	// Version is the version of the CLI, used in help statements and HTTP requests via "User-Agent".
	// Override this via "-X main.version=XXX"
	Version version.Version
	// EnvoyVersionsURL is the path to the envoy-versions.json. Defaults to DefaultEnvoyVersionsURL
	EnvoyVersionsURL string
	// EnvoyVersion is the default version of Envoy to run. Defaults to the contents of "$HomeDir/versions/version".
	// When that file is missing, it is generated from ".latestVersion" from the EnvoyVersionsURL.
	EnvoyVersion version.Version
	// HomeDir is an absolute path which most importantly contains "versions" installed from EnvoyVersionsURL. Defaults to DefaultHomeDir
	HomeDir string
	// Quiet means don't Logf to Out
	Quiet bool
	// Out is where status messages are written. Defaults to os.Stdout
	Out io.Writer
	// The platform to target for the Envoy install.
	Platform version.Platform
	// FuncEVersions is the interface for fetching Envoy release versions map from the EnvoyVersionsURL.
	FuncEVersions version.FuncEVersions
}

// Logf is used for shared functions that log conditionally on GlobalOpts.Quiet
func (o *GlobalOpts) Logf(format string, a ...interface{}) {
	if o.Quiet { // TODO: we may want to do scoped logging via a Context property, if this becomes common.
		return
	}
	moreos.Fprintf(o.Out, format, a...) //nolint
}

const (
	// DefaultEnvoyVersionsURL is the default value for GlobalOpts.EnvoyVersionsURL
	DefaultEnvoyVersionsURL = "https://archive.tetratelabs.io/envoy/envoy-versions.json"
	// DefaultEnvoyVersionsSchemaURL is the JSON schema used to validate GlobalOpts.EnvoyVersionsURL
	DefaultEnvoyVersionsSchemaURL = "https://archive.tetratelabs.io/release-versions-schema.json"
	// DefaultPlatform is the current platform of the host machine
	DefaultPlatform = version.Platform(runtime.GOOS + "/" + runtime.GOARCH)
)

var (
	// DefaultHomeDir is the default value for GlobalOpts.HomeDir
	DefaultHomeDir = moreos.ReplacePathSeparator("${HOME}/.func-e")
	// EnvoyVersionPattern is used to validate versions and is the same pattern as release-versions-schema.json.
	EnvoyVersionPattern = regexp.MustCompile(`^[1-9][0-9]*\.[0-9]+\.[0-9]+(_debug)?$`)
	// EnvoyMinorVersionPattern is EnvoyVersionPattern but with optional patch and _debug components.
	EnvoyMinorVersionPattern = regexp.MustCompile(`^[1-9][0-9]*\.[0-9]+(\.[0-9]+)?(_debug)?$`)
	// EnvoyStrictMinorVersionPattern is used to validated minor versions. A Minor version is just
	// like envoy.Version format, except missing the patch. For example: 1.18 or 1.20_debug.
	EnvoyStrictMinorVersionPattern = regexp.MustCompile(`^[1-9][0-9]*\.[0-9]+(_debug)?$`)
)
