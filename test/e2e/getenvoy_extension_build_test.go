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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	e2e "github.com/tetratelabs/getenvoy/test/e2e/util"
)

// TestGetEnvoyExtensionBuild runs the equivalent of "getenvoy extension build" for a matrix of extension.Categories and
// extension.Languages. "getenvoy extension init" is a prerequisite, so run first.
//
// Note: "getenvoy extension build" can be extremely slow due to implicit responsibilities such as downloading modules
// or compilation. This uses Docker, so changes to the Dockerfile or contents like "commands.sh" effect performance.
//
// Note: Pay close attention to values of util.E2E_BUILTIN_TOOLCHAIN_CONTAINER_OPTIONS as these can change assumptions.
// CI may override this to set HOME or CARGO_HOME (rust) used by "getenvoy" and effect its execution.
func TestGetEnvoyExtensionBuild(t *testing.T) {
	const extensionName = "getenvoy_extension_build"
	requireEnvoyBinaryPath(t) // Ex. After running "make bin", E2E_GETENVOY_BINARY=$PWD/build/bin/darwin/amd64/getenvoy

	for _, test := range e2e.GetCategoryLanguageCombinations() {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.String(), func(t *testing.T) {
			workDir, removeWorkDir := requireNewTempDir(t)
			defer removeWorkDir()

			revertChDir := requireChDir(t, workDir)
			defer revertChDir()

			// test requires "get envoy extension init" to have succeeded
			requireExtensionInit(t, workDir, test.Category, test.Language, extensionName)
			defer requireExtensionClean(t, workDir)

			// "getenvoy extension build" only returns stdout because `docker run -t` redirects stderr to stdout.
			// We don't verify stdout because it is low signal vs looking at files created.
			cmd := GetEnvoy("extension build").Args(e2e.Env.GetBuiltinContainerOptions()...)
			_ = requireExecNoStderr(t, cmd)

			// Verify the extension built
			extensionWasmFile := filepath.Join(workDir, extensionWasmPath(test.Language))
			require.FileExists(t, extensionWasmFile, `extension wasm file %s missing after running [%v]`, extensionWasmFile, cmd)
		})
	}
}
