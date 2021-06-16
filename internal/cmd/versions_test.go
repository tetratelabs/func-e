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

func TestGetInstalledVersions_ErrorsWhenFileIsInVersionsDir(t *testing.T) {
	homeDir, removeHomeDir := morerequire.RequireNewTempDir(t)
	defer removeHomeDir()

	versionsDir := filepath.Join(homeDir, "versions")
	require.NoError(t, os.WriteFile(versionsDir, []byte{}, 0700))

	_, err := getInstalledVersions(homeDir)
	require.Error(t, err)
}

func TestGetInstalledVersions_MissingOrEmptyVersionsDir(t *testing.T) {
	homeDir, removeHomeDir := morerequire.RequireNewTempDir(t)
	defer removeHomeDir()

	rows, err := getInstalledVersions(homeDir)
	require.NoError(t, err) // ensures we don't error just because nothing is installed yet.
	require.Empty(t, rows)

	// Now, create the versions directory but don't add one
	versionsDir := filepath.Join(homeDir, "versions")
	require.NoError(t, os.Mkdir(versionsDir, 0700))

	rows, err = getInstalledVersions(homeDir)
	require.NoError(t, err)
	require.Empty(t, rows)
}

func TestGetInstalledVersions_ReleaseDateFromMtime(t *testing.T) {
	homeDir, removeHomeDir := morerequire.RequireNewTempDir(t)
	defer removeHomeDir()

	oneOneTwo := filepath.Join(homeDir, "versions", "1.1.2")
	require.NoError(t, os.MkdirAll(oneOneTwo, 0700))
	morerequire.RequireSetMtime(t, oneOneTwo, "2020-12-31")

	rows, err := getInstalledVersions(homeDir)
	require.NoError(t, err)
	require.EqualValues(t, rows, []versionReleaseDate{{"1.1.2", "2020-12-31"}})
}

func TestGetInstalledVersions_SkipsFileInVersionsDir(t *testing.T) {
	homeDir, removeHomeDir := morerequire.RequireNewTempDir(t)
	defer removeHomeDir()

	// make the versions directory
	versionsDir := filepath.Join(homeDir, "versions")
	require.NoError(t, os.Mkdir(versionsDir, 0700))

	// create a file which looks like a version
	oneOneTwo := filepath.Join(versionsDir, "1.1.2")
	require.NoError(t, os.WriteFile(oneOneTwo, []byte{}, 0700))
	morerequire.RequireSetMtime(t, oneOneTwo, "2020-12-31")

	// ensure there are no versions in the output
	rows, err := getInstalledVersions(homeDir)
	require.NoError(t, err)
	require.Empty(t, rows)
}

func TestAddAvailableVersions(t *testing.T) {
	goodVersions := map[version.Version]version.Release{
		"1.14.7": {
			ReleaseDate: "2021-04-15",
			Tarballs: map[version.Platform]version.TarballURL{
				"darwin/amd64": "https://getenvoy.io/versions/1.14.7/envoy-1.14.7-darwin-x86_64.tar.gz",
				"linux/amd64":  "https://getenvoy.io/versions/1.14.7/envoy-1.14.7-linux-x86_64.tar.gz",
			},
		},
		"1.17.3": {
			ReleaseDate: "2021-05-11",
			Tarballs: map[version.Platform]version.TarballURL{
				"linux/amd64": "https://getenvoy.io/versions/1.17.3/envoy-1.17.3-linux-x86_64.tar.gz",
			},
		},
		"1.18.3": {
			ReleaseDate: "2021-05-11",
			Tarballs: map[version.Platform]version.TarballURL{
				"darwin/amd64": "https://getenvoy.io/versions/1.18.3/envoy-1.18.3-darwin-x86_64.tar.gz",
				"linux/amd64":  "https://getenvoy.io/versions/1.18.3/envoy-1.18.3-linux-x86_64.tar.gz",
			},
		},
	}

	tests := []struct {
		name     string
		existing []versionReleaseDate
		update   map[version.Version]version.Release
		platform version.Platform
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

func TestAddAvailableVersions_Validates(t *testing.T) {
	tests := []struct {
		name   string
		update map[version.Version]version.Release
	}{
		{
			name: "invalid releaseDate",
			update: map[version.Version]version.Release{
				"1.14.7": {
					ReleaseDate: "ice cream",
					Tarballs: map[version.Platform]version.TarballURL{
						"darwin/amd64": "https://getenvoy.io/versions/1.14.7/envoy-1.14.7-darwin-x86_64.tar.gz",
					},
				},
			},
		},
		{
			name: "missing releaseDate",
			update: map[version.Version]version.Release{
				"1.14.7": {
					Tarballs: map[version.Platform]version.TarballURL{
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
