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
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/version"
)

func TestGetEnvoyVersions_NothingYet(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	c, stdout, stderr := newApp(o)
	o.Out = stdout
	err := c.Run([]string{"getenvoy", "versions"})

	require.NoError(t, err)
	require.Equal(t, "No Envoy versions, yet\n", stdout.String())
	require.Empty(t, stderr)
}

// TestGetEnvoyVersions verifies output is sorted
func TestGetEnvoyVersions_Sorted(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	oneOneTwo := filepath.Join(o.HomeDir, "versions", "1.1.2")
	require.NoError(t, os.MkdirAll(oneOneTwo, 0700))
	requireSetMtime(t, oneOneTwo, "2020-12-31")

	oneTwoOne := filepath.Join(o.HomeDir, "versions", "1.2.1")
	require.NoError(t, os.MkdirAll(oneTwoOne, 0700))
	requireSetMtime(t, oneTwoOne, "2020-12-30")

	oneTwoTwo := filepath.Join(o.HomeDir, "versions", "1.2.2")
	require.NoError(t, os.MkdirAll(oneTwoTwo, 0700))
	requireSetMtime(t, oneTwoTwo, "2020-12-31")

	c, stdout, stderr := newApp(o)
	o.Out = stdout
	err := c.Run([]string{"getenvoy", "versions"})

	require.NoError(t, err)
	require.Equal(t, `VERSION	RELEASE_DATE
1.2.2	2020-12-31
1.1.2	2020-12-31
1.2.1	2020-12-30
`, stdout.String())
	require.Empty(t, stderr)
}

func TestGetEnvoyVersions_DirIsFile(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	require.NoError(t, os.WriteFile(filepath.Join(o.HomeDir, "versions"), []byte{}, 0700))

	c, stdout, stderr := newApp(o)
	o.Out = stdout
	err := c.Run([]string{"getenvoy", "versions"})

	require.Error(t, err)
	require.Empty(t, stdout)
	require.Empty(t, stderr)
}

func TestGetEnvoyVersions_SkipsFiles(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	oneOneTwo := filepath.Join(o.HomeDir, "versions", "1.1.2")
	require.NoError(t, os.MkdirAll(oneOneTwo, 0700))
	requireSetMtime(t, oneOneTwo, "2020-12-31")

	oneTwoOne := filepath.Join(o.HomeDir, "versions", "1.2.1")
	require.NoError(t, os.WriteFile(oneTwoOne, []byte{}, 0700)) // notice a file not a directory!
	requireSetMtime(t, oneTwoOne, "2020-12-30")

	oneTwoTwo := filepath.Join(o.HomeDir, "versions", "1.2.2")
	require.NoError(t, os.MkdirAll(oneTwoTwo, 0700))
	requireSetMtime(t, oneTwoTwo, "2020-12-31")

	c, stdout, stderr := newApp(o)
	o.Out = stdout
	err := c.Run([]string{"getenvoy", "versions"})

	require.NoError(t, err)
	require.Equal(t, `VERSION	RELEASE_DATE
1.2.2	2020-12-31
1.1.2	2020-12-31
`, stdout.String())
	require.Empty(t, stderr)
}

func TestGetEnvoyVersions_Remote(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	// Run "getenvoy versions"
	c, stdout, stderr := newApp(o)
	o.Out = stdout
	err := c.Run([]string{"getenvoy", "versions", "-a"})

	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(`VERSION	RELEASE_DATE
%s	2020-12-31
`, version.LastKnownEnvoy), stdout.String())
	require.Empty(t, stderr)
}

func TestGetEnvoyVersions_All_Mixed(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	oneOneTwo := filepath.Join(o.HomeDir, "versions", "1.1.2")
	require.NoError(t, os.MkdirAll(oneOneTwo, 0700))
	requireSetMtime(t, oneOneTwo, "2020-12-31")

	// Run "getenvoy versions"
	c, stdout, stderr := newApp(o)
	o.Out = stdout
	err := c.Run([]string{"getenvoy", "versions", "-a"})

	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(`VERSION	RELEASE_DATE
%s	2020-12-31
1.1.2	2020-12-31
`, version.LastKnownEnvoy), stdout.String())
	require.Empty(t, stderr)
}

func TestGetEnvoyVersions_All_PrefersInstallDate(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	installed := filepath.Join(o.HomeDir, "versions", version.LastKnownEnvoy)
	require.NoError(t, os.MkdirAll(installed, 0700))
	requireSetMtime(t, installed, "2021-12-31") // not 2020-12-31

	// Run "getenvoy versions"
	c, stdout, stderr := newApp(o)
	o.Out = stdout
	err := c.Run([]string{"getenvoy", "versions", "-a"})

	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(`VERSION	RELEASE_DATE
%s	2021-12-31
`, version.LastKnownEnvoy), stdout.String())
	require.Empty(t, stderr)
}

func requireSetMtime(t *testing.T, dir, date string) {
	td, err := time.Parse("2006-01-02", date)
	require.NoError(t, err)
	require.NoError(t, os.Chtimes(dir, td, td))
}
