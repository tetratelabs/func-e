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

	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
	toolchains "github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain"
)

// TestGetEnvoyExtensionInit runs the equivalent of "getenvoy extension init" for a matrix of extension.Categories and
// extension.Languages.
//
// "getenvoy extension init" does not use Docker. See TestMain for general notes on about the test runtime.
func TestGetEnvoyExtensionInit(t *testing.T) {
	const extensionName = "getenvoy_extension_init"

	type testTuple struct {
		testName string
		extension.Category
		extension.Language
		currentDirectory bool
	}

	tests := make([]testTuple, 0)
	for _, c := range getExtensionTestMatrix() {
		tests = append(tests,
			testTuple{c.String() + "-currentDirectory", c.Category, c.Language, true},
			testTuple{c.String() + "-newDirectory", c.Category, c.Language, false},
		)
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.testName, func(t *testing.T) {
			workDir, removeWorkDir := requireNewTempDir(t)
			defer removeWorkDir()

			revertChDir := requireChDir(t, workDir)
			defer revertChDir()

			if !test.currentDirectory {
				workDir = filepath.Join(workDir, "newDirectory")
			}

			// "getenvoy extension init" should result in stderr describing files created.
			cmd := getEnvoy("extension init").
				Arg(workDir).
				Arg("--category").Arg(test.Category.String()).
				Arg("--language").Arg(test.Language.String()).
				Arg("--name").Arg(extensionName)
			stderr := requireExecNoStdout(t, cmd)

			// Check that the contents look valid for the inputs.
			for _, regex := range []string{
				`^\QScaffolding a new extension:\E\n`,
				fmt.Sprintf(`\QGenerating files in %s:\E\n`, workDir),
				`\Q* .getenvoy/extension/extension.yaml\E\n`,
				`\QDone!\E\n$`,
			} {
				require.Regexp(t, regex, stderr, `invalid stderr running [%v]`, cmd)
			}

			// Check to see that the extension.yaml mentioned in stderr exists.
			// Note: we don't check all files as extensions are language-specific.
			require.FileExists(t, filepath.Join(workDir, ".getenvoy/extension/extension.yaml"), `extension.yaml missing after running [%v]`, cmd)

			// Check the generated extension.yaml includes values we passed and includes the default toolchain.
			workspace, err := workspaces.GetWorkspaceAt(workDir)
			require.NoError(t, err, `error getting workspace after running [%v]`, cmd)
			require.NotNil(t, workspace, `nil workspace running [%v]`, cmd)
			require.Equal(t, extensionName, workspace.GetExtensionDescriptor().Name, `wrong extension name running [%v]`, cmd)
			require.Equal(t, test.Category, workspace.GetExtensionDescriptor().Category, `wrong extension category running [%v]`, cmd)
			require.Equal(t, test.Language, workspace.GetExtensionDescriptor().Language, `wrong extension language running [%v]`, cmd)

			// Check the default toolchain is loadable
			toolchain, err := toolchains.LoadToolchain(toolchains.Default, workspace)
			require.NoError(t, err, `error loading toolchain running [%v]`, cmd)
			require.NotNil(t, toolchain, `nil toolchain running [%v]`, cmd)
		})
	}
}
