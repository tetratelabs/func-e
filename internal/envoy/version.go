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
	versionVarName      = "$ENVOY_VERSION"
	wdVersionFileName   = filepath.Join("$PWD", ".envoy-version")
	homeVersionFileName = filepath.Join("$GETENVOY_HOME", "version")
)

// CurrentVersion returns the first version in priority of VersionUsageList and its source or an error.
func CurrentVersion(homeVersion string) (string, string, error) {
	v, source, err := getCurrentVersion(homeVersion)
	if err != nil {
		return "", "", fmt.Errorf("couldn't read version from %s: %w", source, err)
	}

	if matched := globals.EnvoyVersionPattern.MatchString(v); !matched {
		return "", "", fmt.Errorf("invalid version in %q: %q should look like %q", source, v, version.LastKnownEnvoy)
	}

	return v, source, err
}

func getCurrentVersion(homeVersion string) (v, source string, err error) {
	// Priority 3: $ENVOY_VERSION
	source = versionVarName
	if v = os.Getenv("ENVOY_VERSION"); v != "" {
		return
	}

	// Priority 2: $PWD/.envoy-version
	source = wdVersionFileName
	var data []byte
	data, err = os.ReadFile(".envoy-version")
	if err == nil {
		v = string(data)
		return
	} else if !os.IsNotExist(err) {
		return
	}

	// Priority 3: $GETENVOY_HOME/versions
	source = homeVersionFileName
	v = homeVersion
	err = nil
	return
}

// VersionUsageList is the priority order of Envoy version sources.
// This includes unresolved variables as it is both used statically for markdown generation, and also at runtime.
func VersionUsageList() string {
	return strings.Join([]string{versionVarName, wdVersionFileName, homeVersionFileName}, ", ")
}
