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
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/tetratelabs/getenvoy/pkg/test/morerequire"
	"github.com/tetratelabs/getenvoy/pkg/wasm"
)

// TestGetEnvoyExtensionPush runs the equivalent of "getenvoy extension push"
//
// This test does not attempt to use a wasm image. Rather, this focuses on whether we can read back exactly what was
// pushed to the registry.
//
// See TestMain for general notes on about the test runtime.
func TestGetEnvoyExtensionPush(t *testing.T) {
	const extensionName = "getenvoy_extension_push"
	// localRegistryWasmImageRef corresponds to a Docker container running the image "registry:2"
	const localRegistryWasmImageRef = "localhost:5000/getenvoy/" + extensionName
	// The above registry is HTTP not HTTPS
	const useHTTP = true
	// When unspecified, we default the tag to Docker's default "latest". Note: recent tools enforce qualifying this!
	const defaultTag = "latest"

	t.Parallel() // tinygo is safe to run parallel even with Docker

	wasmFile := RequireAbs(t, "testdata/extension.wasm")
	wasmBytes, err := ioutil.ReadFile(wasmFile)
	require.NoError(t, err)

	// After pushing, stderr should include the registry URL and the image tag.
	c := getEnvoy("extension push").
		Arg(wasmFile).Arg(localRegistryWasmImageRef).Arg("--use-http").Arg(strconv.FormatBool(useHTTP))
	stderr := requireExecNoStdout(t, c)

	// Assemble a fully-qualified image ref as we'll pull this later
	imageRef := localRegistryWasmImageRef + ":" + defaultTag

	// Verify stderr shows the latest tag and the correct image ref
	require.Contains(t, stderr, fmt.Sprintf(`Using default tag: %s
Pushed %s
digest: sha256`, defaultTag, imageRef), `unexpected stderr after running [%v]`, c)

	// Get a puller we can use to pull what we just pushed.
	puller, err := wasm.NewPuller(false, useHTTP)
	require.NoError(t, err, `error getting puller instance after running [%v]`, c)
	require.NotNil(t, puller, `nil puller instance after running [%v]`, c)

	// Pull the wasm we just pushed, writing it to a local file.
	tempDir, deleteTempDir := RequireNewTempDir(t)
	defer deleteTempDir()

	dstPath := filepath.Join(tempDir, "pulled_extension.wasm")
	desc, err := puller.Pull(imageRef, dstPath)
	require.NoError(t, err, `error pulling wasm after running [%v]: %s`, c)

	// Verify the pulled image descriptor is valid and the image file exists/
	require.Equal(t, "application/vnd.module.wasm.content.layer.v1+wasm", desc.MediaType, `invalid media type after running [%v]`, c)
	require.Equal(t, "extension.wasm", desc.Annotations["org.opencontainers.image.title"], `invalid image title after running [%v]`, c)
	require.FileExists(t, dstPath, `image not written after running [%v]`, c)

	// Verify the bytes pulled are exactly the same as what we pushed.
	pulledBytes, err := os.ReadFile(dstPath)
	require.NoError(t, err, `error reading file wasm %s after running [%v]`, dstPath, c)
	require.NotEmpty(t, wasmBytes, `%s empty after running [%v]`, dstPath, c)
	require.Equal(t, wasmBytes, pulledBytes, `pulled bytes don't match source after running [%v]`, c)
}
