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

package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestEnsureEnvoyVersion(t *testing.T) {
	o := &globals.GlobalOpts{HomeDir: t.TempDir()}
	require.NoError(t, os.WriteFile(filepath.Join(o.HomeDir, "version"), []byte(version.LastKnownEnvoy.String()), 0o600))

	err := EnsureEnvoyVersion(context.Background(), o)
	require.NoError(t, err)
	require.Equal(t, version.LastKnownEnvoy, o.EnvoyVersion)
}

func TestEnsureEnvoyVersion_ErrorIsAValidationError(t *testing.T) {
	o := &globals.GlobalOpts{HomeDir: t.TempDir()}
	require.NoError(t, os.WriteFile(filepath.Join(o.HomeDir, "version"), []byte("a.b.c"), 0o600))

	expectedErr := fmt.Sprintf(`invalid version in "$FUNC_E_HOME/version": "a.b.c" should look like %q or %q`, version.LastKnownEnvoy, version.LastKnownEnvoyMinor)
	err := EnsureEnvoyVersion(context.Background(), o)
	require.EqualError(t, err, moreos.ReplacePathSeparator(expectedErr))
}

func TestSetEnvoyVersion_ReadsExistingPatchVersion(t *testing.T) {
	o := &globals.GlobalOpts{HomeDir: t.TempDir()}

	require.NoError(t, os.WriteFile(filepath.Join(o.HomeDir, "version"), []byte("1.18.13"), 0o600))

	err := setEnvoyVersion(context.Background(), o)
	require.NoError(t, err)
	require.Equal(t, version.PatchVersion("1.18.13"), o.EnvoyVersion)
}

func TestSetEnvoyVersion_LooksUpLatestPatchForExistingMinorVersion(t *testing.T) {
	o := &globals.GlobalOpts{
		GetEnvoyVersions: func(context.Context) (*version.ReleaseVersions, error) {
			return &version.ReleaseVersions{Versions: map[version.PatchVersion]version.Release{
				"1.18.12": {Tarballs: map[version.Platform]version.TarballURL{globals.DefaultPlatform: ""}},
				"1.18.13": {Tarballs: map[version.Platform]version.TarballURL{globals.DefaultPlatform: ""}},
			}}, nil
		},
		HomeDir:  t.TempDir(),
		Out:      new(bytes.Buffer), // we expect logging
		Platform: globals.DefaultPlatform,
	}

	require.NoError(t, os.WriteFile(filepath.Join(o.HomeDir, "version"), []byte("1.18"), 0o600))

	err := setEnvoyVersion(context.Background(), o)
	require.NoError(t, err)
	require.Equal(t, version.PatchVersion("1.18.13"), o.EnvoyVersion)

	// We notified the user about the remote lookup
	require.Contains(t, o.Out.(*bytes.Buffer).String(), moreos.Sprintf("looking up the latest patch for Envoy version 1.18\n"))
}

func TestSetEnvoyVersion_ErrorReadingExistingVersion(t *testing.T) {
	o := &globals.GlobalOpts{HomeDir: t.TempDir()}
	require.NoError(t, os.WriteFile(filepath.Join(o.HomeDir, "version"), []byte("a.b.c"), 0o600))

	expectedErr := fmt.Sprintf(`invalid version in "$FUNC_E_HOME/version": "a.b.c" should look like %q or %q`, version.LastKnownEnvoy, version.LastKnownEnvoyMinor)
	err := setEnvoyVersion(context.Background(), o)
	require.EqualError(t, err, moreos.ReplacePathSeparator(expectedErr))
}

func TestSetEnvoyVersion_UsesLatestVersionOnInitialRun(t *testing.T) {
	o := &globals.GlobalOpts{
		GetEnvoyVersions: func(context.Context) (*version.ReleaseVersions, error) {
			return &version.ReleaseVersions{Versions: map[version.PatchVersion]version.Release{
				"1.19.2": {Tarballs: map[version.Platform]version.TarballURL{globals.DefaultPlatform: ""}},
				"1.18.3": {Tarballs: map[version.Platform]version.TarballURL{globals.DefaultPlatform: ""}},
				"1.20.4": {Tarballs: map[version.Platform]version.TarballURL{"solaris/sparc64": ""}},
			}}, nil
		},
		HomeDir:  t.TempDir(),
		Out:      new(bytes.Buffer), // we expect logging
		Platform: globals.DefaultPlatform,
	}

	err := setEnvoyVersion(context.Background(), o)
	require.NoError(t, err)

	// The highest version for this platform was set
	require.Equal(t, version.PatchVersion("1.19.2"), o.EnvoyVersion)

	// We notified the user about the remote lookup
	require.Contains(t, o.Out.(*bytes.Buffer).String(), moreos.Sprintf("looking up the latest Envoy version\n"))

	// We persisted the minor component for next run!
	writtenVersion, err := os.ReadFile(filepath.Join(o.HomeDir, "version"))
	require.NoError(t, err)
	require.Equal(t, o.EnvoyVersion.ToMinor().String(), string(writtenVersion))
}

