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

package e2e_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	reference "github.com/tetratelabs/getenvoy/pkg"
	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
	toolchains "github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain"
	. "github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

// TestGetEnvoyExtensionInit runs the equivalent of "getenvoy extension init" for a matrix of extension.Categories and
// extension.Languages.
//
// "getenvoy extension init" does not use Docker. See TestMain for general notes on about the test runtime.
func TestGetEnvoyExtensionInit(t *testing.T) {
	const extensionName = "getenvoy_extension_init"

	type testCase struct {
		name string
		extension.Category
		extension.Language
		currentDirectory bool
	}

	tests := make([]testCase, 0)
	for _, cell := range getExtensionTestMatrix() {
		tests = append(tests,
			testCase{cell.String() + "-currentDirectory", cell.Category, cell.Language, true},
			testCase{cell.String() + "-newDirectory", cell.Category, cell.Language, false},
		)
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			t.Parallel() // does not use Docker, so safe to run parallel

			outputDir, removeOutputDir := RequireNewTempDir(t)
			defer removeOutputDir()

			// "getenvoy extension init" should result in stderr describing files created.
			c := getEnvoy("extension init").
				Arg("--category").Arg(test.Category.String()).
				Arg("--language").Arg(test.Language.String()).
				Arg("--name").Arg(extensionName)

			if test.currentDirectory {
				c.WorkingDir(outputDir)
			} else {
				c.Arg(outputDir)
			}

			stderr := requireExecNoStdout(t, c)

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
			require.Equal(t, reference.Latest, descriptor.Runtime.Envoy.Version, `wrong extension envoy version running [%v]: %s`, c, descriptor)

			// Check the default toolchain is loadable
			toolchain, err := toolchains.LoadToolchain(toolchains.Default, workspace)
			require.NoError(t, err, `error loading toolchain running [%v]`, c)
			require.NotNil(t, toolchain, `nil toolchain running [%v]`, c)

			// Verify ignore files didn't end up in the output directory
			for _, ignore := range []string{".gitignore", ".licenserignore"} {
				require.NotContains(t, stderr, fmt.Sprintf("* %s\n", ignore), `ignore file %s found in stderr running [%v]`, ignore, c)
			}

			// Verify language-specific files
			var languageSpecificPaths []string
			switch test.Language {
			case extension.LanguageRust:
				languageSpecificPaths = []string{
					".cargo/config.toml",
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
