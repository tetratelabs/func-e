// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/test/morerequire"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestGetInstalledVersions_ErrorsWhenFileIsInVersionsDir(t *testing.T) {
	homeDir := t.TempDir()

	versionsDir := filepath.Join(homeDir, "versions")
	require.NoError(t, os.WriteFile(versionsDir, []byte{}, 0o600))

	_, err := getInstalledVersions(homeDir)
	require.Error(t, err)
}

func TestGetInstalledVersions_MissingOrEmptyVersionsDir(t *testing.T) {
	homeDir := t.TempDir()

	rows, err := getInstalledVersions(homeDir)
	require.NoError(t, err) // ensures we don't error just because nothing is installed yet.
	require.Empty(t, rows)

	// Now, create the versions directory but don't add one
	versionsDir := filepath.Join(homeDir, "versions")
	require.NoError(t, os.Mkdir(versionsDir, 0o700))

	rows, err = getInstalledVersions(homeDir)
	require.NoError(t, err)
	require.Empty(t, rows)
}

func TestGetInstalledVersions_ReleaseDateFromMtime(t *testing.T) {
	homeDir := t.TempDir()

	oneOneTwo := filepath.Join(homeDir, "versions", "1.1.2")
	require.NoError(t, os.MkdirAll(oneOneTwo, 0o700))
	morerequire.RequireSetMtime(t, oneOneTwo, "2020-12-31")

	rows, err := getInstalledVersions(homeDir)
	require.NoError(t, err)
	require.Equal(t, []versionReleaseDate{{"1.1.2", "2020-12-31"}}, rows)
}

func TestGetInstalledVersions_SkipsFileInVersionsDir(t *testing.T) {
	homeDir := t.TempDir()

	// make the versions directory
	versionsDir := filepath.Join(homeDir, "versions")
	require.NoError(t, os.Mkdir(versionsDir, 0o700))

	// create a file which looks like a version
	oneOneTwo := filepath.Join(versionsDir, "1.1.2")
	require.NoError(t, os.WriteFile(oneOneTwo, []byte{}, 0o600))
	morerequire.RequireSetMtime(t, oneOneTwo, "2020-12-31")

	// ensure there are no versions in the output
	rows, err := getInstalledVersions(homeDir)
	require.NoError(t, err)
	require.Empty(t, rows)
}

func TestAddAvailableVersions(t *testing.T) {
	goodVersions := map[version.PatchVersion]version.Release{
		version.PatchVersion("1.14.7"): {
			ReleaseDate: "2021-04-15",
			Tarballs: map[version.Platform]version.TarballURL{
				"darwin/amd64": "https://func-e.io/versions/1.14.7/envoy-1.14.7-darwin-x86_64.tar.gz",
				"linux/amd64":  "https://func-e.io/versions/1.14.7/envoy-1.14.7-linux-x86_64.tar.gz",
			},
		},
		version.PatchVersion("1.17.3"): {
			ReleaseDate: "2021-05-11",
			Tarballs: map[version.Platform]version.TarballURL{
				"linux/amd64": "https://func-e.io/versions/1.17.3/envoy-1.17.3-linux-x86_64.tar.gz",
			},
		},
		version.PatchVersion("1.18.3"): {
			ReleaseDate: "2021-05-11",
			Tarballs: map[version.Platform]version.TarballURL{
				"darwin/amd64": "https://func-e.io/versions/1.18.3/envoy-1.18.3-darwin-x86_64.tar.gz",
				"linux/amd64":  "https://func-e.io/versions/1.18.3/envoy-1.18.3-linux-x86_64.tar.gz",
			},
		},
	}

	tests := []struct {
		name     string
		existing []versionReleaseDate
		update   map[version.PatchVersion]version.Release
		platform version.Platform
		expected []versionReleaseDate
	}{
		{
			name:     "darwin",
			existing: []versionReleaseDate{},
			update:   goodVersions,
			platform: "darwin/amd64",
			expected: []versionReleaseDate{{version.PatchVersion("1.14.7"), "2021-04-15"}, {version.PatchVersion("1.18.3"), "2021-05-11"}},
		},
		{
			name:     "linux",
			platform: "linux/amd64",
			existing: []versionReleaseDate{},
			update:   goodVersions,
			expected: []versionReleaseDate{{version.PatchVersion("1.14.7"), "2021-04-15"}, {version.PatchVersion("1.17.3"), "2021-05-11"}, {version.PatchVersion("1.18.3"), "2021-05-11"}},
		},
		{
			name:     "already exists",
			existing: []versionReleaseDate{{version.PatchVersion("1.14.7"), "2020-01-01"}},
			update:   goodVersions,
			platform: "darwin/amd64",
			expected: []versionReleaseDate{{version.PatchVersion("1.14.7"), "2020-01-01"}, {version.PatchVersion("1.18.3"), "2021-05-11"}},
		},
		{
			name:     "unsupported OS",
			existing: []versionReleaseDate{},
			update:   goodVersions,
			platform: "solaris/amd64",
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
		update map[version.PatchVersion]version.Release
	}{
		{
			name: "invalid releaseDate",
			update: map[version.PatchVersion]version.Release{
				version.PatchVersion("1.14.7"): {
					ReleaseDate: "ice cream",
					Tarballs: map[version.Platform]version.TarballURL{
						"darwin/amd64": "https://func-e.io/versions/1.14.7/envoy-1.14.7-darwin-x86_64.tar.gz",
					},
				},
			},
		},
		{
			name: "missing releaseDate",
			update: map[version.PatchVersion]version.Release{
				version.PatchVersion("1.14.7"): {
					Tarballs: map[version.Platform]version.TarballURL{
						"darwin/amd64": "https://func-e.io/versions/1.14.7/envoy-1.14.7-darwin-x86_64.tar.gz",
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
