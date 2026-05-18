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
	"time"

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
			err := c.Run(t.Context(), []string{"func-e", "use", tc.version})

			// Verify the command failed with the expected error
			require.EqualError(t, err, tc.expectedErr)
			// func-e handles logging of errors, so we expect nothing in stdout or stderr
			require.Empty(t, stdout)
			require.Empty(t, stderr)
		})
	}
}

func TestFuncEUse(t *testing.T) {
	installDev := func(t *testing.T, o *globals.GlobalOpts) {
		t.Helper()
		c, _, _ := newApp(o)
		require.NoError(t, c.Run(t.Context(), []string{"func-e", "use", "dev"}))
		o.Out = new(bytes.Buffer)
	}
	installMinor := func(t *testing.T, o *globals.GlobalOpts, minor string) {
		t.Helper()
		c, _, _ := newApp(o)
		require.NoError(t, c.Run(t.Context(), []string{"func-e", "use", minor}))
		o.Out = new(bytes.Buffer)
	}

	tests := []struct {
		name         string
		setup        func(t *testing.T, o *globals.GlobalOpts)
		version      string
		stdout       string
		savedVersion string
	}{
		{name: "patch", version: version.LastKnownEnvoy.String(), stdout: "downloading", savedVersion: version.LastKnownEnvoy.String()},
		{name: "dev", version: "dev", stdout: "downloading", savedVersion: "dev"},
		{name: "dev already downloaded", setup: installDev, version: "dev", stdout: "already downloaded", savedVersion: "dev"},
		{name: "dev-latest up to date", setup: installDev, version: "dev-latest", savedVersion: "dev"},
		{name: "dev-latest out of date", setup: func(t *testing.T, o *globals.GlobalOpts) {
			t.Helper()
			installDev(t, o)
			stale := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
			require.NoError(t, os.Chtimes(filepath.Join(o.DataHome, "envoy-versions", "dev"), stale, stale))
		}, version: "dev-latest", stdout: "downloading", savedVersion: "dev"},
		{
			name: "minor upgradable",
			setup: func(t *testing.T, o *globals.GlobalOpts) {
				t.Helper()
				overrideAvailableVersions(t, o, []version.PatchVersion{"1.18.3"})
				installMinor(t, o, "1.18")
				overrideAvailableVersions(t, o, []version.PatchVersion{"1.18.3", "1.18.4"})
			},
			version: "1.18",
			stdout:  "downloading",
		},
		{
			name: "minor not upgraded",
			setup: func(t *testing.T, o *globals.GlobalOpts) {
				t.Helper()
				overrideAvailableVersions(t, o, []version.PatchVersion{"1.29.5"})
				installMinor(t, o, "1.29")
			},
			version: "1.29",
			stdout:  "already downloaded",
		},
		{
			name: "minor offline after install",
			setup: func(t *testing.T, o *globals.GlobalOpts) {
				t.Helper()
				overrideAvailableVersions(t, o, []version.PatchVersion{"1.29.3"})
				installMinor(t, o, "1.29")
				o.GetEnvoyVersions = func(_ context.Context) (*version.ReleaseVersions, error) {
					return nil, errors.New("offline")
				}
			},
			version: "1.29",
			stdout:  "already downloaded",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := setupTest(t)
			if tt.setup != nil {
				tt.setup(t, o)
			}
			c, stdout, _ := newApp(o)
			require.NoError(t, c.Run(t.Context(), []string{"func-e", "use", tt.version}))
			if tt.stdout != "" {
				require.Contains(t, stdout.String(), tt.stdout)
			}
			if tt.savedVersion != "" {
				f, err := os.ReadFile(filepath.Join(o.ConfigHome, "envoy-version"))
				require.NoError(t, err)
				require.Equal(t, tt.savedVersion, string(f))
			}
		})
	}
}

func overrideAvailableVersions(t *testing.T, o *globals.GlobalOpts, patches []version.PatchVersion) {
	t.Helper()
	evs, err := envoy.NewGetVersions(o.HTTPClient, o.EnvoyVersionsURL, o.UserAgent)(t.Context())
	require.NoError(t, err)
	var base version.Release
	for _, r := range evs.Versions {
		base = r
		break
	}
	versions := make(map[version.PatchVersion]version.Release, len(patches))
	for _, p := range patches {
		versions[p] = base
	}
	o.GetEnvoyVersions = func(_ context.Context) (*version.ReleaseVersions, error) {
		return &version.ReleaseVersions{Versions: versions, SHA256Sums: evs.SHA256Sums, Dev: evs.Dev}, nil
	}
}
