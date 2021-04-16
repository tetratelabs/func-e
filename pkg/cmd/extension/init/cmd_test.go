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

package init_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
	toolchains "github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain"
	"github.com/tetratelabs/getenvoy/pkg/test/cmd"
	. "github.com/tetratelabs/getenvoy/pkg/test/morerequire"
	cmdutil "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

func TestGetEnvoyExtensionInitValidateFlag(t *testing.T) {
	type testCase struct {
		name        string
		args        []string
		expectedErr string
	}

	cwd, err := os.Getwd()
	require.NoError(t, err, "error getting current working directory")

	outputDir, revertOutputDir := RequireNewTempDir(t)
	defer revertOutputDir()

	tests := []testCase{
		{
			name:        "extension category is missing",
			args:        []string{},
			expectedErr: `extension category cannot be empty`,
		},
		{
			name:        "extension category with invalid value",
			args:        []string{"--category", "invalid.category"},
			expectedErr: `"invalid.category" is not a supported extension category`,
		},
		{
			name:        "programming language is missing",
			args:        []string{"--category", "envoy.filters.http"},
			expectedErr: `programming language cannot be empty`,
		},
		{
			name:        "programming language with invalid value",
			args:        []string{"--category", "envoy.filters.http", "--language", "invalid.language"},
			expectedErr: `"invalid.language" is not a supported programming language`,
		},
		{
			name:        "output directory exists but is not empty",
			args:        []string{"--category", "envoy.filters.http", "--language", "tinygo"},
			expectedErr: fmt.Sprintf(`output directory must be empty or new: %s`, cwd),
		},
		{
			name:        "extension name is missing",
			args:        []string{"--category", "envoy.filters.http", "--language", "tinygo", outputDir},
			expectedErr: `extension name cannot be empty`,
		},
		{
			name:        "extension name with invalid value",
			args:        []string{"--category", "envoy.filters.http", "--language", "tinygo", "--name", "?!", outputDir},
			expectedErr: `"?!" is not a valid extension name. Extension name must match the format "^[a-z0-9_]+(\\.[a-z0-9_]+)*$". E.g., 'mycompany.filters.http.custom_metrics'`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			// Run "getenvoy extension init" with the flags we are testing
			c, stdout, stderr := cmd.NewRootCommand()
			c.SetArgs(append([]string{"extension", "init", "--no-prompt", "--no-colors"}, test.args...))
			err := cmdutil.Execute(c)
			require.EqualError(t, err, test.expectedErr, `expected an error running [%v]`, c)

			// Verify the command failed with the expected error
			require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
			expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy extension init --help' for usage.\n", test.expectedErr)
			require.Equal(t, expectedStderr, stderr.String(), `unexpected stderr running [%v]`, c)
		})
	}
}

func TestGetEnvoyExtensionInit(t *testing.T) {
	const extensionName = "getenvoy_extension_init"
	const envoyVersion = "standard:1.17.1"

	type testCase struct {
		name string
		extension.Category
		extension.Language
		currentDirectory bool
	}

	var tests []testCase
	for _, c := range extension.Categories {
		for _, l := range extension.Languages {
			name := fmt.Sprintf(`category=%s language=%s`, c, l)
			tests = append(tests,
				testCase{name + "-currentDirectory", c, l, true},
				testCase{name + "-newDirectory", c, l, false},
			)
		}
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			outputDir, revertOutputDir := RequireNewTempDir(t)
			defer revertOutputDir()

			// Run "getenvoy extension init"
			c, stdout, stderr := cmd.NewRootCommand()
			args := []string{"extension", "init", "--no-prompt", "--no-colors",
				"--category", test.Category.String(),
				"--language", test.Language.String(),
				"--name", extensionName,
			}

			if test.currentDirectory {
				_, revertChDir := RequireChDir(t, outputDir)
				defer revertChDir()
			} else {
				args = append(args, outputDir)
			}

			c.SetArgs(args)
			err := cmdutil.Execute(c)

			require.NoError(t, err, `expected no error running [%v]`, c)
			require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
			// Check that the contents look valid for the inputs.
			for _, regex := range []string{
				`^\QScaffolding a new extension:\E\n`,
				fmt.Sprintf(`\QGenerating files in %s:\E\n`, outputDir),
				`\Q* .getenvoy/extension/extension.yaml\E\n`,
				`\QDone!\E\n$`,
			} {
				require.Regexp(t, regex, stderr, `invalid stderr running [%v]`, c)
			}

			// Check to see that the extension.yaml mentioned in stderr exists.
			// Note: we don't check all files as extensions are language-specific.
			require.FileExists(t, filepath.Join(outputDir, ".getenvoy/extension/extension.yaml"), `extension.yaml missing after running [%v]`, c)

			// Check the generated extension.yaml includes values we passed and includes the default toolchain.
			workspace, err := workspaces.GetWorkspaceAt(outputDir)
			require.NoError(t, err, `error getting workspace after running [%v]`, c)
			require.NotNil(t, workspace, `nil workspace running [%v]`, c)
			descriptor := workspace.GetExtensionDescriptor()
			require.Equal(t, extensionName, descriptor.Name, `wrong extension name running [%v]: %s`, c, descriptor)
			require.Equal(t, test.Category, descriptor.Category, `wrong extension category running [%v]: %s`, c, descriptor)
			require.Equal(t, test.Language, descriptor.Language, `wrong extension language running [%v]: %s`, c, descriptor)
			require.Equal(t, envoyVersion, descriptor.Runtime.Envoy.Version, `wrong extension envoy version running [%v]: %s`, c, descriptor)

			// Check the default toolchain is loadable
			toolchain, err := toolchains.LoadToolchain(toolchains.Default, workspace)
			require.NoError(t, err, `error loading toolchain running [%v]`, c)
			require.NotNil(t, toolchain, `nil toolchain running [%v]`, c)

			// Verify ignore files didn't end up in the output directory
			for _, ignore := range []string{".gitignore", ".licenserignore"} {
				require.NotContains(t, stderr.String(), fmt.Sprintf("* %s\n", ignore), `ignore file %s found in stderr running [%v]`, ignore, c)
			}

			// Verify language-specific files
			var languageSpecificPaths []string
			switch test.Language {
			case extension.LanguageRust:
				languageSpecificPaths = []string{
					".cargo/config",
					"Cargo.toml",
					"README.md",
					"src/config.rs",
					"src/lib.rs",
					"wasm/module/Cargo.toml",
					"wasm/module/src/lib.rs",
				}

			case extension.LanguageTinyGo:
				languageSpecificPaths = []string{
					"go.mod",
					"go.sum",
					"main.go",
					"main_test.go",
				}
			}

			// Verify the paths were in stderr and actually exist.
			for _, f := range languageSpecificPaths {
				require.Regexp(t, fmt.Sprintf(`\Q* %s\E\n`, f), stderr, `expected stderr to include %s running [%v]`, f, c)
				require.FileExists(t, filepath.Join(outputDir, f), `%s missing after running [%v]`, f, c)
			}
		})
	}
}
