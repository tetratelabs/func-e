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

import "log"

// RunOpts support invocations of "getenvoy run" and "getenvoy extension run"
type RunOpts struct {
	// EnvoyPath is the exec.Cmd path to "envoy". Defaults to "$HomeDir/builds/$flavor/$version/$platform/bin/envoy"
	EnvoyPath string
	// WorkingDir is the working directory of EnvoyPath and includes any generated configuration or debug files.
	// Upon termination, this directory is archived as "../$(basename $WorkingDir).tar.gz"
	// Defaults to "$HomeDir/debug/$epochtime"
	WorkingDir string
	// Log generates output users should be aware of. Defaults to os.Stdout with "run: " prefix.
	Log *log.Logger
	// DebugLog generates output for debug features such as access logging. Defaults to os.Stdout with "debug: " prefix.
	DebugLog *log.Logger
}

// GlobalOpts represents options that affect more than one getenvoy commands.
//
// Fields representing non-hidden flags have values set according to the following rules:
//  1) value that precedes flag parsing, used in tests
//  2) to a value of the command line argument, e.g. `--home-dir`
//  3) optional mapping to an environment variable, e.g. `GETENVOY_HOME` (not all flags are mapped to ENV)
//  4) otherwise, to the default value, e.g. `${HOME}/.getenvoy`
type GlobalOpts struct {
	// RunOpts are inlined to allow tests to override parameters without changing ENV variables or flags
	RunOpts
	// HomeDir most importantly contains envoy binaries fetched from ManifestURL. Defaults to $HOME/.getenvoy
	HomeDir string
	// ManifestURL is the path to the getenvoy manifest json
	ManifestURL string
}

// DefaultManifestURL is the default value for GlobalOpts.ManifestURL
const DefaultManifestURL = "https://dl.getenvoy.io/public/raw/files/manifest.json"
