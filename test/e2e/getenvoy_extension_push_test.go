// Copyright 2021 Tetrate
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
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/extension/wasmimage"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
)

// TestGetEnvoyExtensionPush runs the equivalent of "getenvoy extension push". "getenvoy extension init" and
// "getenvoy extension build" are a prerequisites, so run first.
//
// This test does not attempt to use the image built as that would be redundant to other tests. Rather, this focuses on
// whether we can read back exactly what was pushed to the registry.
//
// "getenvoy extension run" depends on a Docker container, and "getenvoy extension build" uses Docker.
// See TestMain for general notes on about the test runtime.
func TestGetEnvoyExtensionPush(t *testing.T) {
	const extensionName = "getenvoy_extension_push"
	// localRegistryWasmImageRef corresponds to a Docker container running the image "registry:2"
	const localRegistryWasmImageRef = "localhost:5000/getenvoy/" + extensionName
	// When unspecified, we default the tag to Docker's default "latest". Note: recent tools enforce qualifying this!
	const defaultTag = "latest"

	type testTuple struct {
		name string
		extension.Category
		extension.Language
	}

	// Push is not language-specific, so we don't need to test a large matrix, and doing so would slow down e2e runtime.
	// Instead, we choose something that executes "getenvoy extension build" quickly.
	tests := []testTuple{
		{"tinygo HTTP filter", extension.EnvoyHTTPFilter, extension.LanguageTinyGo},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			workDir, removeWorkDir := requireNewTempDir(t)
			defer removeWorkDir()

			revertChDir := requireChDir(t, workDir)
			defer revertChDir()

			// push requires "get envoy extension init" and "get envoy extension build" to have succeeded
			requireExtensionInit(t, workDir, test.Category, test.Language, extensionName)
			defer requireExtensionClean(t, workDir)
			wasmBytes := requireExtensionBuild(t, test.Language, workDir)

			// After pushing, stderr should include the registry URL and the image tag.
			cmd := getEnvoy("extension push").Arg(localRegistryWasmImageRef)
			stderr := requireExecNoStdout(t, cmd)

			// Assemble a fully-qualified image ref as we'll pull this later
			imageRef := localRegistryWasmImageRef + ":" + defaultTag

			// Verify stderr shows the latest tag and the correct image ref
			require.Contains(t, stderr, fmt.Sprintf(`Using default tag: %s
Pushed %s
digest: sha256`, defaultTag, imageRef), `unexpected stderr after running [%v]`, cmd)

			// Get a puller we can use to pull what we just pushed.
			puller, err := wasmimage.NewPuller(false, false)
			require.NoError(t, err, `error getting puller instance after running [%v]`, cmd)
			require.NotNil(t, puller, `nil puller instance after running [%v]`, cmd)

			// Pull the wasm we just pushed, writing it to a local file.
			dstPath := filepath.Join(workDir, "pulled_extension.wasm")
			desc, err := puller.Pull(imageRef, dstPath)
			require.NoError(t, err, `error pulling wasm after running [%v]: %s`, cmd)

			// Verify the pulled image descriptor is valid and the image file exists/
			require.Equal(t, "application/vnd.module.wasm.content.layer.v1+wasm", desc.MediaType, `invalid media type after running [%v]`, cmd)
			require.Equal(t, "extension.wasm", desc.Annotations["org.opencontainers.image.title"], `invalid image title after running [%v]`, cmd)
			require.FileExists(t, dstPath, `image not written after running [%v]`, cmd)

			// Verify the bytes pulled are exactly the same as what we pushed.
			pulledBytes, err := ioutil.ReadFile(dstPath)
			require.NoError(t, err, `error reading file wasm %s after running [%v]`, dstPath, cmd)
			require.NotEmpty(t, wasmBytes, `%s empty after running [%v]`, dstPath, cmd)
			require.Equal(t, wasmBytes, pulledBytes, `pulled bytes don't match source after running [%v]`, cmd)
		})
	}
}
