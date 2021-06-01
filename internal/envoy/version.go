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

package envoy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tetratelabs/getenvoy/internal/globals"
	"github.com/tetratelabs/getenvoy/internal/version"
)

var (
	currentVersionVar            = "$ENVOY_VERSION"
	currentVersionWorkingDirFile = filepath.Join("$PWD", ".envoy-version")
	currentVersionHomeDirFile    = filepath.Join("$GETENVOY_HOME", "version")
)

// GetHomeVersion returns the default version in the "homeDir" and path to to it (homeVersionFile). When "v" is empty,
// homeVersionFile is not yet initialized.
func GetHomeVersion(homeDir string) (v, homeVersionFile string, err error) {
	v, homeVersionFile, err = getHomeVersion(homeDir)
	if err == nil && v == "" { // no home version, yet
		return
	}
	err = verifyVersion(v, currentVersionHomeDirFile, err)
	return
}

// CurrentVersion returns the first version in priority of VersionUsageList and its source or an error. The "source"
// and error messages returned include unexpanded variables to clarify the intended context.
func CurrentVersion(homeDir string) (v, source string, err error) {
	v, source, err = getCurrentVersion(homeDir)
	err = verifyVersion(v, source, err)
	return
}

func verifyVersion(v, source string, err error) error {
	if err != nil {
		return fmt.Errorf("couldn't read version from %s: %w", source, err)
	}
	if matched := globals.EnvoyVersionPattern.MatchString(v); !matched {
		return fmt.Errorf("invalid version in %q: %q should look like %q", source, v, version.LastKnownEnvoy)
	}
	return nil
}

func getCurrentVersion(homeDir string) (v, source string, err error) {
	// Priority 1: $ENVOY_VERSION
	if ev, ok := os.LookupEnv("ENVOY_VERSION"); ok {
		return ev, currentVersionVar, nil
	}

	// Priority 2: $PWD/.envoy-version
	data, err := os.ReadFile(".envoy-version")
	if err == nil {
		return string(data), currentVersionWorkingDirFile, nil
	} else if !os.IsNotExist(err) {
		return "", currentVersionWorkingDirFile, err
	}

	// Priority 3: $GETENVOY_HOME/version
	source = currentVersionHomeDirFile
	v, _, err = getHomeVersion(homeDir)
	return
}

func getHomeVersion(homeDir string) (v, homeVersionFile string, err error) {
	homeVersionFile = filepath.Join(homeDir, "version")
	var data []byte
	if data, err = os.ReadFile(homeVersionFile); err == nil {
		v = string(data)
	} else if os.IsNotExist(err) {
		err = nil // ok on file-not-found
	}
	return
}

// VersionUsageList is the priority order of Envoy version sources.
// This includes unresolved variables as it is both used statically for markdown generation, and also at runtime.
func VersionUsageList() string {
	return strings.Join([]string{currentVersionVar, currentVersionWorkingDirFile, currentVersionHomeDirFile}, ", ")
}
