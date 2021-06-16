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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/globals"
	"github.com/tetratelabs/getenvoy/internal/test/morerequire"
	"github.com/tetratelabs/getenvoy/internal/version"
)

func TestGetEnvoyVersions_NothingYet(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	c, stdout, stderr := newApp(o)
	err := c.Run([]string{"getenvoy", "versions"})

	require.NoError(t, err)
	require.Empty(t, stdout) // allows consistent parsing even when nothing yet installed
	require.Empty(t, stderr)
}

func TestGetEnvoyVersions_NoCurrentVersion(t *testing.T) {
	o, cleanup := setupTestVersions(t)
	defer cleanup()
	require.NoError(t, os.Remove(filepath.Join(o.HomeDir, "version")))

	c, stdout, stderr := newApp(o)
	err := c.Run([]string{"getenvoy", "versions"})

	require.NoError(t, err)
	require.Equal(t, `  1.2.2 2021-01-31
  1.1.2 2021-01-31
  1.2.1 2021-01-30
`, stdout.String())
	require.Empty(t, stderr)
}

// TestGetEnvoyVersions_CurrentVersion tests depend on prior state, so execute sequentially. This doesn't use a matrix
// to improve readability
func TestGetEnvoyVersions_CurrentVersion(t *testing.T) {
	o, cleanup := setupTestVersions(t)
	defer cleanup()

	t.Run("no current version", func(t *testing.T) {
		require.NoError(t, os.Remove(filepath.Join(o.HomeDir, "version")))

		c, stdout, _ := newApp(o)
		require.NoError(t, c.Run([]string{"getenvoy", "versions"}))
		require.Equal(t, `  1.2.2 2021-01-31
  1.1.2 2021-01-31
  1.2.1 2021-01-30
`, stdout.String())
	})

	t.Run("set by $GETENVOY_HOME/version", func(t *testing.T) {
		require.NoError(t, os.WriteFile(filepath.Join(o.HomeDir, "version"), []byte("1.1.2"), 0600))

		c, stdout, _ := newApp(o)
		require.NoError(t, c.Run([]string{"getenvoy", "versions"}))
		require.Equal(t, `  1.2.2 2021-01-31
* 1.1.2 2021-01-31 (set by $GETENVOY_HOME/version)
  1.2.1 2021-01-30
`, stdout.String())
	})

	t.Run("set by $PWD/.envoy-version", func(t *testing.T) {
		revertTempWd := morerequire.RequireChdirIntoTemp(t)
		defer revertTempWd()
		require.NoError(t, os.WriteFile(".envoy-version", []byte("1.2.2"), 0600))

		c, stdout, _ := newApp(o)
		require.NoError(t, c.Run([]string{"getenvoy", "versions"}))
		require.Equal(t, `* 1.2.2 2021-01-31 (set by $PWD/.envoy-version)
  1.1.2 2021-01-31
  1.2.1 2021-01-30
`, stdout.String())
	})

	t.Run("set by $ENVOY_VERSION", func(t *testing.T) {
		revertEnv := morerequire.RequireSetenv(t, "ENVOY_VERSION", "1.2.1")
		defer revertEnv()

		c, stdout, _ := newApp(o)
		require.NoError(t, c.Run([]string{"getenvoy", "versions"}))
		require.Equal(t, `  1.2.2 2021-01-31
  1.1.2 2021-01-31
* 1.2.1 2021-01-30 (set by $ENVOY_VERSION)
`, stdout.String())
	})
}

func TestGetEnvoyVersions_Sorted(t *testing.T) {
	o, cleanup := setupTestVersions(t)
	defer cleanup()

	c, stdout, stderr := newApp(o)
	err := c.Run([]string{"getenvoy", "versions"})

	require.NoError(t, err)
	require.Equal(t, `  1.2.2 2021-01-31
  1.1.2 2021-01-31
* 1.2.1 2021-01-30 (set by $GETENVOY_HOME/version)
`, stdout.String())
	require.Empty(t, stderr)
}

func TestGetEnvoyVersions_All_OnlyRemote(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	c, stdout, stderr := newApp(o)
	err := c.Run([]string{"getenvoy", "versions", "-a"})

	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("  %s 2020-12-31\n", version.LastKnownEnvoy), stdout.String())
	require.Empty(t, stderr)
}

func TestGetEnvoyVersions_All_RemoteIsCurrent(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	versionDir := filepath.Join(o.HomeDir, "versions", string(version.LastKnownEnvoy))
	require.NoError(t, os.MkdirAll(versionDir, 0700))
	morerequire.RequireSetMtime(t, versionDir, "2020-12-31")
	require.NoError(t, os.WriteFile(filepath.Join(o.HomeDir, "version"), []byte(version.LastKnownEnvoy), 0600))

	expected := fmt.Sprintf("* %s 2020-12-31 (set by $GETENVOY_HOME/version)\n", version.LastKnownEnvoy)

	c, stdout, stderr := newApp(o)
	err := c.Run([]string{"getenvoy", "versions", "-a"})

	require.NoError(t, err)
	require.Equal(t, expected, stdout.String())
	require.Empty(t, stderr)
}

func TestGetEnvoyVersions_All_Mixed(t *testing.T) {
	o, cleanup := setupTestVersions(t)
	defer cleanup()

	c, stdout, stderr := newApp(o)
	err := c.Run([]string{"getenvoy", "versions", "-a"})

	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(`  1.2.2 2021-01-31
  1.1.2 2021-01-31
* 1.2.1 2021-01-30 (set by $GETENVOY_HOME/version)
  %s 2020-12-31
`, version.LastKnownEnvoy), stdout.String())
	require.Empty(t, stderr)
}

func setupTestVersions(t *testing.T) (o *globals.GlobalOpts, cleanup func()) {
	o, cleanup = setupTest(t)

	oneOneTwo := filepath.Join(o.HomeDir, "versions", "1.1.2")
	require.NoError(t, os.MkdirAll(oneOneTwo, 0700))
	morerequire.RequireSetMtime(t, oneOneTwo, "2021-01-31")

	// Set the middle version current
	oneTwoOne := filepath.Join(o.HomeDir, "versions", "1.2.1")
	require.NoError(t, os.MkdirAll(oneTwoOne, 0700))
	morerequire.RequireSetMtime(t, oneTwoOne, "2021-01-30")
	require.NoError(t, os.WriteFile(filepath.Join(o.HomeDir, "version"), []byte("1.2.1"), 0600))

	oneTwoTwo := filepath.Join(o.HomeDir, "versions", "1.2.2")
	require.NoError(t, os.MkdirAll(oneTwoTwo, 0700))
	morerequire.RequireSetMtime(t, oneTwoTwo, "2021-01-31")
	return
}