func TestSetEnvoyVersion_NotFound(t *testing.T) {
	o := &globals.GlobalOpts{
		GetEnvoyVersions: func(context.Context) (*version.ReleaseVersions, error) {
			return &version.ReleaseVersions{Versions: map[version.PatchVersion]version.Release{
				"1.18.14": {Tarballs: map[version.Platform]version.TarballURL{"solaris/sparc64": ""}},
			}}, nil
		},
		EnvoyVersionsURL: "fake URL", // for logging
		HomeDir:          t.TempDir(),
		Out:              new(bytes.Buffer), // we expect logging
		Platform:         globals.DefaultPlatform,
	}

	err := setEnvoyVersion(context.Background(), o)
	expectedErr := fmt.Sprintf("fake URL does not contain an Envoy release for platform %s", o.Platform)
	require.EqualError(t, err, expectedErr)

	// We notified the user about the remote lookup
	require.Contains(t, o.Out.(*bytes.Buffer).String(), moreos.Sprintf("looking up the latest Envoy version\n"))
}

func TestSetEnvoyVersion_ErrorLookingUpLatestVersionOnInitialRun(t *testing.T) {
	o := &globals.GlobalOpts{
		GetEnvoyVersions: func(context.Context) (*version.ReleaseVersions, error) {
			return nil, errors.New("file not found")
		},
		EnvoyVersionsURL: "fake URL", // for logging
		HomeDir:          t.TempDir(),
		Out:              new(bytes.Buffer), // we expect logging
	}

	err := setEnvoyVersion(context.Background(), o)
	require.EqualError(t, err, "couldn't lookup the latest Envoy version from fake URL: file not found")

	// We notified the user about the remote lookup
	require.Contains(t, o.Out.(*bytes.Buffer).String(), moreos.Sprintf("looking up the latest Envoy version\n"))

	// No version file was written
	require.NoFileExists(t, filepath.Join(o.HomeDir, "version"))
}

func TestEnsurePatchVersion(t *testing.T) {
	versions := map[version.PatchVersion]version.Release{
		version.PatchVersion("1.18.3"):       {Tarballs: map[version.Platform]version.TarballURL{globals.DefaultPlatform: ""}},
		version.PatchVersion("1.18.13"):      {Tarballs: map[version.Platform]version.TarballURL{globals.DefaultPlatform: ""}},
		version.PatchVersion("1.18.14"):      {Tarballs: map[version.Platform]version.TarballURL{"solaris/sparc64": ""}},
		version.PatchVersion("1.18.4"):       {Tarballs: map[version.Platform]version.TarballURL{globals.DefaultPlatform: ""}},
		version.PatchVersion("1.18.4_debug"): {Tarballs: map[version.Platform]version.TarballURL{globals.DefaultPlatform: ""}},
	}

	o := &globals.GlobalOpts{
		GetEnvoyVersions: func(context.Context) (*version.ReleaseVersions, error) {
			return &version.ReleaseVersions{Versions: versions}, nil
		},
		HomeDir:  t.TempDir(),
		Out:      new(bytes.Buffer), // we expect logging
		Platform: globals.DefaultPlatform,
	}

	actual, err := EnsurePatchVersion(context.Background(), o, version.MinorVersion("1.18"))
	require.NoError(t, err)
	require.Equal(t, version.PatchVersion("1.18.13"), actual)

	// We notified the user about the remote lookup
	require.Contains(t, o.Out.(*bytes.Buffer).String(), moreos.Sprintf("looking up the latest patch for Envoy version 1.18\n"))
}

func TestEnsurePatchVersion_NotFound(t *testing.T) {
	versions := map[version.PatchVersion]version.Release{
		version.PatchVersion("1.18.14"):   {Tarballs: map[version.Platform]version.TarballURL{"solaris/sparc64": ""}},
		version.PatchVersion("1.20.0"):    {Tarballs: map[version.Platform]version.TarballURL{globals.DefaultPlatform: ""}},
		version.PatchVersion("1.1_debug"): {Tarballs: map[version.Platform]version.TarballURL{globals.DefaultPlatform: ""}},
	}

	o := &globals.GlobalOpts{
		GetEnvoyVersions: func(context.Context) (*version.ReleaseVersions, error) {
			return &version.ReleaseVersions{Versions: versions}, nil
		},
		EnvoyVersionsURL: "fake URL", // for logging
		HomeDir:          t.TempDir(),
		Out:              new(bytes.Buffer), // we expect logging
		Platform:         globals.DefaultPlatform,
	}

	_, err := EnsurePatchVersion(context.Background(), o, version.MinorVersion("1.18"))
	expectedErr := fmt.Sprintf("fake URL does not contain an Envoy release for version 1.18 on platform %s", o.Platform)
	require.EqualError(t, err, expectedErr)

	// We notified the user about the remote lookup
	require.Contains(t, o.Out.(*bytes.Buffer).String(), moreos.Sprintf("looking up the latest patch for Envoy version 1.18\n"))
}

