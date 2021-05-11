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
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	rootcmd "github.com/tetratelabs/getenvoy/pkg/cmd"
	"github.com/tetratelabs/getenvoy/pkg/globals"
)

func TestGetEnvoyValidateArgs(t *testing.T) {
	o := &globals.GlobalOpts{}

	tests := []struct {
		name        string
		args        []string
		expectedErr string
	}{
		{
			name:        "--home-dir empty",
			args:        []string{"--home-dir", "", "help"},
			expectedErr: `GetEnvoy home directory cannot be empty`,
		},
		{
			name:        "--manifest empty",
			args:        []string{"--manifest", "", "help"},
			expectedErr: `GetEnvoy manifest URL cannot be empty`,
		},
		{
			name:        "--manifest not a URL",
			args:        []string{"--manifest", "/not/url", "help"},
			expectedErr: `"/not/url" is not a valid manifest URL`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			c, stdout, stderr := newRootCommand(o)
			c.SetArgs(test.args)
			err := rootcmd.Execute(c)

			// Verify the command failed with the expected error
			require.EqualError(t, err, test.expectedErr, `expected an error running [%v]`, c)
			require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
			expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy help --help' for usage.\n", test.expectedErr)
			require.Equal(t, expectedStderr, stderr.String(), `unexpected stderr running [%v]`, c)
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

	tests := []testCase{ // we don't test default as that depends on the runtime env
		{
			name:     "default is ~/.getenvoy",
			args:     []string{"help"},
			expected: globals.DefaultHomeDir(),
		},
		{
			name: "GETENVOY_HOME env",
			args: []string{"help"},
			setup: func() func() {
				return requireSetenv(t, "GETENVOY_HOME", "/from/GETENVOY_HOME/env")
			},
			expected: "/from/GETENVOY_HOME/env",
		},
		{
			name:     "--home-dir arg",
			args:     []string{"--home-dir", "/from/home-dir/arg", "help"},
			expected: "/from/home-dir/arg",
		},
		{
			name: "prioritizes --home-dir arg over GETENVOY_HOME env",
			args: []string{"--home-dir", "/from/home-dir/arg", "help"},
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
			c, stdout, stderr := newRootCommand(o)
			c.SetArgs(test.args)
			err := rootcmd.Execute(c)

			require.NoError(t, err, `expected no error running [%v]`, c)
			require.NotEmpty(t, stdout.String(), `expected stdout running [%v]`, c)
			require.Empty(t, stderr.String(), `expected no stderr running [%v]`, c)

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
			expected: "https://dl.getenvoy.io/public/raw/files/manifest.json",
		},
		{
			name: "GETENVOY_MANIFEST_URL env",
			setup: func() func() {
				return requireSetenv(t, "GETENVOY_MANIFEST_URL", "http://GETENVOY_MANIFEST_URL/env")
			},
			expected: "http://GETENVOY_MANIFEST_URL/env",
		},
		{
			name:     "--manifest arg",
			args:     []string{"--manifest", "http://manifest/arg"},
			expected: "http://manifest/arg",
		},
		{
			name: "prioritizes --manifest arg over GETENVOY_MANIFEST_URL env",
			args: []string{"--manifest", "http://manifest/arg"},
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
			c, stdout, stderr := newRootCommand(o)
			c.SetArgs(append(test.args, "help"))
			err := rootcmd.Execute(c)

			require.NoError(t, err, `expected no error running [%v]`, c)
			require.NotEmpty(t, stdout.String(), `expected stdout running [%v]`, c)
			require.Empty(t, stderr.String(), `expected no stderr running [%v]`, c)

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

// newRootCommand initializes a command with buffers for stdout and stderr.
func newRootCommand(o *globals.GlobalOpts) (c *cobra.Command, stdout, stderr *bytes.Buffer) {
	stdout = new(bytes.Buffer)
	stderr = new(bytes.Buffer)
	c = rootcmd.NewRoot(o)
	c.SetOut(stdout)
	c.SetErr(stderr)
	return c, stdout, stderr
}
