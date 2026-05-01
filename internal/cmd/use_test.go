// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	rootcmd "github.com/tetratelabs/func-e/internal/cmd"
	"github.com/tetratelabs/func-e/internal/envoy"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestFuncEUse_VersionValidates(t *testing.T) {
	o := setupTest(t)

	// The empty case is enforced by Kong before dispatch, so we exercise the
	// underlying validator directly. The non-empty case still flows through DoMain.
	_, err := version.NewVersion("[version] argument", "")
	require.EqualError(t, err, "missing [version] argument")

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	err = rootcmd.DoMain(t.Context(), stdout, stderr, []string{"use", "a.b.c"}, o, "test")
	require.EqualError(t, err, fmt.Sprintf(`invalid [version] argument: "a.b.c" should look like %q or %q`, version.LastKnownEnvoy, version.LastKnownEnvoyMinor))
}

func TestFuncEUse_InstallsAndWritesHomeVersion(t *testing.T) {
	o := setupTest(t)
	evs := o.EnvoyVersion.String()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	require.NoError(t, rootcmd.DoMain(t.Context(), stdout, stderr, []string{"use", evs}, o, "test"))

	// The binary was installed
	require.FileExists(t, filepath.Join(o.DataHome, "envoy-versions", evs, "bin", "envoy"+""))

	// The current version was written
	f, err := os.ReadFile(filepath.Join(o.ConfigHome, "envoy-version"))
	require.NoError(t, err)
	require.Equal(t, evs, string(f))
}

// TODO: everything from here down in this file needs to be rewritten
func TestFuncEUse_InstallMinorVersion(t *testing.T) {
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
				latestPatch: "5",
				versions:    []version.PatchVersion{version.PatchVersion("1.29.5")},
			},
			secondVersions: availableVersions{
				latestPatch: "5",
				versions:    []version.PatchVersion{version.PatchVersion("1.29.5")},
			},
			minorVersion: "1.29",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := setupTest(t)
			var err error
			o.GetEnvoyVersions, err = newFuncEVersionsTester(t.Context(), o, tc.firstVersions)
			require.NoError(t, err)

			stdout := new(bytes.Buffer)
			stderr := new(bytes.Buffer)
			require.NoError(t, rootcmd.DoMain(t.Context(), stdout, stderr, []string{"use", tc.minorVersion}, o, "test"))
			f, err := os.ReadFile(filepath.Join(o.ConfigHome, "envoy-version"))
			require.NoError(t, err)
			require.Equal(t, tc.minorVersion, string(f))

			// Set o.EnvoyVersion to empty string so the logic for ensuring installed Envoy version works.
			o.EnvoyVersion = ""
			stdout.Reset()
			stderr.Reset()
			require.NoError(t, rootcmd.DoMain(t.Context(), stdout, stderr, []string{"which"}, o, "test"))
			envoyPath := filepath.Join(o.DataHome, "envoy-versions", tc.minorVersion+"."+tc.firstVersions.latestPatch, "bin", "envoy"+"")
			require.Equal(t, envoyPath+"\n", stdout.String())
			require.Empty(t, stderr.String())

			// Update the map returned by Get.
			o.GetEnvoyVersions, err = newFuncEVersionsTester(t.Context(), o, tc.secondVersions)
			require.NoError(t, err)
			stdout.Reset()
			stderr.Reset()
			require.NoError(t, rootcmd.DoMain(t.Context(), stdout, stderr, []string{"use", tc.minorVersion}, o, "test"))
			f, err = os.ReadFile(filepath.Join(o.ConfigHome, "envoy-version"))
			require.NoError(t, err)
			require.Equal(t, tc.minorVersion, string(f))

			// Set o.EnvoyVersion to empty string so the logic for ensuring installed Envoy version works.
			o.EnvoyVersion = ""
			stdout.Reset()
			stderr.Reset()
			require.NoError(t, rootcmd.DoMain(t.Context(), stdout, stderr, []string{"which"}, o, "test"))
			envoyPath = filepath.Join(o.DataHome, "envoy-versions", tc.minorVersion+"."+tc.secondVersions.latestPatch, "bin", "envoy"+"")
			require.Equal(t, envoyPath+"\n", stdout.String())
			require.Empty(t, stderr.String())
		})
	}
}

func TestFuncEUse_InstallMinorVersionCheckLatestPatchFailed(t *testing.T) {
	o := setupTest(t)

	// The initial version to be installed.
	minorVersion := "1.29"
	latestPatch := "3"
	initial := availableVersions{
		latestPatch: latestPatch,
		versions:    []version.PatchVersion{version.PatchVersion(minorVersion + "." + latestPatch)},
	}

	var err error
	o.GetEnvoyVersions, err = newFuncEVersionsTester(t.Context(), o, initial)
	require.NoError(t, err)

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	require.NoError(t, rootcmd.DoMain(t.Context(), stdout, stderr, []string{"use", minorVersion}, o, "test"))
	f, err := os.ReadFile(filepath.Join(o.ConfigHome, "envoy-version"))
	require.NoError(t, err)
	require.Equal(t, minorVersion, string(f))

	o.EnvoyVersion = ""
	stdout.Reset()
	stderr.Reset()
	require.NoError(t, rootcmd.DoMain(t.Context(), stdout, stderr, []string{"which"}, o, "test"))
	envoyPath := filepath.Join(o.DataHome, "envoy-versions", minorVersion+"."+latestPatch, "bin", "envoy"+"")
	require.Equal(t, envoyPath+"\n", stdout.String())
	require.Empty(t, stderr.String())

	o.GetEnvoyVersions = func(_ context.Context) (*version.ReleaseVersions, error) {
		return nil, errors.New("ice cream")
	}
	stdout.Reset()
	stderr.Reset()
	require.NoError(t, rootcmd.DoMain(t.Context(), stdout, stderr, []string{"use", minorVersion}, o, "test"))
	f, err = os.ReadFile(filepath.Join(o.ConfigHome, "envoy-version"))
	require.NoError(t, err)
	require.Equal(t, minorVersion, string(f))

	o.EnvoyVersion = ""
	stdout.Reset()
	stderr.Reset()
	require.NoError(t, rootcmd.DoMain(t.Context(), stdout, stderr, []string{"which"}, o, "test"))
	// The path points to the latest installed version.
	envoyPath = filepath.Join(o.DataHome, "envoy-versions", minorVersion+"."+latestPatch, "bin", "envoy"+"")
	t.Log(stdout.String())
	require.Equal(t, envoyPath+"\n", stdout.String())
	require.Empty(t, stderr.String())
}

type availableVersions struct {
	latestPatch string
	versions    []version.PatchVersion
}

func newFuncEVersionsTester(ctx context.Context, o *globals.GlobalOpts, av availableVersions) (version.GetReleaseVersions, error) {
	feV := envoy.NewGetVersions(o.HTTPClientFunc, o.EnvoyVersionsURL, o.UserAgent)
	ev, err := feV(ctx)
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