func TestEnsurePatchVersion_NoOpWhenAlreadyAPatchVersion(t *testing.T) {
	expected := version.PatchVersion("1.19.1")
	actual, err := EnsurePatchVersion(context.Background(), &globals.GlobalOpts{}, expected)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestEnsurePatchVersion_FallbackSuccess(t *testing.T) {
	tests := []struct {
		name             string
		getEnvoyVersions version.GetReleaseVersions
	}{
		{
			"error on lookup",
			func(context.Context) (*version.ReleaseVersions, error) {
				return nil, errors.New("file not found")
			},
		},
		{
			"no versions",
			func(context.Context) (*version.ReleaseVersions, error) {
				return &version.ReleaseVersions{}, nil
			},
		},
		{
			"no versions for this platform",
			func(context.Context) (*version.ReleaseVersions, error) {
				return &version.ReleaseVersions{
					Versions: map[version.PatchVersion]version.Release{
						"1.18.14": {Tarballs: map[version.Platform]version.TarballURL{"solaris/sparc64": ""}},
					},
				}, nil
			},
		},
	}

	for _, tt := range tests {
		tc := tt // pin! see https://github.com/kyoh86/scopelint for why
		t.Run(tc.name, func(t *testing.T) {
			o := &globals.GlobalOpts{
				GetEnvoyVersions: tc.getEnvoyVersions,
				HomeDir:          t.TempDir(),
				Out:              new(bytes.Buffer), // we expect logging
			}

			lastKnownEnvoyDir := filepath.Join(o.HomeDir, "versions", "1.18.14")
			require.NoError(t, os.MkdirAll(lastKnownEnvoyDir, 0o700))

			// Ensure that when we ask for a minor, the latest version is returned from the filesystem
			actual, err := EnsurePatchVersion(context.Background(), o, version.MinorVersion("1.18"))
			require.NoError(t, err)
			require.Equal(t, version.PatchVersion("1.18.14"), actual)

			// We notified the user about the remote lookup
			require.Contains(t, o.Out.(*bytes.Buffer).String(), moreos.Sprintf("looking up the latest patch for Envoy version 1.18\n"))
		})
	}
}

func TestEnsurePatchVersion_FallbackFailure(t *testing.T) {
	o := &globals.GlobalOpts{
		GetEnvoyVersions: func(context.Context) (*version.ReleaseVersions, error) {
			return nil, errors.New("file not found")
		},
		HomeDir: t.TempDir(),
		Out:     new(bytes.Buffer), // we expect logging
	}

	// Since we have nothing local to fall back to, we should raise the remote error
	_, err := EnsurePatchVersion(context.Background(), o, version.MinorVersion("1.18"))
	require.EqualError(t, err, "file not found")

	// We notified the user about the remote lookup
	require.Contains(t, o.Out.(*bytes.Buffer).String(), moreos.Sprintf("looking up the latest patch for Envoy version 1.18\n"))
}

func TestVersionsForPlatform(t *testing.T) {
	type testCase struct {
		name     string
		versions map[version.PatchVersion]version.Release
		expected []version.PatchVersion
	}
	tests := []testCase{
		{
			name:     "empty",
			versions: map[version.PatchVersion]version.Release{},
		},
		{
			name: "skips other platform",
			versions: map[version.PatchVersion]version.Release{
				version.PatchVersion("1.18.3"):       {Tarballs: map[version.Platform]version.TarballURL{globals.DefaultPlatform: ""}},
				version.PatchVersion("1.18.13"):      {Tarballs: map[version.Platform]version.TarballURL{globals.DefaultPlatform: ""}},
				version.PatchVersion("1.18.14"):      {Tarballs: map[version.Platform]version.TarballURL{"solaris/sparc64": ""}},
				version.PatchVersion("1.18.4"):       {Tarballs: map[version.Platform]version.TarballURL{globals.DefaultPlatform: ""}},
				version.PatchVersion("1.18.4_debug"): {Tarballs: map[version.Platform]version.TarballURL{globals.DefaultPlatform: ""}},
			},
			expected: []version.PatchVersion{"1.18.3", "1.18.13", "1.18.4", "1.18.4_debug"},
		},
	}

	for _, tt := range tests {
		tc := tt // pin! see https://github.com/kyoh86/scopelint for why
		t.Run(tc.name, func(t *testing.T) {
			actual := versionsForPlatform(tc.versions, globals.DefaultPlatform)
			require.ElementsMatch(t, tc.expected, actual)
		})
	}
}
