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
)

// RunOpts support invocations of "getenvoy run"
type RunOpts struct {
	// EnvoyPath is the exec.Cmd path to "envoy". Defaults to "$HomeDir/versions/$version/bin/envoy"
	EnvoyPath string
	// RunDir is the location any generated configuration or debug files are written.
	// This is not Envoy's working directory, which remains the same as the $PWD of getenvoy.
	//
	// Upon termination, this directory is archived as "../$(basename $RunDir).tar.gz"
	// Defaults to "$HomeDir/runs/$epochtime"
	RunDir string
	// DontArchiveRunDir is used in testing and prevents archiving the RunDir
	DontArchiveRunDir bool
}

// GlobalOpts represents options that affect more than one getenvoy commands.
//
// Fields representing non-hidden flags have values set according to the following rules:
//  1) value that precedes flag parsing, used in tests
//  2) to a value of the command line argument, e.g. `--home-dir`
//  3) optional mapping to an environment variable, e.g. `GETENVOY_HOME` (not all flags are mapped to ENV)
//  4) otherwise, to the default value, e.g. DefaultHomeDir
type GlobalOpts struct {
	// RunOpts are inlined to allow tests to override parameters without changing ENV variables or flags
	RunOpts
	// EnvoyVersionsURL is the path to the envoy-versions.json. Defaults to DefaultEnvoyVersionsURL
	EnvoyVersionsURL string
	// EnvoyVersion is the default version of Envoy to run. Defaults to the contents of "$HomeDir/versions/version".
	// When that file is missing, it is generated from ".latestVersion" from the EnvoyVersionsURL.
	EnvoyVersion string
	// HomeDir is an absolute path which most importantly contains "versions" installed from EnvoyVersionsURL. Defaults to DefaultHomeDir
	HomeDir string
	// Out is where status messages are written. Defaults to os.Stdout
	Out io.Writer
}

const (
	// DefaultHomeDir is the default value for GlobalOpts.HomeDir
	DefaultHomeDir = "${HOME}/.getenvoy"
	// DefaultEnvoyVersionsURL is the default value for GlobalOpts.EnvoyVersionsURL
	DefaultEnvoyVersionsURL = "https://getenvoy.io/envoy-versions.json"
)

var (
	// EnvoyVersionPattern is used to validate versions and is the same pattern as envoy-versions-schema.json.
	EnvoyVersionPattern = regexp.MustCompile(`^[1-9][0-9]*\.[0-9]+\.[0-9]+$`)
	// CurrentPlatform is the platform of the current process. This is used as a key in EnvoyVersion.Tarballs.
	CurrentPlatform = runtime.GOOS + "/" + runtime.GOARCH
)
