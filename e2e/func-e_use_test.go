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
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/version"
)

// TestFuncEUse needs to always execute, so we run it in a separate home directory
func TestFuncEUse(t *testing.T) {
	homeDir := t.TempDir()

	t.Run("not yet installed", func(t *testing.T) {
		stdout, stderr, err := funcEExec("--home-dir", homeDir, "use", version.LastKnownEnvoy.String())
		require.NoError(t, err)
		require.Regexp(t, `^downloading https:.*tar.*z\r?\n$`, stdout)
		require.Empty(t, stderr)

		// The binary was installed
		envoyBin := filepath.Join(homeDir, "versions", version.LastKnownEnvoy.String(), "bin", "envoy"+moreos.Exe)
		require.FileExists(t, envoyBin)

		// The current version was written
		f, err := os.ReadFile(filepath.Join(homeDir, "version"))
		require.NoError(t, err)
		require.Equal(t, version.LastKnownEnvoy, version.PatchVersion(f))
	})

	t.Run("already installed", func(t *testing.T) {
		stdout, stderr, err := funcEExec("--home-dir", homeDir, "use", version.LastKnownEnvoy.String())

		require.NoError(t, err)
		require.Equal(t, moreos.Sprintf("%s is already downloaded\n", version.LastKnownEnvoy.String()), stdout)
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

func TestFuncEUse_UnknownMinorVersion(t *testing.T) {
	v := "1.1"
	stdout, stderr, err := funcEExec("use", v)

	require.EqualError(t, err, "exit status 1")
	require.Regexp(t, `^looking up the latest patch for Envoy version 1.1\r?\n$`, stdout)
	stderrPattern := fmt.Sprintf("^error: https://.*json does not contain an Envoy release for version 1.1 on platform %s/%s\r?\n$", runtime.GOOS, runtime.GOARCH)
	require.Regexp(t, stderrPattern, stderr)
}

// TODO: this test overuses bandwidth, making it tedious especially on slow networks.
// Parse the envoy-versions.json to dynamically select the version before LastKnownEnvoyMinor if it isn't consistent.
// That or don't update LastKnownEnvoy until it is consistent.
func TestFuncEUse_MinorVersion(t *testing.T) {
	// The intended minor version to be installed. This version is known to have darwin, linux, and windows binaries.
	minorVersion := "1.26"

	allVersions, _, err := funcEExec("versions", "-a")
	require.NoError(t, err)

	baseVersion, upgradedVersion := getVersionsRange(allVersions, minorVersion)

	homeDir := t.TempDir()

	t.Run("install last known", func(t *testing.T) {
		stdout, stderr, err := funcEExec("--home-dir", homeDir, "use", version.LastKnownEnvoy.String())

		require.NoError(t, err)
		require.Regexp(t, `^downloading https:.*tar.*z\r?\n$`, stdout)
		require.Empty(t, stderr)

		// The binary was installed.
		envoyBin := filepath.Join(homeDir, "versions", version.LastKnownEnvoy.String(), "bin", "envoy"+moreos.Exe)
		require.FileExists(t, envoyBin)

		// The current version was written.
		f, err := os.ReadFile(filepath.Join(homeDir, "version"))
		require.NoError(t, err)
		require.Equal(t, version.LastKnownEnvoy, version.PatchVersion(f))
	})

	t.Run(fmt.Sprintf("install %s as base version", baseVersion), func(t *testing.T) {
		stdout, stderr, err := funcEExec("--home-dir", homeDir, "use", baseVersion)

		require.NoError(t, err)
		require.Regexp(t, `^downloading https:.*tar.*z\r?\n$`, stdout)
		require.Empty(t, stderr)

		// The binary was installed.
		envoyBin := filepath.Join(homeDir, "versions", baseVersion, "bin", "envoy"+moreos.Exe)
		require.FileExists(t, envoyBin)

		// The base version was written.
		f, err := os.ReadFile(filepath.Join(homeDir, "version"))
		require.NoError(t, err)
		require.Equal(t, baseVersion, string(f))
	})

	t.Run(fmt.Sprintf("install %s as upgraded version", upgradedVersion), func(t *testing.T) {
		stdout, stderr, err := funcEExec("--home-dir", homeDir, "use", minorVersion)

		require.NoError(t, err)
		require.Regexp(t, `^looking up the latest patch for Envoy version 1.26\r?\ndownloading https:.*tar.*z\r?\n$`, stdout)
		require.Empty(t, stderr)

		// The binary was installed.
		envoyBin := filepath.Join(homeDir, "versions", upgradedVersion, "bin", "envoy"+moreos.Exe)
		require.FileExists(t, envoyBin)

		// The upgraded version was written.
		f, err := os.ReadFile(filepath.Join(homeDir, "version"))
		require.NoError(t, err)
		require.Equal(t, minorVersion, string(f))
	})

	t.Run("use upgraded version after downloaded", func(t *testing.T) {
		stdout, stderr, err := funcEExec("--home-dir", homeDir, "use", minorVersion)
		require.NoError(t, err)
		require.Equal(t, moreos.Sprintf("looking up the latest patch for Envoy version 1.26\n%s is already downloaded\n", upgradedVersion), stdout)
		require.Empty(t, stderr)
	})

	t.Run("which upgraded version", func(t *testing.T) {
		stdout, stderr, err := funcEExec("--home-dir", homeDir, "which")
		relativeEnvoyBin := filepath.Join("versions", upgradedVersion, "bin", "envoy"+moreos.Exe)
		require.Contains(t, stdout, moreos.Sprintf("%s\n", relativeEnvoyBin))
		require.Empty(t, stderr)
		require.NoError(t, err)
	})
}

// getVersionsRange returns the first and latest patch of a minor version.
func getVersionsRange(stdout, minor string) (first, latest string) {
	s := bufio.NewScanner(strings.NewReader(stdout))
	for s.Scan() {
		// Ex. "  1.20.0 2021-10-05" or "* 1.20.0 2021-10-05" -> "1.20.0"
		v := strings.Split(s.Text()[2:], " ")[0]
		if strings.HasPrefix(v, minor+".") {
			// "func-e versions" returns in descending order
			if latest == "" {
				latest = v
			} else {
				first = v
			}
		}
	}
	return
}
