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

	"github.com/tetratelabs/getenvoy/internal/test/morerequire"
	"github.com/tetratelabs/getenvoy/internal/version"
)

func TestVersionUsageList(t *testing.T) {
	require.Equal(t, "$ENVOY_VERSION, $PWD/.envoy-version, $GETENVOY_HOME/version", VersionUsageList())
}

func TestGetHomeVersion(t *testing.T) {
	homeDir, removeHomeDir := morerequire.RequireNewTempDir(t)
	removeHomeDir() // delete the home directory

	t.Run("doesn't error on no home dir", func(t *testing.T) {
		v, homeVersionFile, err := GetHomeVersion(homeDir)
		require.Empty(t, v)
		require.Equal(t, filepath.Join(homeDir, "version"), homeVersionFile)
		require.NoError(t, err)
	})

	homeDir, removeHomeDir = morerequire.RequireNewTempDir(t)
	defer removeHomeDir()

	t.Run("doesn't error on no home version file", func(t *testing.T) {
		v, homeVersionFile, err := GetHomeVersion(homeDir)
		require.Empty(t, v)
		require.Equal(t, filepath.Join(homeDir, "version"), homeVersionFile)
		require.NoError(t, err)
	})

	homeVersionFile := filepath.Join(homeDir, "version")
	require.NoError(t, os.WriteFile(homeVersionFile, []byte(""), 0600))

	// This could be considered unexpected, but it is less to code to overwrite on this edge-case
	t.Run("doesn't error on no empty home version file", func(t *testing.T) {
		v, _, err := GetHomeVersion(homeDir)
		require.Empty(t, v)
		require.NoError(t, err)
	})

	require.NoError(t, os.WriteFile(homeVersionFile, []byte(version.LastKnownEnvoy), 0600))
	t.Run("reads valid home version file", func(t *testing.T) {
		v, _, err := GetHomeVersion(homeDir)
		require.Equal(t, version.LastKnownEnvoy, v)
		require.NoError(t, err)
	})

	require.NoError(t, os.WriteFile(homeVersionFile, []byte("a.a.a"), 0600))
	t.Run("errors on invalid home version file", func(t *testing.T) {
		_, _, err := GetHomeVersion(homeDir)
		require.EqualError(t, err, fmt.Sprintf(`invalid version in "$GETENVOY_HOME/version": "a.a.a" should look like "%s"`, version.LastKnownEnvoy))
	})
}

func TestCurrentVersion(t *testing.T) {
	homeDir, removeHomeDir := morerequire.RequireNewTempDir(t)
	defer removeHomeDir()
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, "version"), []byte("1.1.1"), 0600))

	t.Run("defaults to home version", func(t *testing.T) {
		v, source, err := CurrentVersion(homeDir)
		require.Equal(t, "1.1.1", v)
		require.Equal(t, "$GETENVOY_HOME/version", source)
		require.NoError(t, err)
	})

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(wd) //nolint

	tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
	defer removeTempDir()
	require.NoError(t, os.Chdir(tempDir))
	require.NoError(t, os.WriteFile(".envoy-version", []byte("2.2.2"), 0600))

	t.Run("prefers $PWD/.envoy-version over home version", func(t *testing.T) {
		v, source, err := CurrentVersion(homeDir)
		require.Equal(t, "2.2.2", v)
		require.Equal(t, "$PWD/.envoy-version", source)
		require.NoError(t, err)
	})

	revert := morerequire.RequireSetenv(t, "ENVOY_VERSION", "3.3.3")
	defer revert()

	t.Run("prefers $ENVOY_VERSION over $PWD/.envoy-version", func(t *testing.T) {
		v, source, err := CurrentVersion(homeDir)
		require.Equal(t, "3.3.3", v)
		require.Equal(t, "$ENVOY_VERSION", source)
		require.NoError(t, err)
	})
}

func TestCurrentVersion_Validates(t *testing.T) {
	homeDir, removeHomeDir := morerequire.RequireNewTempDir(t)
	defer removeHomeDir()
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, "version"), []byte("a.a.a"), 0600))

	t.Run("validates home version", func(t *testing.T) {
		_, _, err := CurrentVersion(homeDir)
		require.EqualError(t, err, fmt.Sprintf(`invalid version in "$GETENVOY_HOME/version": "a.a.a" should look like "%s"`, version.LastKnownEnvoy))
	})

	wd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(wd) //nolint

	tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
	defer removeTempDir()
	require.NoError(t, os.Chdir(tempDir))
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
