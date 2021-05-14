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
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/reference"
	"github.com/tetratelabs/getenvoy/pkg/globals"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
	manifesttest "github.com/tetratelabs/getenvoy/pkg/test/manifest"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

func TestGetEnvoyRunValidateFlag(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedErr string
	}{
		{
			name:        "arg[0] missing",
			args:        []string{"getenvoy", "run"},
			expectedErr: `missing reference parameter`,
		},
		{
			name:        "arg[0] with invalid reference",
			args:        []string{"getenvoy", "run", "???"},
			expectedErr: `"???" is not a valid GetEnvoy reference. Expected format: [<flavor>:]<version>[/<platform>]`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			// Run "getenvoy run"
			c, stdout, stderr := newApp(&globals.GlobalOpts{})
			err := c.Run(test.args)

			// Verify the command failed with the expected error
			require.EqualError(t, err, test.expectedErr)
			// Main handles logging of errors, so we expect nothing in stdout or stderr
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
			args: []string{"getenvoy", "run", reference.Latest},
		},
		{
			name: "empty envoy args",
			args: []string{"getenvoy", "run", reference.Latest, "--"},
		},
		{
			name:              "envoy args",
			args:              []string{"getenvoy", "run", reference.Latest, "--", "-c", "envoy.yaml"},
			expectedEnvoyArgs: ` -c envoy.yaml`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			o, cleanup := setupTest(t)
			defer cleanup()

			// Run "getenvoy run standard:1.17.1 -- -c envoy.yaml"
			c, stdout, stderr := newApp(&o.GlobalOpts)
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
	c, stdout, stderr := newApp(&o.GlobalOpts)

	// Run "getenvoy run unknown"
	version := "unknown"
	err := c.Run([]string{"getenvoy", "run", version})

	// Verify the command failed with the expected error.
	r := version + "/" + o.platform
	expectedErr := fmt.Sprintf(`unable to find matching GetEnvoy build for reference "%s"`, r)
	require.EqualError(t, err, expectedErr)
	// Main handles logging of errors, so we expect nothing in stdout or stderr
	require.Empty(t, stdout)
	require.Empty(t, stderr)
}

type testEnvoyExtensionConfig struct {
	globals.GlobalOpts
	// platform is the types.Reference.Platform used in manifest commands
	platform string
}

// setupTest returns testEnvoyExtensionConfig and a tear-down function.
// The tear-down functions reverts side-effects such as temp directories and a fake manifest server.
func setupTest(t *testing.T) (*testEnvoyExtensionConfig, func()) {
	result := testEnvoyExtensionConfig{}
	result.Out = io.Discard // ignore logging
	var tearDown []func()

	tempDir, deleteTempDir := morerequire.RequireNewTempDir(t)
	tearDown = append(tearDown, deleteTempDir)

	result.HomeDir = filepath.Join(tempDir, "envoy_home")
	err := os.Mkdir(result.HomeDir, 0700)
	require.NoError(t, err, `error creating directory: %s`, result.HomeDir)

	ref, err := manifest.ParseReference(reference.Latest)
	require.NoError(t, err, `error resolving manifest for reference: %s`, ref)
	result.platform = ref.Platform

	testManifest, err := manifesttest.NewSimpleManifest(ref.String())
	require.NoError(t, err, `error creating test manifest`)

	manifestServer := manifesttest.RequireManifestTestServer(t, testManifest)
	result.ManifestURL = manifestServer.URL + "/manifest.json"
	tearDown = append(tearDown, manifestServer.Close)

	return &result, func() {
		for i := len(tearDown) - 1; i >= 0; i-- {
			tearDown[i]()
		}
	}
}
