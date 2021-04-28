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

package example_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/globals"
	rootcmd "github.com/tetratelabs/getenvoy/pkg/cmd"
	"github.com/tetratelabs/getenvoy/pkg/test/cmd"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

func TestGetEnvoyExtensionExamplesRemoveValidateFlag(t *testing.T) {
	type testCase struct {
		name        string
		args        []string
		expectedErr string
	}

	tests := []testCase{
		{
			name:        "--name is missing",
			args:        []string{"extension", "examples", "remove"},
			expectedErr: `example name cannot be empty`,
		},
		{
			name:        "--name with invalid value",
			args:        []string{"extension", "examples", "remove", "--name", "my:example"},
			expectedErr: `"my:example" is not a valid example name. Example name must match the format "^[a-z0-9._-]+$". E.g., 'my.example', 'my-example' or 'my_example'`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			// Run "getenvoy extension examples remove" with the flags we are testing
			c, stdout, stderr := cmd.NewRootCommand(&globals.GlobalOpts{})
			c.SetArgs(test.args)
			err := rootcmd.Execute(c)

			// Verify the command failed with the expected error
			require.EqualError(t, err, test.expectedErr, `expected an error running [%v]`, c)
			require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
			expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy extension examples remove --help' for usage.\n", test.expectedErr)
			require.Equal(t, expectedStderr, stderr.String(), `unexpected stderr running [%v]`, c)
		})
	}
}

func TestGetEnvoyExtensionExamplesRemoveFailsOutsideExtensionDirectory(t *testing.T) {
	// Change to a non-workspace dir
	o := &globals.GlobalOpts{ExtensionDir: morerequire.RequireAbs(t, relativeRustExtensionDirWithOneExample+"/..")}

	// Run "getenvoy extension examples remove"
	c, stdout, stderr := cmd.NewRootCommand(o)
	c.SetArgs([]string{"extension", "examples", "remove", "--name", "default"})
	err := rootcmd.Execute(c)

	// Verify the command failed with the expected error
	expectedErr := fmt.Sprintf("not an extension directory %q", o.ExtensionDir)
	require.EqualError(t, err, expectedErr, `expected an error running [%v]`, c)
	require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
	expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy extension examples remove --help' for usage.\n", expectedErr)
	require.Equal(t, expectedStderr, stderr.String(), `unexpected stderr running [%v]`, c)
}

func TestGetEnvoyExtensionExamplesRemoveWarnsOnMissingExample(t *testing.T) {
	o := &globals.GlobalOpts{ExtensionDir: morerequire.RequireAbs(t, relativeRustExtensionDirWithOneExample)}

	// Run "getenvoy extension examples remove"
	c, stdout, stderr := cmd.NewRootCommand(o)
	name := "doesnt-exist"
	c.SetArgs([]string{"extension", "examples", "remove", "--name", name})
	err := rootcmd.Execute(c)

	// The command shouldn't err, but it should warn to stderr
	require.NoError(t, err, `expected no error running [%v]`, c)
	require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
	require.Equal(t, fmt.Sprintf(`There is no example setup named "%s".

Use "getenvoy extension examples list" to list existing example setups.
`, name), stderr.String(), `unexpected stderr running [%v]`, c)
}

func TestGetEnvoyExtensionExamplesRemoveDefault(t *testing.T) {
	// Copy the workspace as this test will delete it, and we don't want to mutate our test data!
	extensionDir, revertExtensionDir := morerequire.RequireCopyOfDir(t, relativeRustExtensionDirWithOneExample)
	defer revertExtensionDir()

	// "getenvoy extension examples remove" must be in a valid extension directory
	o := &globals.GlobalOpts{ExtensionDir: extensionDir}

	// Run "getenvoy extension examples remove"
	c, stdout, stderr := cmd.NewRootCommand(o)
	c.SetArgs([]string{"extension", "examples", "remove", "--name", "default"})
	err := rootcmd.Execute(c)

	// Verify the files deleted end up in stderr
	require.NoError(t, err, `expected no error running [%v]`, c)
	require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
	require.Equal(t, `Removing example setup:
* .getenvoy/extension/examples/default/envoy.tmpl.yaml
* .getenvoy/extension/examples/default/example.yaml
* .getenvoy/extension/examples/default/extension.json
Done!
`, stderr.String(), `unexpected stderr running [%v]`, c)

	// Verify the directory actually deleted
	require.NoDirExists(t, filepath.Join(extensionDir, ".getenvoy/extension/examples/default"), `expected to not delete example "default"`)
}

func TestGetEnvoyExtensionExamplesRemoveDoesntEffectOtherExample(t *testing.T) {
	// Copy the workspace as this test will delete it, and we don't want to mutate our test data!
	extensionDir, revertExtensionDir := morerequire.RequireCopyOfDir(t, relativeExtensionDirWithTwoExamples)
	defer revertExtensionDir()

	// "getenvoy extension examples remove" must be in a valid extension directory
	o := &globals.GlobalOpts{ExtensionDir: extensionDir}

	// Run "getenvoy extension examples remove"
	c, stdout, stderr := cmd.NewRootCommand(o)
	c.SetArgs([]string{"extension", "examples", "remove", "--name", "another"})
	err := rootcmd.Execute(c)

	// Verify the files deleted end up in stderr
	require.NoError(t, err, `expected no error running [%v]`, c)
	require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
	require.Equal(t, `Removing example setup:
* .getenvoy/extension/examples/another/envoy.tmpl.yaml
* .getenvoy/extension/examples/another/example.yaml
* .getenvoy/extension/examples/another/extension.json
Done!
`, stderr.String(), `unexpected stderr running [%v]`, c)

	// Verify the other example still exists
	require.NoDirExists(t, filepath.Join(extensionDir, ".getenvoy/extension/examples/another"), `expected to delete example "another"`)
	require.DirExists(t, filepath.Join(extensionDir, ".getenvoy/extension/examples/default"), `expected to not delete example "default"`)
}
