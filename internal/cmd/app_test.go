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
	"io"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	rootcmd "github.com/tetratelabs/getenvoy/internal/cmd"
	"github.com/tetratelabs/getenvoy/internal/globals"
	manifesttest "github.com/tetratelabs/getenvoy/internal/test/manifest"
	"github.com/tetratelabs/getenvoy/internal/test/morerequire"
	"github.com/tetratelabs/getenvoy/internal/version"
)

func TestGetEnvoyValidateArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedErr string
	}{
		{
			name:        "--manifest not a URL",
			args:        []string{"getenvoy", "--manifest", "/not/url"},
			expectedErr: `"/not/url" is not a valid manifest URL`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			err := runTestCommand(t, &globals.GlobalOpts{}, test.args)
			require.EqualError(t, err, test.expectedErr)
		})
	}
}

func TestGetEnvoyHomeDir(t *testing.T) {
	type testCase struct {
		name string
		args []string
		// setup returns a tear-down function
		setup    func() func()
		expected string
	}

	u, err := user.Current()
	require.NoError(t, err)

	tests := []testCase{ // we don't test default as that depends on the runtime env
		{
			name:     "default is ~/.getenvoy",
			args:     []string{"getenvoy"},
			expected: filepath.Join(u.HomeDir, ".getenvoy"),
		},
		{
			name: "GETENVOY_HOME env",
			args: []string{"getenvoy"},
			setup: func() func() {
				return requireSetenv(t, "GETENVOY_HOME", "/from/GETENVOY_HOME/env")
			},
			expected: "/from/GETENVOY_HOME/env",
		},
		{
			name:     "--home-dir arg",
			args:     []string{"getenvoy", "--home-dir", "/from/home-dir/arg"},
			expected: "/from/home-dir/arg",
		},
		{
			name: "prioritizes --home-dir arg over GETENVOY_HOME env",
			args: []string{"getenvoy", "--home-dir", "/from/home-dir/arg"},
			setup: func() func() {
				return requireSetenv(t, "GETENVOY_HOME", "/from/GETENVOY_HOME/env")
			},
			expected: "/from/home-dir/arg",
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			if test.setup != nil {
				tearDown := test.setup()
				defer tearDown()
			}

			o := &globals.GlobalOpts{}
			err := runTestCommand(t, o, test.args)

			require.NoError(t, err)
			require.Equal(t, test.expected, o.HomeDir)
		})
	}
}

func TestGetEnvoyManifest(t *testing.T) {
	type testCase struct {
		name string
		args []string
		// setup returns a tear-down function
		setup    func() func()
		expected string
	}

	tests := []testCase{ // we don't test default as that depends on the runtime env
		{
			name:     "default is https://dl.getenvoy.io/public/raw/files/manifest.json",
			args:     []string{"getenvoy"},
			expected: "https://dl.getenvoy.io/public/raw/files/manifest.json",
		},
		{
			name: "GETENVOY_MANIFEST_URL env",
			args: []string{"getenvoy"},
			setup: func() func() {
				return requireSetenv(t, "GETENVOY_MANIFEST_URL", "http://GETENVOY_MANIFEST_URL/env")
			},
			expected: "http://GETENVOY_MANIFEST_URL/env",
		},
		{
			name:     "--manifest arg",
			args:     []string{"getenvoy", "--manifest", "http://manifest/arg"},
			expected: "http://manifest/arg",
		},
		{
			name: "prioritizes --manifest arg over GETENVOY_MANIFEST_URL env",
			args: []string{"getenvoy", "--manifest", "http://manifest/arg"},
			setup: func() func() {
				return requireSetenv(t, "GETENVOY_MANIFEST_URL", "http://GETENVOY_MANIFEST_URL/env")
			},
			expected: "http://manifest/arg",
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			if test.setup != nil {
				tearDown := test.setup()
				defer tearDown()
			}

			o := &globals.GlobalOpts{}
			err := runTestCommand(t, o, test.args)

			require.NoError(t, err)
			require.Equal(t, test.expected, o.ManifestURL)
		})
	}
}

// requireSetenv will os.Setenv the given key and value. The function returned reverts to the original.
func requireSetenv(t *testing.T, key, value string) func() {
	previous := os.Getenv(key)
	err := os.Setenv(key, value)
	require.NoError(t, err, `error setting env variable %s=%s`, key, value)
	return func() {
		e := os.Setenv(key, previous)
		require.NoError(t, e, `error reverting env variable %s=%s`, key, previous)
	}
}

// newApp initializes a command with buffers for stdout and stderr.
func newApp(o *globals.GlobalOpts) (c *cli.App, stdout, stderr *bytes.Buffer) {
	stdout = new(bytes.Buffer)
	stderr = new(bytes.Buffer)
	c = rootcmd.NewApp(o)
	c.Name = "getenvoy"
	c.Writer = stdout
	c.ErrWriter = stderr
	return
}

func runTestCommand(t *testing.T, o *globals.GlobalOpts, args []string) error {
	c, stdout, stderr := newApp(o)
	c.Commands = append(c.Commands, &cli.Command{Name: "test", Action: func(_ *cli.Context) error {
		return nil
	}})

	err := c.Run(append(args, "test"))

	// Main handles logging of errors, so we expect nothing in stdout or stderr even in error case
	require.Empty(t, stdout)
	require.Empty(t, stderr)
	return err
}

// setupTest returns globals.GlobalOpts and a tear-down function.
// The tear-down functions reverts side-effects such as temp directories and a fake manifest server.
func setupTest(t *testing.T) (*globals.GlobalOpts, func()) {
	result := globals.GlobalOpts{}
	result.Out = io.Discard // ignore logging by default
	var tearDown []func()

	tempDir, deleteTempDir := morerequire.RequireNewTempDir(t)
	tearDown = append(tearDown, deleteTempDir)

	result.HomeDir = filepath.Join(tempDir, "envoy_home")
	err := os.Mkdir(result.HomeDir, 0700)
	require.NoError(t, err, `error creating directory: %s`, result.HomeDir)

	manifestServer := manifesttest.RequireManifestTestServer(t, version.Envoy)
	result.ManifestURL = manifestServer.URL + "/manifest.json"
	tearDown = append(tearDown, manifestServer.Close)

	return &result, func() {
		for i := len(tearDown) - 1; i >= 0; i-- {
			tearDown[i]()
		}
	}
}
