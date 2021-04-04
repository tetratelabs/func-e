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
	"fmt"
	"path/filepath"
	"testing"

	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/common"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
	cmdtest "github.com/tetratelabs/getenvoy/pkg/test/cmd"
	. "github.com/tetratelabs/getenvoy/pkg/test/morerequire"
	cmdutil "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

func TestGetEnvoyValidateArgs(t *testing.T) {
	type testCase struct {
		name        string
		args        []string
		expectedErr string
	}

	tests := []testCase{
		{
			name:        "--home-dir empty",
			args:        []string{"--home-dir", ""},
			expectedErr: `GetEnvoy home directory cannot be empty`,
		},
		{
			name:        "--manifest empty",
			args:        []string{"--manifest", ""},
			expectedErr: `GetEnvoy manifest URL cannot be empty`,
		},
		{
			name:        "--manifest not a URL",
			args:        []string{"--manifest", "/not/url"},
			expectedErr: `"/not/url" is not a valid manifest URL`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			c, stdout, stderr := cmdtest.NewRootCommand()
			c.SetArgs(append(test.args, "help"))
			err := cmdutil.Execute(c)
			require.EqualError(t, err, test.expectedErr, `expected an error running [%v]`, c)

			// Verify the command failed with the expected error
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

	emptySetup := func() func() {
		return func() {
		}
	}

	home, err := homedir.Dir()
	require.NoError(t, err, `error getting current user's home dir'`)
	defaultHomeDir, err := filepath.Abs(filepath.Join(home, ".getenvoy"))
	require.NoError(t, err, `error resolving absolute path to default GETENVOY_HOME'`)

	tests := []testCase{ // we don't test default as that depends on the runtime env
		{
			name:     "default is ~/.getenvoy",
			setup:    emptySetup,
			expected: defaultHomeDir,
		},
		{
			name: "GETENVOY_HOME env",
			setup: func() func() {
				return RequireSetenv(t, "GETENVOY_HOME", "/from/GETENVOY_HOME/env")
			},
			expected: "/from/GETENVOY_HOME/env",
		},
		{
			name:     "--home-dir arg",
			args:     []string{"--home-dir", "/from/home-dir/arg"},
			setup:    emptySetup,
			expected: "/from/home-dir/arg",
		},
		{
			name: "prioritizes --home-dir arg over GETENVOY_HOME env",
			args: []string{"--home-dir", "/from/home-dir/arg"},
			setup: func() func() {
				return RequireSetenv(t, "GETENVOY_HOME", "/from/GETENVOY_HOME/env")
			},
			expected: "/from/home-dir/arg",
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			tearDown := test.setup()
			defer tearDown()
			c, stdout, stderr := cmdtest.NewRootCommand()
			c.SetArgs(append(test.args, "help"))
			err := cmdutil.Execute(c)

			require.NoError(t, err, `expected no error running [%v]`, c)
			require.NotEmpty(t, stdout.String(), `expected stdout running [%v]`, c)
			require.Empty(t, stderr.String(), `expected no stderr running [%v]`, c)

			require.Equal(t, test.expected, common.HomeDir)
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

	emptySetup := func() func() {
		return func() {
		}
	}

	tests := []testCase{ // we don't test default as that depends on the runtime env
		{
			name:     "default is https://tetrate.bintray.com/getenvoy/manifest.json",
			setup:    emptySetup,
			expected: "https://tetrate.bintray.com/getenvoy/manifest.json",
		},
		{
			name: "GETENVOY_MANIFEST_URL env",
			setup: func() func() {
				return RequireSetenv(t, "GETENVOY_MANIFEST_URL", "http://GETENVOY_MANIFEST_URL/env")
			},
			expected: "http://GETENVOY_MANIFEST_URL/env",
		},
		{
			name:     "--manifest arg",
			args:     []string{"--manifest", "http://manifest/arg"},
			setup:    emptySetup,
			expected: "http://manifest/arg",
		},
		{
			name: "prioritizes --manifest arg over GETENVOY_MANIFEST_URL env",
			args: []string{"--manifest", "http://manifest/arg"},
			setup: func() func() {
				return RequireSetenv(t, "GETENVOY_MANIFEST_URL", "http://GETENVOY_MANIFEST_URL/env")
			},
			expected: "http://manifest/arg",
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			tearDown := test.setup()
			defer tearDown()
			c, stdout, stderr := cmdtest.NewRootCommand()
			c.SetArgs(append(test.args, "help"))
			err := cmdutil.Execute(c)

			require.NoError(t, err, `expected no error running [%v]`, c)
			require.NotEmpty(t, stdout.String(), `expected stdout running [%v]`, c)
			require.Empty(t, stderr.String(), `expected no stderr running [%v]`, c)

			require.Equal(t, test.expected, manifest.GetURL())
		})
	}
}
