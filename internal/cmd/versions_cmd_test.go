// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/test/morerequire"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestFuncEVersions(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		setup    func(t *testing.T) *globals.GlobalOpts
		expected string
	}{
		{
			name:     "nothing installed",
			args:     []string{"func-e", "versions"},
			setup:    setupTest,
			expected: "",
		},
		{
			name: "no current version",
			args: []string{"func-e", "versions"},
			setup: func(t *testing.T) *globals.GlobalOpts {
				t.Helper()
				o := setupTestVersions(t)
				require.NoError(t, os.Remove(filepath.Join(o.ConfigHome, "envoy-version")))
				return o
			},
			expected: "  1.2.2 2021-01-31\n  1.1.2 2021-01-31\n  1.2.1 2021-01-30\n",
		},
		{
			name:     "sorted with current",
			args:     []string{"func-e", "versions"},
			setup:    setupTestVersions,
			expected: "  1.2.2 2021-01-31\n  1.1.2 2021-01-31\n* 1.2.1 2021-01-30 (set by $FUNC_E_CONFIG_HOME/envoy-version)\n",
		},
		{
			name: "current set by config",
			args: []string{"func-e", "versions"},
			setup: func(t *testing.T) *globals.GlobalOpts {
				t.Helper()
				o := setupTestVersions(t)
				require.NoError(t, os.WriteFile(filepath.Join(o.ConfigHome, "envoy-version"), []byte("1.1.2"), 0o600))
				return o
			},
			expected: "  1.2.2 2021-01-31\n* 1.1.2 2021-01-31 (set by $FUNC_E_CONFIG_HOME/envoy-version)\n  1.2.1 2021-01-30\n",
		},
		{
			name: "current set by PWD",
			args: []string{"func-e", "versions"},
			setup: func(t *testing.T) *globals.GlobalOpts {
				t.Helper()
				o := setupTestVersions(t)
				t.Chdir(t.TempDir())
				require.NoError(t, os.WriteFile(".envoy-version", []byte("1.2.2"), 0o600))
				return o
			},
			expected: "* 1.2.2 2021-01-31 (set by $PWD/.envoy-version)\n  1.1.2 2021-01-31\n  1.2.1 2021-01-30\n",
		},
		{
			name: "current set by ENVOY_VERSION",
			args: []string{"func-e", "versions"},
			setup: func(t *testing.T) *globals.GlobalOpts {
				t.Helper()
				o := setupTestVersions(t)
				t.Setenv("ENVOY_VERSION", "1.2.1")
				return o
			},
			expected: "  1.2.2 2021-01-31\n  1.1.2 2021-01-31\n* 1.2.1 2021-01-30 (set by $ENVOY_VERSION)\n",
		},
		{
			name:     "all only remote",
			args:     []string{"func-e", "versions", "-a"},
			setup:    setupTest,
			expected: fmt.Sprintf("  dev 2020-12-31 (92c6cb58)\n  %s 2020-12-31\n", version.LastKnownEnvoy),
		},
		{
			name: "all remote is current",
			args: []string{"func-e", "versions", "-a"},
			setup: func(t *testing.T) *globals.GlobalOpts {
				t.Helper()
				o := setupTest(t)
				v := version.LastKnownEnvoy.String()
				versionDir := filepath.Join(o.DataHome, "envoy-versions", v)
				require.NoError(t, os.MkdirAll(versionDir, 0o700))
				morerequire.RequireSetMtime(t, versionDir, "2020-12-31")
				require.NoError(t, os.WriteFile(filepath.Join(o.ConfigHome, "envoy-version"), []byte(v), 0o600))
				return o
			},
			expected: fmt.Sprintf("  dev 2020-12-31 (92c6cb58)\n* %s 2020-12-31 (set by $FUNC_E_CONFIG_HOME/envoy-version)\n", version.LastKnownEnvoy),
		},
		{
			name: "all no dev in JSON",
			args: []string{"func-e", "versions", "-a"},
			setup: func(t *testing.T) *globals.GlobalOpts {
				t.Helper()
				o := setupTest(t)
				o.GetEnvoyVersions = func(_ context.Context) (*version.ReleaseVersions, error) {
					return &version.ReleaseVersions{
						Versions: map[version.PatchVersion]version.Release{
							version.LastKnownEnvoy: {
								ReleaseDate: "2020-12-31",
								Tarballs: map[version.Platform]version.TarballURL{
									o.Platform: "https://example.com/envoy.tar.gz",
								},
							},
						},
						SHA256Sums: map[version.Tarball]version.SHA256Sum{},
					}, nil
				}
				return o
			},
			expected: fmt.Sprintf("  %s 2020-12-31\n", version.LastKnownEnvoy),
		},
		{
			name:  "all mixed local and remote",
			args:  []string{"func-e", "versions", "-a"},
			setup: setupTestVersions,
			expected: fmt.Sprintf("  dev 2020-12-31 (92c6cb58)\n  1.2.2 2021-01-31\n  1.1.2 2021-01-31\n* 1.2.1 2021-01-30 (set by $FUNC_E_CONFIG_HOME/envoy-version)\n  %s 2020-12-31\n",
				version.LastKnownEnvoy),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.setup(t)
			c, stdout, stderr := newApp(o)
			require.NoError(t, c.Run(t.Context(), tt.args))
			require.Equal(t, tt.expected, stdout.String())
			require.Empty(t, stderr)
		})
	}
}

func setupTestVersions(t *testing.T) (o *globals.GlobalOpts) {
	t.Helper()
	o = setupTest(t)

	oneOneTwo := filepath.Join(o.DataHome, "envoy-versions", "1.1.2")
	require.NoError(t, os.MkdirAll(oneOneTwo, 0o700))
	morerequire.RequireSetMtime(t, oneOneTwo, "2021-01-31")

	// Set the middle version current
	oneTwoOne := filepath.Join(o.DataHome, "envoy-versions", "1.2.1")
	require.NoError(t, os.MkdirAll(oneTwoOne, 0o700))
	morerequire.RequireSetMtime(t, oneTwoOne, "2021-01-30")
	require.NoError(t, os.WriteFile(filepath.Join(o.ConfigHome, "envoy-version"), []byte("1.2.1"), 0o600))

	oneTwoTwo := filepath.Join(o.DataHome, "envoy-versions", "1.2.2")
	require.NoError(t, os.MkdirAll(oneTwoTwo, 0o700))
	morerequire.RequireSetMtime(t, oneTwoTwo, "2021-01-31")
	return o
}
