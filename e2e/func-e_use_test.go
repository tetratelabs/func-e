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

package e2e

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/version"
)

// TestFuncEUse needs to always execute, so we run it in a separate home directory
func TestFuncEUse(t *testing.T) {
	homeDir := t.TempDir()

	t.Run("not yet installed", func(t *testing.T) {
		stdout, stderr, err := funcEExec("--home-dir", homeDir, "use", string(version.LastKnownEnvoy))

		require.NoError(t, err)
		require.Regexp(t, `^downloading https:.*tar.*z\r?\n$`, stdout)
		require.Empty(t, stderr)

		// The binary was installed
		envoyBin := filepath.Join(homeDir, "versions", string(version.LastKnownEnvoy), "bin", "envoy"+moreos.Exe)
		require.FileExists(t, envoyBin)

		// The current version was written
		f, err := os.ReadFile(filepath.Join(homeDir, "version"))
		require.NoError(t, err)
		require.Equal(t, version.LastKnownEnvoy, version.Version(f))
	})

	t.Run("already installed", func(t *testing.T) {
		stdout, stderr, err := funcEExec("--home-dir", homeDir, "use", string(version.LastKnownEnvoy))

		require.NoError(t, err)
		require.Equal(t, moreos.Sprintf("%s is already downloaded\n", version.LastKnownEnvoy), stdout)
		require.Empty(t, stderr)
	})
}

func TestFuncEUse_UnknownVersion(t *testing.T) {
	v := "1.1.1"
	stdout, stderr, err := funcEExec("use", v)

	require.EqualError(t, err, "exit status 1")
	require.Empty(t, stdout)
	require.Equal(t, moreos.Sprintf(`error: couldn't find version "%s" for platform "%s/%s"
`, v, runtime.GOOS, runtime.GOARCH), stderr)
}

func TestFuncEUse_MinorVersion(t *testing.T) {
	homeDir := t.TempDir()

	t.Run("install last known", func(t *testing.T) {
		stdout, stderr, err := funcEExec("--home-dir", homeDir, "use", string(version.LastKnownEnvoy))

		require.NoError(t, err)
		require.Regexp(t, `^downloading https:.*tar.*z\r?\n$`, stdout)
		require.Empty(t, stderr)

		// The binary was installed.
		envoyBin := filepath.Join(homeDir, "versions", string(version.LastKnownEnvoy), "bin", "envoy"+moreos.Exe)
		require.FileExists(t, envoyBin)

		// The current version was written.
		f, err := os.ReadFile(filepath.Join(homeDir, "version"))
		require.NoError(t, err)
		require.Equal(t, version.LastKnownEnvoy, version.Version(f))
	})

	baseVersion := version.Version("1.12.0")
	t.Run("install base version", func(t *testing.T) {
		stdout, stderr, err := funcEExec("--home-dir", homeDir, "use", string(baseVersion))

		require.NoError(t, err)
		require.Regexp(t, `^downloading https:.*tar.*z\r?\n$`, stdout)
		require.Empty(t, stderr)

		// The binary was installed.
		envoyBin := filepath.Join(homeDir, "versions", string(baseVersion), "bin", "envoy"+moreos.Exe)
		require.FileExists(t, envoyBin)

		// The base version was written.
		f, err := os.ReadFile(filepath.Join(homeDir, "version"))
		require.NoError(t, err)
		require.Equal(t, baseVersion, version.Version(f))
	})

	// The intended minor version to be installed.
	minorVersion := version.Version("1.12")

	// This version is EOL'd based on: https://github.com/envoyproxy/envoy/blob/main/RELEASES.md#release-schedule.
	upgradedVersion := version.Version("1.12.7")
	t.Run("install upgraded version", func(t *testing.T) {
		stdout, stderr, err := funcEExec("--home-dir", homeDir, "use", string(minorVersion))

		require.NoError(t, err)
		require.Regexp(t, `^downloading https:.*tar.*z\r?\n$`, stdout)
		require.Empty(t, stderr)

		// The binary was installed.
		envoyBin := filepath.Join(homeDir, "versions", string(upgradedVersion), "bin", "envoy"+moreos.Exe)
		require.FileExists(t, envoyBin)

		// The upgraded version was written.
		f, err := os.ReadFile(filepath.Join(homeDir, "version"))
		require.NoError(t, err)
		require.Equal(t, minorVersion, version.Version(f))
	})

	t.Run("use upgraded version after downloaded", func(t *testing.T) {
		stdout, stderr, err := funcEExec("--home-dir", homeDir, "use", string(minorVersion))
		require.NoError(t, err)
		require.Equal(t, moreos.Sprintf("%s is already downloaded\n", upgradedVersion), stdout)
		require.Empty(t, stderr)
	})

	t.Run("which upgraded version", func(t *testing.T) {
		stdout, stderr, err := funcEExec("--home-dir", homeDir, "which")
		relativeEnvoyBin := filepath.Join("versions", string(upgradedVersion), "bin", "envoy"+moreos.Exe)
		require.Contains(t, stdout, moreos.Sprintf("%s\n", relativeEnvoyBin))
		require.Empty(t, stderr)
		require.NoError(t, err)
	})
}
