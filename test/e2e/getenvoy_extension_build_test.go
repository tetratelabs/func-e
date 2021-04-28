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

	. "github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

// TestGetEnvoyExtensionBuild runs the equivalent of "getenvoy extension build" for a matrix of extension.Categories and
// extension.Languages. "getenvoy extension init" is a prerequisite, so run first.
//
// "getenvoy extension build" uses Docker. See TestMain for general notes on about the test runtime.
func TestGetEnvoyExtensionBuild(t *testing.T) {
	const extensionName = "getenvoy_extension_build"

	for _, test := range getExtensionTestMatrix() {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.String(), func(t *testing.T) {
			// TODO: uses Docker, but may be possible to run parallel
			extensionDir, removeExtensionDir := RequireNewTempDir(t)
			defer removeExtensionDir()

			// test requires "get envoy extension init" to have succeeded
			requireExtensionInit(t, extensionDir, test.Category, test.Language, extensionName)
			defer requireExtensionClean(t, extensionDir)

			// "getenvoy extension build" only returns stdout because `docker run -t` redirects stderr to stdout.
			// We don't verify stdout because it is low signal vs looking at files created.
			c := getEnvoy("extension build").Args(getToolchainContainerOptions()...).WorkingDir(extensionDir)
			_ = requireExecNoStderr(t, c)

			// Verify the extension built
			extensionWasmFile := filepath.Join(extensionDir, extensionWasmPath(test.Language))
			require.FileExists(t, extensionWasmFile, `extension wasm file %s missing after running [%v]`, extensionWasmFile, c)
		})
	}
}
