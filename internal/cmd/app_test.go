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

	rootcmd "github.com/tetratelabs/func-e/internal/cmd"
	"github.com/tetratelabs/func-e/internal/envoy"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/test"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestFuncEValidateArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedErr string
	}{
		{
			name:        "--envoy-versions-url not a URL",
			args:        []string{"func-e", "--envoy-versions-url", "/not/url"},
			expectedErr: `"/not/url" is not a valid Envoy versions URL`,
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			err := runTestCommand(t, &globals.GlobalOpts{}, tc.args)
			require.EqualError(t, err, tc.expectedErr)
		})
	}
}

func TestHomeDir(t *testing.T) {
	type testCase struct {
		name     string
		args     []string
		setup    func()
		expected string
	}

	u, err := user.Current()
	require.NoError(t, err)

	alt1 := filepath.Join(u.HomeDir, "alt1")
	alt2 := filepath.Join(u.HomeDir, "alt2")

	tests := []testCase{ // we don't test default as that depends on the runtime env
		{
			name:     "default is ~/.func-e",
			args:     []string{"func-e"},
			expected: filepath.Join(u.HomeDir, ".func-e"),
		},
		{
			name: "FUNC_E_HOME env",
			args: []string{"func-e"},
			setup: func() {
				t.Setenv("FUNC_E_HOME", alt1)
			},
			expected: alt1,
		},
		{
			name:     "--home-dir arg",
			args:     []string{"func-e", "--home-dir", alt1},
			expected: alt1,
		},
		{
			name: "prioritizes --home-dir arg over FUNC_E_HOME env",
			args: []string{"func-e", "--home-dir", alt1},
			setup: func() {
				t.Setenv("FUNC_E_HOME", alt2)
			},
			expected: alt1,
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}

			o := &globals.GlobalOpts{}
			err := runTestCommand(t, o, tc.args)

			require.NoError(t, err)
			require.Equal(t, tc.expected, o.HomeDir)
		})
	}
}

func TestPlatformArg(t *testing.T) {
	type testCase struct {
		name     string
		args     []string
		setup    func()
		expected version.Platform
	}

	tests := []testCase{
		{
			name: "FUNC_E_PLATFORM env",
			args: []string{"func-e"},
			setup: func() {
				t.Setenv("FUNC_E_PLATFORM", "linux/amd64")
			},
			expected: version.Platform("linux/amd64"),
		},
		{
			name:     "--platform flag",
			args:     []string{"func-e", "--platform", "darwin/amd64"},
			expected: version.Platform("darwin/amd64"),
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}

			o := &globals.GlobalOpts{}
			err := runTestCommand(t, o, tc.args)
			require.NoError(t, err)
			require.Equal(t, tc.expected, o.Platform)
		})
	}
}

func TestEnvoyVersionsURL(t *testing.T) {
	type testCase struct {
		name     string
		args     []string
		setup    func()
		expected string
	}

	tests := []testCase{ // we don't test default as that depends on the runtime env
		{
			name:     "default is https://archive.tetratelabs.io/envoy/envoy-versions.json",
			args:     []string{"func-e"},
			expected: "https://archive.tetratelabs.io/envoy/envoy-versions.json",
		},
		{
			name: "ENVOY_VERSIONS_URL env",
			args: []string{"func-e"},
			setup: func() {
				t.Setenv("ENVOY_VERSIONS_URL", "http://ENVOY_VERSIONS_URL/env")
			},
			expected: "http://ENVOY_VERSIONS_URL/env",
		},
		{
			name:     "--envoy-versions-url flag",
			args:     []string{"func-e", "--envoy-versions-url", "http://versions/arg"},
			expected: "http://versions/arg",
		},
		{
			name: "prioritizes --envoy-versions-url arg over ENVOY_VERSIONS_URL env",
			args: []string{"func-e", "--envoy-versions-url", "http://versions/arg"},
			setup: func() {
				t.Setenv("ENVOY_VERSIONS_URL", "http://ENVOY_VERSIONS_URL/env")
			},
			expected: "http://versions/arg",
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}

			o := &globals.GlobalOpts{}
			err := runTestCommand(t, o, tc.args)

			require.NoError(t, err)
			require.Equal(t, tc.expected, o.EnvoyVersionsURL)
		})
	}
}

// newApp initializes a command with buffers for stdout and stderr.
func newApp(o *globals.GlobalOpts) (c *cli.App, stdout, stderr *bytes.Buffer) {
	stdout = new(bytes.Buffer)
	stderr = new(bytes.Buffer)
	c = rootcmd.NewApp(o)
	c.Name = "func-e"
	c.Writer = stdout
	c.ErrWriter = stderr
	o.Out = stdout
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
// The tear-down functions reverts side-effects such as temp directories and a fake Envoy versions server.
func setupTest(t *testing.T) *globals.GlobalOpts {
	result := globals.GlobalOpts{}
	result.EnvoyVersion = version.LastKnownEnvoy
	result.Out = io.Discard // ignore logging by default

	tempDir := t.TempDir()

	result.HomeDir = filepath.Join(tempDir, "envoy_home")
	err := os.Mkdir(result.HomeDir, 0o700)
	require.NoError(t, err, `error creating directory: %s`, result.HomeDir)

	versionsServer := test.RequireEnvoyVersionsTestServer(t, version.LastKnownEnvoy)
	result.EnvoyVersionsURL = versionsServer.URL + "/envoy-versions.json"
	result.GetEnvoyVersions = envoy.NewGetVersions(result.EnvoyVersionsURL, result.Platform, result.Version)

	t.Cleanup(versionsServer.Close)
	return &result
}
