// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tetratelabs/func-e/internal/version"
)

const (
	currentVersionVar = "$ENVOY_VERSION"
	// CurrentVersionWorkingDirFile is used for stable "versions" and "help" output
	CurrentVersionWorkingDirFile = "$PWD/.envoy-version"
	// CurrentVersionHomeDirFile is used for stable "versions" and "help" output
	CurrentVersionHomeDirFile = "$FUNC_E_HOME/version"
)

// WriteCurrentVersion writes the version to CurrentVersionWorkingDirFile or CurrentVersionHomeDirFile depending on
// if the former is present.
func WriteCurrentVersion(v version.Version, homeDir string) error {
	if _, err := os.Stat(".envoy-version"); os.IsNotExist(err) {
		if e := os.MkdirAll(homeDir, 0o750); e != nil {
			return e
		}
		return os.WriteFile(filepath.Join(homeDir, "version"), []byte(v.String()), 0o600)
	} else if err != nil {
		return err
	}
	return os.WriteFile(".envoy-version", []byte(v.String()), 0o600)
}

// CurrentVersion returns the first version in priority of VersionUsageList and its source or an error. The "source"
// and error messages returned include unexpanded variables to clarify the intended context.
// In the case no version was found, the version returned will be nil, not an error.
func CurrentVersion(homeDir string) (v version.Version, source string, err error) {
	s, source, err := getCurrentVersion(homeDir)
	v, err = verifyVersion(s, source, err)
	return
}

func verifyVersion(v, source string, err error) (version.Version, error) {
	if err != nil {
		return nil, fmt.Errorf(`couldn't read version from %s: %w`, source, err)
	} else if v == "" && source == CurrentVersionHomeDirFile { // don't error on initial state
		return nil, nil
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
	v, err = getHomeVersion(homeDir)
	return
}

func getHomeVersion(homeDir string) (v string, err error) {
	var data []byte
	if data, err = os.ReadFile(filepath.Join(homeDir, "version")); err == nil { //nolint:gosec
		v = strings.TrimSpace(string(data))
	} else if os.IsNotExist(err) {
		err = nil // ok on file-not-found
	}
	return
}

// VersionUsageList is the priority order of Envoy version sources.
// This includes unresolved variables as it is both used statically for markdown generation, and also at runtime.
func VersionUsageList() string {
	return strings.Join([]string{currentVersionVar, CurrentVersionWorkingDirFile, CurrentVersionHomeDirFile}, ", ")
}
