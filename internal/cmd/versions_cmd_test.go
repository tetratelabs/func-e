// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/test/morerequire"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestFuncEVersions_NothingYet(t *testing.T) {
	o := setupTest(t)

	c, stdout, stderr := newApp(o)
	err := c.Run([]string{"func-e", "versions"})

	require.NoError(t, err)
	require.Empty(t, stdout) // allows consistent parsing even when nothing yet installed
	require.Empty(t, stderr)
}

func TestFuncEVersions_NoCurrentVersion(t *testing.T) {
	o := setupTestVersions(t)
	require.NoError(t, os.Remove(filepath.Join(o.HomeDir, "version")))

	c, stdout, stderr := newApp(o)
	err := c.Run([]string{"func-e", "versions"})

	require.NoError(t, err)
	require.Equal(t, `  1.2.2 2021-01-31
  1.1.2 2021-01-31
  1.2.1 2021-01-30
`, stdout.String())
	require.Empty(t, stderr)
}

// TestFuncEVersions_CurrentVersion tests depend on prior state, so execute sequentially. This doesn't use a matrix
// to improve readability
func TestFuncEVersions_CurrentVersion(t *testing.T) {
	o := setupTestVersions(t)

	t.Run("no current version", func(t *testing.T) {
		require.NoError(t, os.Remove(filepath.Join(o.HomeDir, "version")))

		c, stdout, _ := newApp(o)
		require.NoError(t, c.Run([]string{"func-e", "versions"}))
		require.Equal(t, `  1.2.2 2021-01-31
  1.1.2 2021-01-31
  1.2.1 2021-01-30
`, stdout.String())
	})

	t.Run("set by $FUNC_E_HOME/version", func(t *testing.T) {
		require.NoError(t, os.WriteFile(filepath.Join(o.HomeDir, "version"), []byte("1.1.2"), 0o600))

		c, stdout, _ := newApp(o)
		require.NoError(t, c.Run([]string{"func-e", "versions"}))
		require.Equal(t, `  1.2.2 2021-01-31
* 1.1.2 2021-01-31 (set by $FUNC_E_HOME/version)
  1.2.1 2021-01-30
`, stdout.String())
	})

	t.Run("set by $PWD/.envoy-version", func(t *testing.T) {
		revertWd := morerequire.RequireChdir(t, t.TempDir())
		defer revertWd()
		require.NoError(t, os.WriteFile(".envoy-version", []byte("1.2.2"), 0o600))

		c, stdout, _ := newApp(o)
		require.NoError(t, c.Run([]string{"func-e", "versions"}))
		require.Equal(t, `* 1.2.2 2021-01-31 (set by $PWD/.envoy-version)
  1.1.2 2021-01-31
  1.2.1 2021-01-30
`, stdout.String())
	})

	t.Run("set by $ENVOY_VERSION", func(t *testing.T) {
		t.Setenv("ENVOY_VERSION", "1.2.1")

		c, stdout, _ := newApp(o)
		require.NoError(t, c.Run([]string{"func-e", "versions"}))
		require.Equal(t, `  1.2.2 2021-01-31
  1.1.2 2021-01-31
* 1.2.1 2021-01-30 (set by $ENVOY_VERSION)
`, stdout.String())
	})
}

func TestFuncEVersions_Sorted(t *testing.T) {
	o := setupTestVersions(t)

	c, stdout, stderr := newApp(o)
	err := c.Run([]string{"func-e", "versions"})

	require.NoError(t, err)
	require.Equal(t, `  1.2.2 2021-01-31
  1.1.2 2021-01-31
* 1.2.1 2021-01-30 (set by $FUNC_E_HOME/version)
`, stdout.String())
	require.Empty(t, stderr)
}

func TestFuncEVersions_All_OnlyRemote(t *testing.T) {
	o := setupTest(t)

	c, stdout, stderr := newApp(o)
	err := c.Run([]string{"func-e", "versions", "-a"})

	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("  %s 2020-12-31\n", version.LastKnownEnvoy), stdout.String())
	require.Empty(t, stderr)
}

func TestFuncEVersions_All_RemoteIsCurrent(t *testing.T) {
	o := setupTest(t)

	v := version.LastKnownEnvoy.String()
	versionDir := filepath.Join(o.HomeDir, "versions", v)
	require.NoError(t, os.MkdirAll(versionDir, 0o700))
	morerequire.RequireSetMtime(t, versionDir, "2020-12-31")
	require.NoError(t, os.WriteFile(filepath.Join(o.HomeDir, "version"), []byte(v), 0o600))

	expected := fmt.Sprintf("* %s 2020-12-31 (set by $FUNC_E_HOME/version)\n", v)

	c, stdout, stderr := newApp(o)
	err := c.Run([]string{"func-e", "versions", "-a"})

	require.NoError(t, err)
	require.Equal(t, expected, stdout.String())
	require.Empty(t, stderr)
}

func TestFuncEVersions_All_Mixed(t *testing.T) {
	o := setupTestVersions(t)

	c, stdout, stderr := newApp(o)
	err := c.Run([]string{"func-e", "versions", "-a"})

	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(`  1.2.2 2021-01-31
  1.1.2 2021-01-31
* 1.2.1 2021-01-30 (set by $FUNC_E_HOME/version)
  %s 2020-12-31
`, version.LastKnownEnvoy), stdout.String())
	require.Empty(t, stderr)
}

func setupTestVersions(t *testing.T) (o *globals.GlobalOpts) {
	o = setupTest(t)

	oneOneTwo := filepath.Join(o.HomeDir, "versions", "1.1.2")
	require.NoError(t, os.MkdirAll(oneOneTwo, 0o700))
	morerequire.RequireSetMtime(t, oneOneTwo, "2021-01-31")

	// Set the middle version current
	oneTwoOne := filepath.Join(o.HomeDir, "versions", "1.2.1")
	require.NoError(t, os.MkdirAll(oneTwoOne, 0o700))
	morerequire.RequireSetMtime(t, oneTwoOne, "2021-01-30")
	require.NoError(t, os.WriteFile(filepath.Join(o.HomeDir, "version"), []byte("1.2.1"), 0o600))

	oneTwoTwo := filepath.Join(o.HomeDir, "versions", "1.2.2")
	require.NoError(t, os.MkdirAll(oneTwoTwo, 0o700))
	morerequire.RequireSetMtime(t, oneTwoTwo, "2021-01-31")
	return
}
