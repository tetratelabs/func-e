// Copyright 2020 Tetrate
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
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/globals"
	"github.com/tetratelabs/getenvoy/internal/test/morerequire"
	"github.com/tetratelabs/getenvoy/internal/version"
)

func TestGetEnvoyRun(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectedEnvoyArgs string
	}{
		{
			name: "no envoy args",
			args: []string{"getenvoy", "run"},
		},
		{
			name:              "envoy args",
			args:              []string{"getenvoy", "run", "-c", "envoy.yaml"},
			expectedEnvoyArgs: `-c envoy.yaml `,
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			o, cleanup := setupTest(t)
			defer cleanup()

			c, stdout, stderr := newApp(o)
			o.Out = io.Discard // don't verify logging
			err := c.Run(tc.args)

			// Verify the command invoked, passing the correct default commandline
			require.NoError(t, err)

			// We expect getenvoy to print the context it will run, and Envoy to execute the same, except adding the
			// --admin-address-path flag
			expectedEnvoyArgs := fmt.Sprint(tc.expectedEnvoyArgs, "--admin-address-path ", filepath.Join(o.RunDir, "admin-address.txt"))
			expectedStdout := fmt.Sprintf(`starting: %[1]s %[2]s
envoy bin: %[1]s
envoy args: %[2]s`, o.EnvoyPath, expectedEnvoyArgs)
			require.Equal(t, expectedStdout+"\n", stdout.String())
			require.Equal(t, "envoy stderr\n", stderr.String())
		})
	}
}

func TestGetEnvoyRun_ReadsHomeVersionFile(t *testing.T) {
	o, cleanup := setupTest(t)
	o.EnvoyVersion = "" // pretend this is an initial setup
	o.Out = new(bytes.Buffer)
	defer cleanup()

	require.NoError(t, os.WriteFile(filepath.Join(o.HomeDir, "version"), []byte(version.LastKnownEnvoy), 0600))

	c, _, _ := newApp(o)
	require.NoError(t, c.Run([]string{"getenvoy", "run"}))

	// No implicit lookup
	require.NotContains(t, o.Out.(*bytes.Buffer).String(), "looking up latest version\n")
	require.Equal(t, version.LastKnownEnvoy, o.EnvoyVersion)
}

func TestGetEnvoyRun_CreatesHomeVersionFile(t *testing.T) {
	o, cleanup := setupTest(t)
	o.EnvoyVersion = "" // pretend this is an initial setup
	o.Out = new(bytes.Buffer)
	defer cleanup()

	// make sure first run where the home doesn't exist yet, works!
	require.NoError(t, os.RemoveAll(o.HomeDir))

	c, _, _ := newApp(o)
	require.NoError(t, c.Run([]string{"getenvoy", "run"}))

	// We logged the implicit lookup
	require.Contains(t, o.Out.(*bytes.Buffer).String(), "looking up latest version\n")
	require.FileExists(t, filepath.Join(o.HomeDir, "version"))
	require.Equal(t, version.LastKnownEnvoy, o.EnvoyVersion)
}

func TestGetEnvoyRun_ValidatesHomeVersion(t *testing.T) {
	o, cleanup := setupTest(t)
	o.Out = new(bytes.Buffer)
	defer cleanup()

	o.EnvoyVersion = ""
	require.NoError(t, os.WriteFile(filepath.Join(o.HomeDir, "version"), []byte("a.a.a"), 0600))

	c, _, _ := newApp(o)
	err := c.Run([]string{"getenvoy", "run"})

	// Verify the command failed with the expected error
	require.EqualError(t, err, fmt.Sprintf(`invalid version in "$GETENVOY_HOME/version": "a.a.a" should look like "%s"`, version.LastKnownEnvoy))
}

// TestGetEnvoyRun_ValidatesWorkingVersion duplicates logic in version_test.go to ensure a non-home version validates.
func TestGetEnvoyRun_ValidatesWorkingVersion(t *testing.T) {
	o, cleanup := setupTest(t)
	o.Out = new(bytes.Buffer)
	o.EnvoyVersion = ""
	defer cleanup()

	revertTempWd := morerequire.RequireChdirIntoTemp(t)
	defer revertTempWd()
	require.NoError(t, os.WriteFile(".envoy-version", []byte("b.b.b"), 0600))

	c, _, _ := newApp(o)
	err := c.Run([]string{"getenvoy", "run"})

	// Verify the command failed with the expected error
	require.EqualError(t, err, fmt.Sprintf(`invalid version in "$PWD/.envoy-version": "b.b.b" should look like "%s"`, version.LastKnownEnvoy))
}

func TestGetEnvoyRun_ErrsWhenVersionsServerDown(t *testing.T) {
	tempDir, deleteTempDir := morerequire.RequireNewTempDir(t)
	defer deleteTempDir()

	o := &globals.GlobalOpts{
		EnvoyVersionsURL: "https://127.0.0.1:9999",
		HomeDir:          tempDir,
		Out:              new(bytes.Buffer),
	}
	c, _, _ := newApp(o)
	err := c.Run([]string{"getenvoy", "run"})

	require.Contains(t, o.Out.(*bytes.Buffer).String(), "looking up latest version\n")
	require.Contains(t, err.Error(), fmt.Sprintf(`couldn't read latest version from %s`, o.EnvoyVersionsURL))
}
