// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/test/morerequire"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestVersionUsageList(t *testing.T) {
	expected := moreos.ReplacePathSeparator("$ENVOY_VERSION, $PWD/.envoy-version, $FUNC_E_HOME/version")
	require.Equal(t, expected, VersionUsageList())
}

func TestWriteCurrentVersion_HomeDir(t *testing.T) {
	homeDir := t.TempDir()

	for _, tt := range []struct{ name, v string }{
		{"writes initial home version", "1.1.1"},
		{"overwrites home version", "2.2.2"},
	} {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, WriteCurrentVersion(version.PatchVersion(tc.v), homeDir))
			v, src, err := getCurrentVersion(homeDir)
			require.NoError(t, err)
			require.Equal(t, tc.v, v)
			require.Equal(t, CurrentVersionHomeDirFile, src)
			require.NoFileExists(t, ".envoy-version")
		})
	}
}

func TestWriteCurrentVersion_OverwritesWorkingDirVersion(t *testing.T) {
	homeDir := t.TempDir()

	homeVersionFile := filepath.Join(homeDir, "version")
	require.NoError(t, os.WriteFile(homeVersionFile, []byte("1.1.1"), 0o600))

	revertWd := morerequire.RequireChdir(t, t.TempDir())
	defer revertWd()
	require.NoError(t, os.WriteFile(".envoy-version", []byte("2.2.2"), 0o600))

	require.NoError(t, WriteCurrentVersion(version.PatchVersion("3.3.3"), homeDir))
	v, src, err := getCurrentVersion(homeDir)
	require.NoError(t, err)
	require.Equal(t, "3.3.3", v)
	require.Equal(t, CurrentVersionWorkingDirFile, src)

	// didn't overwrite the home version
	v, err = getHomeVersion(homeDir)
	require.NoError(t, err)
	require.Equal(t, "1.1.1", v)
}

// TestCurrentVersion is intentionally written in priority order instead of via a matrix. This particularly helps with
// test setup complexity required to ensure tiered priority (ex layering overridden PWD with an ENV)
func TestCurrentVersion(t *testing.T) {
	homeDir := t.TempDir()

	t.Run("defaults to nil", func(t *testing.T) {
		v, source, err := CurrentVersion(homeDir)
		require.Nil(t, v)
		require.Equal(t, CurrentVersionHomeDirFile, source)
		require.NoError(t, err)
	})

	require.NoError(t, os.WriteFile(filepath.Join(homeDir, "version"), []byte("1.1.1"), 0o600))
	t.Run("reads the home version", func(t *testing.T) {
		v, source, err := CurrentVersion(homeDir)
		require.Equal(t, version.PatchVersion("1.1.1"), v)
		require.Equal(t, CurrentVersionHomeDirFile, source)
		require.NoError(t, err)
	})

	revertWd := morerequire.RequireChdir(t, t.TempDir())
	defer revertWd()
	require.NoError(t, os.WriteFile(".envoy-version", []byte("2.2.2"), 0o600))

	t.Run("prefers $PWD/.envoy-version over home version", func(t *testing.T) {
		v, source, err := CurrentVersion(homeDir)
		require.Equal(t, version.PatchVersion("2.2.2"), v)
		require.Equal(t, CurrentVersionWorkingDirFile, source)
		require.NoError(t, err)
	})

	t.Setenv("ENVOY_VERSION", "3.3.3")

	t.Run("prefers $ENVOY_VERSION over $PWD/.envoy-version", func(t *testing.T) {
		v, source, err := CurrentVersion(homeDir)
		require.Equal(t, version.PatchVersion("3.3.3"), v)
		require.Equal(t, currentVersionVar, source)
		require.NoError(t, err)
	})
}

// TestCurrentVersion_Validates is intentionally written in priority order instead of via a matrix
func TestCurrentVersion_Validates(t *testing.T) {
	homeDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, "version"), []byte("a.a.a"), 0o600))

	t.Run("validates home version", func(t *testing.T) {
		_, _, err := CurrentVersion(homeDir)
		expectedErr := fmt.Sprintf(`invalid version in "$FUNC_E_HOME/version": "a.a.a" should look like %q or %q`, version.LastKnownEnvoy, version.LastKnownEnvoyMinor)
		require.EqualError(t, err, moreos.ReplacePathSeparator(expectedErr))
	})

	revertWd := morerequire.RequireChdir(t, t.TempDir())
	defer revertWd()
	require.NoError(t, os.WriteFile(".envoy-version", []byte("b.b.b"), 0o600))

	t.Run("validates $PWD/.envoy-version", func(t *testing.T) {
		_, _, err := CurrentVersion(homeDir)
		expectedErr := fmt.Sprintf(`invalid version in "$PWD/.envoy-version": "b.b.b" should look like %q or %q`, version.LastKnownEnvoy, version.LastKnownEnvoyMinor)
		require.EqualError(t, err, moreos.ReplacePathSeparator(expectedErr))
	})

	require.NoError(t, os.Remove(".envoy-version"))
	require.NoError(t, os.Mkdir(".envoy-version", 0o700))

	t.Run("shows error reading $PWD/.envoy-version", func(t *testing.T) {
		_, _, err := CurrentVersion(homeDir)
		expectedErr := moreos.ReplacePathSeparator("couldn't read version from $PWD/.envoy-version")
		require.Contains(t, err.Error(), expectedErr)
	})

	t.Setenv("ENVOY_VERSION", "c.c.c")

	t.Run("validates $ENVOY_VERSION", func(t *testing.T) {
		_, _, err := CurrentVersion(homeDir)
		require.EqualError(t, err, fmt.Sprintf(`invalid version in "$ENVOY_VERSION": "c.c.c" should look like %q or %q`, version.LastKnownEnvoy, version.LastKnownEnvoyMinor))
	})
}
