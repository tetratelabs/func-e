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

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/test/morerequire"
	"github.com/tetratelabs/getenvoy/internal/version"
)

func TestGetInstalledVersions(t *testing.T) {
	homeDir, removeHomeDir := morerequire.RequireNewTempDir(t)
	defer removeHomeDir()

	t.Run("empty on missing versions dir", func(t *testing.T) {
		rows, err := getInstalledVersions(homeDir)
		require.NoError(t, err) // skips instead of crashing
		require.Empty(t, rows)
	})

	versionsDir := filepath.Join(homeDir, "versions")
	require.NoError(t, os.WriteFile(versionsDir, []byte{}, 0700))

	t.Run("error on file at versions dir", func(t *testing.T) {
		_, err := getInstalledVersions(homeDir)
		require.Error(t, err)
	})

	require.NoError(t, os.Remove(versionsDir))
	require.NoError(t, os.Mkdir(versionsDir, 0700))

	t.Run("empty on empty versions dir", func(t *testing.T) {
		rows, err := getInstalledVersions(homeDir)
		require.NoError(t, err) // skips instead of crashing
		require.Empty(t, rows)
	})

	oneOneTwo := filepath.Join(versionsDir, "1.1.2")
	require.NoError(t, os.MkdirAll(oneOneTwo, 0700))
	morerequire.RequireSetMtime(t, oneOneTwo, "2020-12-31")

	t.Run("release date from mtime", func(t *testing.T) {
		rows, err := getInstalledVersions(homeDir)
		require.NoError(t, err) // skips instead of crashing
		require.EqualValues(t, rows, []versionReleaseDate{{"1.1.2", "2020-12-31"}})
	})

	oneTwoOne := filepath.Join(versionsDir, "1.2.1")
	require.NoError(t, os.WriteFile(oneTwoOne, []byte{}, 0700)) // notice a file not a directory!
	morerequire.RequireSetMtime(t, oneTwoOne, "2020-12-30")

	t.Run("skips file where version should be", func(t *testing.T) {
		rows, err := getInstalledVersions(homeDir)
		require.NoError(t, err) // skips instead of crashing
		require.EqualValues(t, rows, []versionReleaseDate{{"1.1.2", "2020-12-31"}})
	})
}

func TestAddVersions(t *testing.T) {
	goodVersions := map[string]version.EnvoyVersion{
		"1.14.7": {
			ReleaseDate: "2021-04-15",
			Tarballs: map[string]string{
				"darwin/amd64": "https://getenvoy.io/versions/1.14.7/envoy-1.14.7-darwin-x86_64.tar.gz",
				"linux/amd64":  "https://getenvoy.io/versions/1.14.7/envoy-1.14.7-linux-x86_64.tar.gz",
			},
		},
		"1.17.3": {
			ReleaseDate: "2021-05-11",
			Tarballs: map[string]string{
				"linux/amd64": "https://getenvoy.io/versions/1.17.3/envoy-1.17.3-linux-x86_64.tar.gz",
			},
		},
		"1.18.3": {
			ReleaseDate: "2021-05-11",
			Tarballs: map[string]string{
				"darwin/amd64": "https://getenvoy.io/versions/1.18.3/envoy-1.18.3-darwin-x86_64.tar.gz",
				"linux/amd64":  "https://getenvoy.io/versions/1.18.3/envoy-1.18.3-linux-x86_64.tar.gz",
			},
		},
	}

	tests := []struct {
		name     string
		existing []versionReleaseDate
		update   map[string]version.EnvoyVersion
		platform string
		expected []versionReleaseDate
	}{
		{
			name:     "darwin",
			existing: []versionReleaseDate{},
			update:   goodVersions,
			platform: "darwin/amd64",
			expected: []versionReleaseDate{{"1.14.7", "2021-04-15"}, {"1.18.3", "2021-05-11"}},
		},
		{
			name:     "linux",
			platform: "linux/amd64",
			existing: []versionReleaseDate{},
			update:   goodVersions,
			expected: []versionReleaseDate{{"1.14.7", "2021-04-15"}, {"1.17.3", "2021-05-11"}, {"1.18.3", "2021-05-11"}},
		},
		{
			name:     "already exists",
			existing: []versionReleaseDate{{"1.14.7", "2020-01-01"}},
			update:   goodVersions,
			platform: "darwin/amd64",
			expected: []versionReleaseDate{{"1.14.7", "2020-01-01"}, {"1.18.3", "2021-05-11"}},
		},
		{
			name:     "unsupported OS",
			existing: []versionReleaseDate{},
			update:   goodVersions,
			platform: "windows/amd64",
			expected: []versionReleaseDate{},
		},
		{
			name:     "unsupported Arch",
			existing: []versionReleaseDate{},
			update:   goodVersions,
			platform: "linux/arm64",
			expected: []versionReleaseDate{},
		},
		{
			name:     "empty version list",
			existing: []versionReleaseDate{},
			platform: "darwin/amd64",
			expected: []versionReleaseDate{},
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, addAvailableVersions(&tc.existing, tc.update, tc.platform))
			require.ElementsMatch(t, tc.expected, tc.existing)
		})
	}
}

func TestAddVersions_Validates(t *testing.T) {
	tests := []struct {
		name   string
		update map[string]version.EnvoyVersion
	}{
		{
			name: "invalid releaseDate",
			update: map[string]version.EnvoyVersion{
				"1.14.7": {
					ReleaseDate: "ice cream",
					Tarballs: map[string]string{
						"darwin/amd64": "https://getenvoy.io/versions/1.14.7/envoy-1.14.7-darwin-x86_64.tar.gz",
					},
				},
			},
		},
		{
			name: "missing releaseDate",
			update: map[string]version.EnvoyVersion{
				"1.14.7": {
					Tarballs: map[string]string{
						"darwin/amd64": "https://getenvoy.io/versions/1.14.7/envoy-1.14.7-darwin-x86_64.tar.gz",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			err := addAvailableVersions(&[]versionReleaseDate{}, tc.update, "darwin/amd64")
			require.Error(t, err)
			require.Contains(t, err.Error(), `invalid releaseDate of version "1.14.7" for platform "darwin/amd64":`)
		})
	}
}
