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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/test/morerequire"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestVersionUsageList(t *testing.T) {
	require.Equal(t, "$ENVOY_VERSION, $PWD/.envoy-version, $FUNC_E_HOME/version", VersionUsageList())
}

func TestGetHomeVersion_Empty(t *testing.T) {
	homeDir, removeHomeDir := morerequire.RequireNewTempDir(t)
	defer removeHomeDir()

	homeVersionFile := filepath.Join(homeDir, "version")

	for _, tt := range []struct {
		name  string
		setup func()
	}{
		{"empty home version file", func() {
			require.NoError(t, os.WriteFile(homeVersionFile, []byte(""), 0600))
		}},
		{"missing home version file", func() {
			require.NoError(t, os.Remove(homeVersionFile))
		}},
		{"missing home dir", removeHomeDir},
	} {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			v, hvf, err := GetHomeVersion(homeDir)
			require.Empty(t, v)
			require.Equal(t, homeVersionFile, hvf)
			require.NoError(t, err)
		})
	}
}

func TestGetHomeVersion(t *testing.T) {
	homeDir, removeHomeDir := morerequire.RequireNewTempDir(t)
	defer removeHomeDir()

	homeVersionFile := filepath.Join(homeDir, "version")
	require.NoError(t, os.WriteFile(homeVersionFile, []byte(version.LastKnownEnvoy), 0600))

	v, hvf, err := GetHomeVersion(homeDir)
	require.Equal(t, version.LastKnownEnvoy, v)
	require.Equal(t, homeVersionFile, hvf)
	require.NoError(t, err)
}

func TestGetHomeVersion_Validates(t *testing.T) {
	homeDir, removeHomeDir := morerequire.RequireNewTempDir(t)
	defer removeHomeDir()

	homeVersionFile := filepath.Join(homeDir, "version")
	require.NoError(t, os.WriteFile(homeVersionFile, []byte("a.a.a"), 0600))

	_, _, err := GetHomeVersion(homeDir)
	require.EqualError(t, err, fmt.Sprintf(`invalid version in "%s": "a.a.a" should look like "%s"`, CurrentVersionHomeDirFile, version.LastKnownEnvoy))
}

func TestWriteCurrentVersion_HomeDir(t *testing.T) {
	homeDir, removeHomeDir := morerequire.RequireNewTempDir(t)
	defer removeHomeDir()

	for _, tt := range []struct {
		name string
		v    version.Version
	}{
		{"writes initial home version", "1.1.1"},
		{"overwrites home version", "2.2.2"},
	} {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, WriteCurrentVersion(tc.v, homeDir))
			v, src, err := getCurrentVersion(homeDir)
			require.NoError(t, err)
			require.Equal(t, tc.v, v)
			require.Equal(t, CurrentVersionHomeDirFile, src)
			require.NoFileExists(t, ".envoy-version")
		})
	}
}

func TestWriteCurrentVersion_OverwritesWorkingDirVersion(t *testing.T) {
	homeDir, removeHomeDir := morerequire.RequireNewTempDir(t)
	defer removeHomeDir()

	homeVersionFile := filepath.Join(homeDir, "version")
	require.NoError(t, os.WriteFile(homeVersionFile, []byte("1.1.1"), 0600))

	revertTempWd := morerequire.RequireChdirIntoTemp(t)
	defer revertTempWd()
	require.NoError(t, os.WriteFile(".envoy-version", []byte("2.2.2"), 0600))

	require.NoError(t, WriteCurrentVersion("3.3.3", homeDir))
	v, src, err := getCurrentVersion(homeDir)
	require.NoError(t, err)
	require.Equal(t, version.Version("3.3.3"), v)
	require.Equal(t, CurrentVersionWorkingDirFile, src)

	// didn't overwrite the home version
	v, _, err = getHomeVersion(homeDir)
	require.NoError(t, err)
	require.Equal(t, version.Version("1.1.1"), v)
}

// TestCurrentVersion is intentionally written in priority order instead of via a matrix. This particularly helps with
// test setup complexity required to ensure tiered priority (ex layering overridden PWD with an ENV)
func TestCurrentVersion(t *testing.T) {
	homeDir, removeHomeDir := morerequire.RequireNewTempDir(t)
	defer removeHomeDir()
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, "version"), []byte("1.1.1"), 0600))

	t.Run("defaults to home version", func(t *testing.T) {
		v, source, err := CurrentVersion(homeDir)
		require.Equal(t, version.Version("1.1.1"), v)
		require.Equal(t, CurrentVersionHomeDirFile, source)
		require.NoError(t, err)
	})

	revertTempWd := morerequire.RequireChdirIntoTemp(t)
	defer revertTempWd()
	require.NoError(t, os.WriteFile(".envoy-version", []byte("2.2.2"), 0600))

	t.Run("prefers $PWD/.envoy-version over home version", func(t *testing.T) {
		v, source, err := CurrentVersion(homeDir)
		require.Equal(t, version.Version("2.2.2"), v)
		require.Equal(t, CurrentVersionWorkingDirFile, source)
		require.NoError(t, err)
	})

	revert := morerequire.RequireSetenv(t, "ENVOY_VERSION", "3.3.3")
	defer revert()

	t.Run("prefers $ENVOY_VERSION over $PWD/.envoy-version", func(t *testing.T) {
		v, source, err := CurrentVersion(homeDir)
		require.Equal(t, version.Version("3.3.3"), v)
		require.Equal(t, currentVersionVar, source)
		require.NoError(t, err)
	})
}

// TestCurrentVersion_Validates is intentionally written in priority order instead of via a matrix
func TestCurrentVersion_Validates(t *testing.T) {
	homeDir, removeHomeDir := morerequire.RequireNewTempDir(t)
	defer removeHomeDir()
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, "version"), []byte("a.a.a"), 0600))

	t.Run("validates home version", func(t *testing.T) {
		_, _, err := CurrentVersion(homeDir)
		require.EqualError(t, err, fmt.Sprintf(`invalid version in "$FUNC_E_HOME/version": "a.a.a" should look like "%s"`, version.LastKnownEnvoy))
	})

	revertTempWd := morerequire.RequireChdirIntoTemp(t)
	defer revertTempWd()
	require.NoError(t, os.WriteFile(".envoy-version", []byte("b.b.b"), 0600))

	t.Run("validates $PWD/.envoy-version", func(t *testing.T) {
		_, _, err := CurrentVersion(homeDir)
		require.EqualError(t, err, fmt.Sprintf(`invalid version in "$PWD/.envoy-version": "b.b.b" should look like "%s"`, version.LastKnownEnvoy))
	})

	require.NoError(t, os.Remove(".envoy-version"))
	require.NoError(t, os.Mkdir(".envoy-version", 0700))

	t.Run("shows error reading $PWD/.envoy-version", func(t *testing.T) {
		_, _, err := CurrentVersion(homeDir)
		require.Contains(t, err.Error(), "couldn't read version from $PWD/.envoy-version")
	})

	revert := morerequire.RequireSetenv(t, "ENVOY_VERSION", "c.c.c")
	defer revert()

	t.Run("validates $ENVOY_VERSION", func(t *testing.T) {
		_, _, err := CurrentVersion(homeDir)
		require.EqualError(t, err, fmt.Sprintf(`invalid version in "$ENVOY_VERSION": "c.c.c" should look like "%s"`, version.LastKnownEnvoy))
	})
}
