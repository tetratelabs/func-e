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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetEnvoyInstalled_NothingYet(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	c, stdout, stderr := newApp(o)
	o.Out = stdout
	err := c.Run([]string{"getenvoy", "installed"})

	require.NoError(t, err)
	require.Equal(t, `No envoy versions installed, yet
`, stdout.String())
	require.Empty(t, stderr)
}

// TestGetEnvoyInstalled verifies output is sorted
func TestGetEnvoyInstalled_Sorted(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	require.NoError(t, os.MkdirAll(filepath.Join(o.HomeDir, "versions", "1.16.1"), 0700))
	require.NoError(t, os.MkdirAll(filepath.Join(o.HomeDir, "versions", "1.17.2"), 0700))

	c, stdout, stderr := newApp(o)
	o.Out = stdout
	err := c.Run([]string{"getenvoy", "installed"})

	require.NoError(t, err)
	require.Equal(t, `VERSION
1.17.2
1.16.1
`, stdout.String())
	require.Empty(t, stderr)
}

func TestGetEnvoyInstalled_DirIsFile(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	require.NoError(t, os.WriteFile(filepath.Join(o.HomeDir, "versions"), []byte{}, 0700))

	c, stdout, stderr := newApp(o)
	o.Out = stdout
	err := c.Run([]string{"getenvoy", "installed"})

	require.Error(t, err)
	require.Empty(t, stdout)
	require.Empty(t, stderr)
}

func TestGetEnvoyInstalled_SkipsFiles(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	require.NoError(t, os.MkdirAll(filepath.Join(o.HomeDir, "versions", "1.16.1"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(o.HomeDir, "versions", "1.17.1"), []byte{}, 0700))
	require.NoError(t, os.MkdirAll(filepath.Join(o.HomeDir, "versions", "1.17.2"), 0700))

	c, stdout, stderr := newApp(o)
	o.Out = stdout
	err := c.Run([]string{"getenvoy", "installed"})

	require.NoError(t, err)
	require.Equal(t, `VERSION
1.17.2
1.16.1
`, stdout.String())
	require.Empty(t, stderr)
}
