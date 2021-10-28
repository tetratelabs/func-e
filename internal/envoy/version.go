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

	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/version"
)

const (
	currentVersionVar = "$ENVOY_VERSION"
	// CurrentVersionWorkingDirFile is used for stable "versions" and "help" output
	CurrentVersionWorkingDirFile = "$PWD/.envoy-version"
	// CurrentVersionHomeDirFile is used for stable "versions" and "help" output
	CurrentVersionHomeDirFile = "$FUNC_E_HOME/version"
)

// GetHomeVersion returns the default version in the "homeDir" and path to to it (homeVersionFile). When "v" is empty,
// homeVersionFile is not yet initialized.
func GetHomeVersion(homeDir string) (v version.Version, homeVersionFile string, err error) {
	s, homeVersionFile, err := getHomeVersion(homeDir)
	if err == nil && s == "" { // no home version, yet
		return
	}
	v, err = verifyVersion(s, CurrentVersionHomeDirFile, err)
	return
}

// WriteCurrentVersion writes the version to CurrentVersionWorkingDirFile or CurrentVersionHomeDirFile depending on
// if the former is present.
func WriteCurrentVersion(v version.Version, homeDir string) error {
	if _, err := os.Stat(".envoy-version"); os.IsNotExist(err) {
		return os.WriteFile(filepath.Join(homeDir, "version"), []byte(v.String()), 0600)
	} else if err != nil {
		return err
	}
	return os.WriteFile(".envoy-version", []byte(v.String()), 0600)
}

// CurrentVersion returns the first version in priority of VersionUsageList and its source or an error. The "source"
// and error messages returned include unexpanded variables to clarify the intended context.
func CurrentVersion(homeDir string) (v version.Version, source string, err error) {
	s, source, err := getCurrentVersion(homeDir)
	v, err = verifyVersion(s, source, err)
	return
}

func verifyVersion(v, source string, err error) (version.Version, error) {
	if err != nil {
		return nil, moreos.Errorf(`couldn't read version from %s: %w`, source, err)
	}
	return version.NewVersion(fmt.Sprintf("version in %q", source), v)
}

func getCurrentVersion(homeDir string) (v, source string, err error) {
	// Priority 1: $ENVOY_VERSION
	if ev, ok := os.LookupEnv("ENVOY_VERSION"); ok {
		v = ev
		source = currentVersionVar
		return
	}

	// Priority 2: $PWD/.envoy-version
	data, err := os.ReadFile(".envoy-version")
	if err == nil {
		v = strings.TrimSpace(string(data))
		source = CurrentVersionWorkingDirFile
		return
	} else if !os.IsNotExist(err) {
		return "", CurrentVersionWorkingDirFile, err
	}

	// Priority 3: $FUNC_E_HOME/version
	source = CurrentVersionHomeDirFile
	v, _, err = getHomeVersion(homeDir)
	return
}

func getHomeVersion(homeDir string) (v, homeVersionFile string, err error) {
	homeVersionFile = filepath.Join(homeDir, "version")
	var data []byte
	if data, err = os.ReadFile(homeVersionFile); err == nil {
		v = strings.TrimSpace(string(data))
	} else if os.IsNotExist(err) {
		err = nil // ok on file-not-found
	}
	return
}

// VersionUsageList is the priority order of Envoy version sources.
// This includes unresolved variables as it is both used statically for markdown generation, and also at runtime.
func VersionUsageList() string {
	return moreos.ReplacePathSeparator(
		strings.Join([]string{currentVersionVar, CurrentVersionWorkingDirFile, CurrentVersionHomeDirFile}, ", "),
	)
}
