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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/version"
)

func TestGetEnvoyRunValidateFlag(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		expectedErr string
	}{
		{
			name:        "arg[0] missing",
			args:        []string{"getenvoy", "run"},
			expectedErr: `missing <version> argument`,
		},
		{
			name:        "arg[0] with invalid version",
			args:        []string{"getenvoy", "run", "unknown"},
			expectedErr: fmt.Sprintf(`invalid <version> argument: "unknown" should look like "%s"`, version.Envoy),
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			// Run "getenvoy run"
			c, stdout, stderr := newApp(o)
			err := c.Run(test.args)

			// Verify the command failed with the expected error
			require.EqualError(t, err, test.expectedErr)
			// GetEnvoy handles logging of errors, so we expect nothing in stdout or stderr
			require.Empty(t, stdout)
			require.Empty(t, stderr)
		})
	}
}

func TestGetEnvoyRun(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectedEnvoyArgs string
	}{
		{
			name: "no envoy args",
			args: []string{"getenvoy", "run", version.Envoy},
		},
		{
			name: "empty envoy args",
			args: []string{"getenvoy", "run", version.Envoy},
		},
		{
			name:              "envoy args",
			args:              []string{"getenvoy", "run", version.Envoy, "-c", "envoy.yaml"},
			expectedEnvoyArgs: ` -c envoy.yaml`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			o, cleanup := setupTest(t)
			defer cleanup()

			// Run "getenvoy run 1.17.3 -c envoy.yaml"
			c, stdout, stderr := newApp(o)
			err := c.Run(test.args)

			// Verify the command invoked, passing the correct default commandline
			require.NoError(t, err)

			// We expect getenvoy to print the context it will run, and Envoy to execute the same, except adding the
			// --admin-address-path flag
			expectedStdout := fmt.Sprintf(`starting: %[2]s%[3]s
working directory: %[1]s
envoy wd: %[1]s
envoy bin: %[2]s
envoy args:%[3]s --admin-address-path admin-address.txt`, o.WorkingDir, o.EnvoyPath, test.expectedEnvoyArgs)
			require.Equal(t, expectedStdout+"\n", stdout.String())
			require.Equal(t, "envoy stderr\n", stderr.String())
		})
	}
}

func TestGetEnvoyRunFailWithUnknownVersion(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	o.EnvoyPath = "" // force lookup of version flag
	c, stdout, stderr := newApp(o)

	// Run "getenvoy run unknown"
	err := c.Run([]string{"getenvoy", "run", "unknown"})

	// Verify the command failed with the expected error.
	require.EqualError(t, err, fmt.Sprintf(`invalid <version> argument: "unknown" should look like "%s"`, version.Envoy))
	// GetEnvoy handles logging of errors, so we expect nothing in stdout or stderr
	require.Empty(t, stdout)
	require.Empty(t, stderr)
}
