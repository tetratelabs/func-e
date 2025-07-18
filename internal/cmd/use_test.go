// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/envoy"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestFuncEUse_VersionValidates(t *testing.T) {
	o := setupTest(t)

	tests := []struct{ name, version, expectedErr string }{
		{
			name:        "version empty",
			expectedErr: "missing [version] argument",
		},
		{
			name:        "version invalid",
			version:     "a.b.c",
			expectedErr: fmt.Sprintf(`invalid [version] argument: "a.b.c" should look like %q or %q`, version.LastKnownEnvoy, version.LastKnownEnvoyMinor),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, stdout, stderr := newApp(o)
			err := c.Run([]string{"func-e", "use", tc.version})

			// Verify the command failed with the expected error
			require.EqualError(t, err, tc.expectedErr)
			// func-e handles logging of errors, so we expect nothing in stdout or stderr
			require.Empty(t, stdout)
			require.Empty(t, stderr)
		})
	}
}

func TestFuncEUse_InstallsAndWritesHomeVersion(t *testing.T) {
	o := setupTest(t)
	evs := o.EnvoyVersion.String()

	c, _, _ := newApp(o)
	require.NoError(t, c.Run([]string{"func-e", "use", evs}))

	// The binary was installed
	require.FileExists(t, filepath.Join(o.HomeDir, "versions", evs, "bin", "envoy"+""))

	// The current version was written
	f, err := os.ReadFile(filepath.Join(o.HomeDir, "version"))
	require.NoError(t, err)
	require.Equal(t, evs, string(f))
}

// TODO: everything from here down in this file needs to be rewritten
func TestFuncEUse_InstallMinorVersion(t *testing.T) {
	o := setupTest(t)

	type testCase struct {
		name           string
		firstVersions  availableVersions
		secondVersions availableVersions
		minorVersion   string
	}

	tests := []testCase{
		{
			name: "upgradable",
			firstVersions: availableVersions{
				latestPatch: "3",
				versions:    []version.PatchVersion{version.PatchVersion("1.18.3")},
			},
			secondVersions: availableVersions{
				latestPatch: "4",
				versions:    []version.PatchVersion{version.PatchVersion("1.18.3"), version.PatchVersion("1.18.4")},
			},
			minorVersion: "1.18",
		},
		{
			name: "not-upgraded",
			firstVersions: availableVersions{
				latestPatch: "3",
				versions:    []version.PatchVersion{version.PatchVersion("1.12.3")},
			},
			secondVersions: availableVersions{
				latestPatch: "3",
				versions:    []version.PatchVersion{version.PatchVersion("1.12.3")},
			},
			minorVersion: "1.12",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			o.GetEnvoyVersions, err = newFuncEVersionsTester(o, tc.firstVersions)
			require.NoError(t, err)

			c, _, _ := newApp(o)
			require.NoError(t, c.Run([]string{"func-e", "use", tc.minorVersion}))
			f, err := os.ReadFile(filepath.Join(o.HomeDir, "version"))
			require.NoError(t, err)
			require.Equal(t, tc.minorVersion, string(f))

			// Set o.EnvoyVersion to empty string so the logic for ensuring installed Envoy version works.
			o.EnvoyVersion = ""
			c, stdout, stderr := newApp(o)
			require.NoError(t, c.Run([]string{"func-e", "which"}))
			envoyPath := filepath.Join(o.HomeDir, "versions", tc.minorVersion+"."+tc.firstVersions.latestPatch, "bin", "envoy"+"")
			require.Equal(t, fmt.Sprintf("%s\n", envoyPath), stdout.String())
			require.Empty(t, stderr)

			// Update the map returned by Get.
			o.GetEnvoyVersions, err = newFuncEVersionsTester(o, tc.secondVersions)
			require.NoError(t, err)
			c, _, _ = newApp(o)
			require.NoError(t, c.Run([]string{"func-e", "use", tc.minorVersion}))
			f, err = os.ReadFile(filepath.Join(o.HomeDir, "version"))
			require.NoError(t, err)
			require.Equal(t, tc.minorVersion, string(f))

			// Set o.EnvoyVersion to empty string so the logic for ensuring installed Envoy version works.
			o.EnvoyVersion = ""
			c, stdout, stderr = newApp(o)
			require.NoError(t, c.Run([]string{"func-e", "which"}))
			envoyPath = filepath.Join(o.HomeDir, "versions", tc.minorVersion+"."+tc.secondVersions.latestPatch, "bin", "envoy"+"")
			require.Equal(t, fmt.Sprintf("%s\n", envoyPath), stdout.String())
			require.Empty(t, stderr)
		})
	}
}

func TestFuncEUse_InstallMinorVersionCheckLatestPatchFailed(t *testing.T) {
	o := setupTest(t)

	// The initial version to be installed.
	minorVersion := "1.12"
	latestPatch := "3"
	initial := availableVersions{
		latestPatch: latestPatch,
		versions:    []version.PatchVersion{version.PatchVersion(minorVersion + "." + latestPatch)},
	}

	var err error
	o.GetEnvoyVersions, err = newFuncEVersionsTester(o, initial)
	require.NoError(t, err)

	c, _, _ := newApp(o)
	require.NoError(t, c.Run([]string{"func-e", "use", minorVersion}))
	f, err := os.ReadFile(filepath.Join(o.HomeDir, "version"))
	require.NoError(t, err)
	require.Equal(t, minorVersion, string(f))

	o.EnvoyVersion = ""
	c, stdout, stderr := newApp(o)
	require.NoError(t, c.Run([]string{"func-e", "which"}))
	envoyPath := filepath.Join(o.HomeDir, "versions", minorVersion+"."+latestPatch, "bin", "envoy"+"")
	require.Equal(t, fmt.Sprintf("%s\n", envoyPath), stdout.String())
	require.Empty(t, stderr)

	o.GetEnvoyVersions = func(_ context.Context) (*version.ReleaseVersions, error) {
		return nil, errors.New("ice cream")
	}
	c, _, _ = newApp(o)
	require.NoError(t, c.Run([]string{"func-e", "use", minorVersion}))
	f, err = os.ReadFile(filepath.Join(o.HomeDir, "version"))
	require.NoError(t, err)
	require.Equal(t, minorVersion, string(f))

	o.EnvoyVersion = ""
	c, stdout, stderr = newApp(o)
	require.NoError(t, c.Run([]string{"func-e", "which"}))
	// The path points to the latest installed version.
	envoyPath = filepath.Join(o.HomeDir, "versions", minorVersion+"."+latestPatch, "bin", "envoy"+"")
	t.Log(stdout.String())
	require.Equal(t, fmt.Sprintf("%s\n", envoyPath), stdout.String())
	require.Empty(t, stderr)
}

type availableVersions struct {
	latestPatch string
	versions    []version.PatchVersion
}

func newFuncEVersionsTester(o *globals.GlobalOpts, av availableVersions) (version.GetReleaseVersions, error) {
	feV := envoy.NewGetVersions(o.EnvoyVersionsURL, o.Platform, o.Version)
	ev, err := feV(context.Background())
	if err != nil {
		return nil, err
	}
	// Copy versions releases from the setupTest and append more versions for testing.
	copied := ev
	var m version.Release
	for _, entry := range ev.Versions {
		m = entry
		break
	}
	for _, v := range av.versions {
		copied.Versions[v] = m
	}
	return func(_ context.Context) (*version.ReleaseVersions, error) {
		return copied, nil
	}, nil
}
