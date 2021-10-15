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
		stdout, stderr, err := funcEExec("--home-dir", homeDir, "use", version.LastKnownEnvoy)

		require.NoError(t, err)
		require.Regexp(t, `^downloading https:.*tar.*z\r?\n$`, stdout)
		require.Empty(t, stderr)

		// The binary was installed
		envoyBin := filepath.Join(homeDir, "versions", version.LastKnownEnvoy, "bin", "envoy"+moreos.Exe)
		require.FileExists(t, envoyBin)

		// The current version was written
		f, err := os.ReadFile(filepath.Join(homeDir, "version"))
		require.NoError(t, err)
		require.Equal(t, version.LastKnownEnvoy, f)
	})

	t.Run("already installed", func(t *testing.T) {
		stdout, stderr, err := funcEExec("--home-dir", homeDir, "use", version.LastKnownEnvoy)

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

func TestFuncEUse_UnknownMinorVersion(t *testing.T) {
	v := "1.1"
	stdout, stderr, err := funcEExec("use", v)

	require.EqualError(t, err, "exit status 1")
	require.Empty(t, stdout)
	require.Equal(t, moreos.Sprintf(`error: couldn't find the latest patch for "%s" for platform "%s/%s"
`, v, runtime.GOOS, runtime.GOARCH), stderr)
}

func TestFuncEUse_MinorVersion(t *testing.T) {
	// The intended minor version to be installed. This version is known to have darwin, linux, and windows binaries.
	minorVersion := "1.18"

	allVersions, _, err := funcEExec("versions", "-a")
	require.NoError(t, err)

	baseVersion, upgradedVersion := getVersionsRange(allVersions, minorVersion)

	homeDir := t.TempDir()

	t.Run("install last known", func(t *testing.T) {
		stdout, stderr, err := funcEExec("--home-dir", homeDir, "use", version.LastKnownEnvoy)

		require.NoError(t, err)
		require.Regexp(t, `^downloading https:.*tar.*z\r?\n$`, stdout)
		require.Empty(t, stderr)

		// The binary was installed.
		envoyBin := filepath.Join(homeDir, "versions", version.LastKnownEnvoy, "bin", "envoy"+moreos.Exe)
		require.FileExists(t, envoyBin)

		// The current version was written.
		f, err := os.ReadFile(filepath.Join(homeDir, "version"))
		require.NoError(t, err)
		require.Equal(t, version.LastKnownEnvoy, f)
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
		require.Equal(t, baseVersion, f)
	})

	t.Run(fmt.Sprintf("install %s as upgraded version", upgradedVersion), func(t *testing.T) {
		stdout, stderr, err := funcEExec("--home-dir", homeDir, "use", minorVersion)

		require.NoError(t, err)
		require.Regexp(t, `^downloading https:.*tar.*z\r?\n$`, stdout)
		require.Empty(t, stderr)

		// The binary was installed.
		envoyBin := filepath.Join(homeDir, "versions", upgradedVersion, "bin", "envoy"+moreos.Exe)
		require.FileExists(t, envoyBin)

		// The upgraded version was written.
		f, err := os.ReadFile(filepath.Join(homeDir, "version"))
		require.NoError(t, err)
		require.Equal(t, minorVersion, f)
	})

	t.Run("use upgraded version after downloaded", func(t *testing.T) {
		stdout, stderr, err := funcEExec("--home-dir", homeDir, "use", minorVersion)
		require.NoError(t, err)
		require.Equal(t, moreos.Sprintf("%s is already downloaded\n", upgradedVersion), stdout)
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
	var rows []string
	for s.Scan() {
		row := strings.TrimSpace(s.Text())
		if strings.HasPrefix(row, minor+".") {
			rows = append(rows, row[:strings.Index(row, " ")])
		}
	}
	// The rows is sorted in descending order.
	first = rows[len(rows)-1]
	latest = rows[0]
	return
}
