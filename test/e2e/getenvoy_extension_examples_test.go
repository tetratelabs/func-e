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

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
	e2e "github.com/tetratelabs/getenvoy/test/e2e/util"
)

// TestGetEnvoyExtensionExampleAdd runs the equivalent of "getenvoy extension example XXX" commands for a matrix of
// extension.Categories and extension.Languages.
//
// "getenvoy extension example XXX" should be fast and reliable because they don't use Docker.
func TestGetEnvoyExtensionExample(t *testing.T) {
	const extensionName = "getenvoy_extension_example"
	requireEnvoyBinaryPath(t) // Ex. After running "make bin", E2E_GETENVOY_BINARY=$PWD/build/bin/darwin/amd64/getenvoy

	for _, test := range e2e.GetCategoryLanguageCombinations() {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.String(), func(t *testing.T) {
			var extensionConfigFileName string
			switch test.Language {
			case extension.LanguageTinyGo:
				extensionConfigFileName = "extension.txt"
			default:
				extensionConfigFileName = "extension.json"
			}

			workDir, removeWorkDir := requireNewTempDir(t)
			defer removeWorkDir()

			revertChDir := requireChDir(t, workDir)
			defer revertChDir()

			// "getenvoy extension example XXX" commands require an extension init to succeed
			requireExtensionInit(t, workDir, test.Category, test.Language, extensionName)
			defer requireExtensionClean(t, workDir)

			// "getenvoy extension examples list" should start empty
			cmd := GetEnvoy("extension examples list")
			stderr := requireExecNoStdout(t, cmd)
			require.Equal(t, `Extension has no example setups.

Use "getenvoy extension examples add --help" for more information on how to add one.
`, stderr, `invalid stderr running [%v]`, cmd)

			// "getenvoy extension examples add" should result in stderr describing files created.
			cmd = GetEnvoy("extension examples add")
			stderr = requireExecNoStdout(t, cmd)

			exampleFiles := []string{
				filepath.Join(workDir, ".getenvoy/extension/examples/default/README.md"),
				filepath.Join(workDir, ".getenvoy/extension/examples/default/envoy.tmpl.yaml"),
				filepath.Join(workDir, ".getenvoy/extension/examples/default/example.yaml"),
				fmt.Sprintf(".getenvoy/extension/examples/default/%s", extensionConfigFileName),
			}

			exampleFileText := fmt.Sprintf(`
* .getenvoy/extension/examples/default/README.md
* .getenvoy/extension/examples/default/envoy.tmpl.yaml
* .getenvoy/extension/examples/default/example.yaml
* .getenvoy/extension/examples/default/%s
`, extensionConfigFileName)

			// Check stderr mentions the files created
			require.Equal(t, fmt.Sprintf("Scaffolding a new example setup:%sDone!\n", exampleFileText),
				stderr, `invalid stderr running [%v]`, cmd)

			// Check the files mentioned actually exist
			for _, path := range exampleFiles {
				require.FileExists(t, path, `example file %s missing after running [%v]`, path, cmd)
			}

			// "getenvoy extension examples list" should now include an example
			cmd = GetEnvoy("extension examples list")
			stdout := requireExecNoStderr(t, cmd)
			require.Equal(t, "EXAMPLE\ndefault\n", stdout, `invalid stdout running [%v]`, cmd)

			// "getenvoy extension examples add" should result in stderr describing files created.
			cmd = GetEnvoy("extension examples remove --name default")
			stderr = requireExecNoStdout(t, cmd)

			// Check stderr mentions the files removed
			require.Equal(t, fmt.Sprintf("Removing example setup:%sDone!\n", exampleFileText),
				stderr, `invalid stderr running [%v]`, cmd)

			// Check the files mentioned actually were removed
			for _, path := range exampleFiles {
				require.NoFileExists(t, path, `example file %s still exists after running [%v]`, path, cmd)
			}
		})
	}
}
