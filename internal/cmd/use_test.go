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
	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestFuncEUse_VersionValidates(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

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
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

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
	o, cleanup := setupTest(t)
	evs := o.EnvoyVersion.String()
	defer cleanup()

	c, _, _ := newApp(o)
	require.NoError(t, c.Run([]string{"func-e", "use", evs}))

	// The binary was installed
	require.FileExists(t, filepath.Join(o.HomeDir, "versions", evs, "bin", "envoy"+moreos.Exe))

	// The current version was written
	f, err := os.ReadFile(filepath.Join(o.HomeDir, "version"))
	require.NoError(t, err)
	require.Equal(t, evs, string(f))
}

func TestFuncEUse_InstallMinorVersion(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var err error
			o.FuncEVersions, err = newFuncEVersionsTester(o, tc.firstVersions)
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
			envoyPath := filepath.Join(o.HomeDir, "versions", tc.minorVersion+"."+tc.firstVersions.latestPatch, "bin", "envoy"+moreos.Exe)
			require.Equal(t, moreos.Sprintf("%s\n", envoyPath), stdout.String())
			require.Empty(t, stderr)

			// Update the map returned by Get.
			o.FuncEVersions, err = newFuncEVersionsTester(o, tc.secondVersions)
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
			envoyPath = filepath.Join(o.HomeDir, "versions", tc.minorVersion+"."+tc.secondVersions.latestPatch, "bin", "envoy"+moreos.Exe)
			require.Equal(t, moreos.Sprintf("%s\n", envoyPath), stdout.String())
			require.Empty(t, stderr)
		})
	}
}

func TestFuncEUse_InstallMinorVersionCheckLatestPatchFailed(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	// The initial version to be installed.
	minorVersion := "1.12"
	latestPatch := "3"
	initial := availableVersions{
		latestPatch: latestPatch,
		versions:    []version.PatchVersion{version.PatchVersion(minorVersion + "." + latestPatch)},
	}

	var err error
	o.FuncEVersions, err = newFuncEVersionsTester(o, initial)
	require.NoError(t, err)

	c, _, _ := newApp(o)
	require.NoError(t, c.Run([]string{"func-e", "use", minorVersion}))
	f, err := os.ReadFile(filepath.Join(o.HomeDir, "version"))
	require.NoError(t, err)
	require.Equal(t, minorVersion, string(f))

	o.EnvoyVersion = ""
	c, stdout, stderr := newApp(o)
	require.NoError(t, c.Run([]string{"func-e", "which"}))
	envoyPath := filepath.Join(o.HomeDir, "versions", minorVersion+"."+latestPatch, "bin", "envoy"+moreos.Exe)
	require.Equal(t, moreos.Sprintf("%s\n", envoyPath), stdout.String())
	require.Empty(t, stderr)

	// Simulate failure in fetching Envoy release versions by initializing o.FuncEVersions with empty
	// available versions.
	o.FuncEVersions = &funcEVersionsTester{}
	c, _, _ = newApp(o)
	require.NoError(t, c.Run([]string{"func-e", "use", minorVersion}))
	f, err = os.ReadFile(filepath.Join(o.HomeDir, "version"))
	require.NoError(t, err)
	require.Equal(t, minorVersion, string(f))

	o.EnvoyVersion = ""
	c, stdout, stderr = newApp(o)
	require.NoError(t, c.Run([]string{"func-e", "which"}))
	// The path points to the latest installed version.
	envoyPath = filepath.Join(o.HomeDir, "versions", minorVersion+"."+latestPatch, "bin", "envoy"+moreos.Exe)
	t.Log(stdout.String())
	require.Equal(t, moreos.Sprintf("%s\n", envoyPath), stdout.String())
	require.Empty(t, stderr)
}

type availableVersions struct {
	latestPatch string
	versions    []version.PatchVersion
}

type funcEVersionsTester struct {
	ev version.ReleaseVersions
	av availableVersions
}

func newFuncEVersionsTester(o *globals.GlobalOpts, av availableVersions) (version.FuncEVersions, error) {
	feV := envoy.NewFuncEVersions(o.EnvoyVersionsURL, o.Platform, o.Version)
	ev, err := feV.Get(context.Background())
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
	return &funcEVersionsTester{ev: copied, av: av}, nil
}

func (f *funcEVersionsTester) Get(_ context.Context) (version.ReleaseVersions, error) {
	return f.ev, nil
}

func (f *funcEVersionsTester) FindLatestPatch(_ context.Context, minorVersion version.MinorVersion) (version.PatchVersion, error) {
	// When the input latest patch is empty, send error. This is useful for simulating FindLatestPatch
	// to return error.
	if f.av.latestPatch == "" {
		return "", errors.New("failed to find latest patch")
	}
	return version.PatchVersion(minorVersion.String() + "." + f.av.latestPatch), nil
}
