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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	rootcmd "github.com/tetratelabs/getenvoy/pkg/cmd"
	"github.com/tetratelabs/getenvoy/pkg/globals"
	"github.com/tetratelabs/getenvoy/pkg/test/cmd"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

func TestGetEnvoyExtensionExamplesAddValidateFlag(t *testing.T) {
	type testCase struct {
		name        string
		args        []string
		expectedErr string
	}

	tests := []testCase{
		{
			name:        "--name with invalid value",
			args:        []string{"extension", "examples", "add", "--name", "my:example"},
			expectedErr: `"my:example" is not a valid example name. Example name must match the format "^[a-z0-9._-]+$". E.g., 'my.example', 'my-example' or 'my_example'`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			// Run "getenvoy extension examples add" with the args we are testing
			c, stdout, stderr := cmd.NewRootCommand(&globals.GlobalOpts{})
			c.SetArgs(test.args)

			err := rootcmd.Execute(c)
			require.EqualError(t, err, test.expectedErr, `expected an error running [%v]`, c)

			// Verify the command failed with the expected error
			require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
			expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy extension examples add --help' for usage.\n", test.expectedErr)
			require.Equal(t, expectedStderr, stderr.String(), `unexpected stderr running [%v]`, c)
		})
	}
}

func TestGetEnvoyExtensionExamplesAddFailsOutsideExtensionDirectory(t *testing.T) {
	// Change to a non-workspace dir
	o := &globals.GlobalOpts{ExtensionDir: morerequire.RequireAbs(t, relativeRustExtensionDirWithOneExample+"/..")}

	// Run "getenvoy extension examples add"
	c, stdout, stderr := cmd.NewRootCommand(o)
	c.SetArgs([]string{"extension", "examples", "add"})
	err := rootcmd.Execute(c)

	// Verify the command failed with the expected error
	expectedErr := fmt.Sprintf("not an extension directory %q", o.ExtensionDir)
	require.EqualError(t, err, expectedErr, `expected an error running [%v]`, c)
	require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
	expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy extension examples add --help' for usage.\n", expectedErr)
	require.Equal(t, expectedStderr, stderr.String(), `unexpected stderr running [%v]`, c)
}

func TestGetEnvoyExtensionExamplesAddFailsWhenExampleExists(t *testing.T) {
	o := &globals.GlobalOpts{ExtensionDir: morerequire.RequireAbs(t, relativeRustExtensionDirWithOneExample)}

	// Run "getenvoy extension examples add"
	c, stdout, stderr := cmd.NewRootCommand(o)
	c.SetArgs([]string{"extension", "examples", "add"})
	err := rootcmd.Execute(c)

	// Verify the command failed with the expected error
	expectedErr := `example setup "default" already exists`
	require.EqualError(t, err, expectedErr, `expected an error running [%v]`, c)
	require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
	expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy extension examples add --help' for usage.\n", expectedErr)
	require.Equal(t, expectedStderr, stderr.String(), `unexpected stderr running [%v]`, c)
}

func TestGetEnvoyExtensionExamplesAdd(t *testing.T) {
	type testCase struct {
		name                   string
		templateWorkspace      string
		args                   []string
		expectedName           string
		expectedConfigFileName string
	}

	tests := []testCase{
		{
			name:                   `rust workspace`,
			templateWorkspace:      relativeRustExtensionDirWithNoExample,
			args:                   []string{"--name", "test-example"},
			expectedName:           `test-example`,
			expectedConfigFileName: `extension.json`,
		},
		{
			name:                   `tinygo workspace`,
			templateWorkspace:      relativeTinyGoExtensionDirWithNoExample,
			args:                   []string{"--name", "test-example"},
			expectedName:           `test-example`,
			expectedConfigFileName: `extension.txt`,
		},
		{
			name:                   `--name defaults to "default"`,
			templateWorkspace:      relativeTinyGoExtensionDirWithNoExample,
			args:                   []string{},
			expectedName:           `default`,
			expectedConfigFileName: `extension.txt`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			// Copy the workspace as this test will delete it, and we don't want to mutate our test data!
			extensionDir, removeExtensionDir := morerequire.RequireCopyOfDir(t, test.templateWorkspace)
			defer removeExtensionDir()

			// "getenvoy extension examples add" must be in a valid extension directory
			o := &globals.GlobalOpts{ExtensionDir: extensionDir}

			// Run "getenvoy extension examples add"
			c, stdout, stderr := cmd.NewRootCommand(o)
			c.SetArgs(append([]string{"extension", "examples", "add"}, test.args...))
			err := rootcmd.Execute(c)

			exampleDir := filepath.Join(`.getenvoy`, `extension`, `examples`, test.expectedName)

			// Verify the files created end up in stderr
			require.NoError(t, err, `expected no error running [%v]`, c)
			require.Empty(t, stdout.String(), `expected no stderr running [%v]`, c)
			require.Equal(t, fmt.Sprintf(`Scaffolding a new example setup:
* %[1]s/README.md
* %[1]s/envoy.tmpl.yaml
* %[1]s/example.yaml
* %[1]s/%[2]s
Done!
`, exampleDir, test.expectedConfigFileName), stderr.String(), `unexpected stderr running [%v]`, c)

			// Get the absolute path
			exampleDir = filepath.Join(extensionDir, exampleDir)

			// Verify the files actually exist.
			for _, p := range []string{
				`README.md`, `envoy.tmpl.yaml`, `example.yaml`, test.expectedConfigFileName,
			} {
				require.FileExists(t, filepath.Join(exampleDir, p), `expected to find %s in %s`, p, exampleDir)
			}

			// Check README substitution: ${EXTENSION_CONFIG_FILE_NAME} must be replaced with "extension.json".
			readmePath := filepath.Join(exampleDir, `README.md`)
			data, err := os.ReadFile(readmePath)
			require.NoError(t, err, `expected no error reading README.md file: %s`, readmePath)

			// Check one variable, noting that EXTENSION_CONFIG_FILE_NAME is language specific.
			readme := string(data)
			require.NotContains(t, readme, `EXTENSION_CONFIG_FILE_NAME`, `expected variables to be replaced in %s`, readmePath)
			require.Contains(t, readme, test.expectedConfigFileName, `expected to see config file %s in %s`, test.expectedConfigFileName, readmePath)
		})
	}
}
