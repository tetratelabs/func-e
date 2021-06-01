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

	"github.com/tetratelabs/getenvoy/internal/test"
	"github.com/tetratelabs/getenvoy/internal/version"
)

func TestGetEnvoyInstall_VersionValidates(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	tests := []struct{ name, version, expectedErr string }{
		{
			name:        "version empty",
			expectedErr: fmt.Sprintf(`invalid [version] argument: "" should look like "%s"`, version.LastKnownEnvoy),
		},
		{
			name:        "version invalid",
			version:     "a.b.c",
			expectedErr: fmt.Sprintf(`invalid [version] argument: "a.b.c" should look like "%s"`, version.LastKnownEnvoy),
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			c, stdout, stderr := newApp(o)
			err := c.Run([]string{"getenvoy", "install", tc.version})

			// Verify the command failed with the expected error
			require.EqualError(t, err, tc.expectedErr)
			// GetEnvoy handles logging of errors, so we expect nothing in stdout or stderr
			require.Empty(t, stdout)
			require.Empty(t, stderr)
		})
	}
}

func TestGetEnvoyInstall_PreservesReleaseDate(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	c, _, _ := newApp(o)
	require.NoError(t, c.Run([]string{"getenvoy", "install", o.EnvoyVersion}))

	// The directory was created
	versionDir := filepath.Join(o.HomeDir, "versions", o.EnvoyVersion)
	require.DirExists(t, versionDir)

	// The directory timestamp matches the fake release date, not the current time
	f, err := os.Stat(versionDir)
	require.NoError(t, err)
	require.Equal(t, f.ModTime().Format("2006-01-02"), test.FakeReleaseDate)
}
