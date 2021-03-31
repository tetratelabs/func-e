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

package push_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/test/cmd/extension"
	cmdutil "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

// relativeWorkspaceDir points to a usable pre-initialized workspace
const relativeWorkspaceDir = "testdata/workspace"

// localRegistryWasmImageRef corresponds to a Docker container running the image "registry:2"
// As this is not intended to be an end-to-end test, this could be improved to use a mock/fake HTTP registry instead.
const localRegistryWasmImageRef = "localhost:5000/getenvoy/sample"

// When unspecified, we default the tag to Docker's default "latest". Note: recent tools enforce qualifying this!
const defaultTag = "latest"

// TestGetEnvoyExtensionPush shows current directory is usable, provided it is a valid workspace.
func TestGetEnvoyExtensionPush(t *testing.T) {
	_, revertWd := extension.RequireChDir(t, relativeWorkspaceDir)
	defer revertWd()

	// Run "getenvoy extension push localhost:5000/getenvoy/sample"
	cmd, stdout, stderr := extension.NewRootCommand()
	cmd.SetArgs([]string{"extension", "push", localRegistryWasmImageRef})
	err := cmdutil.Execute(cmd)

	// A fully qualified image ref includes the tag
	imageRef := localRegistryWasmImageRef + ":" + defaultTag

	// Verify stdout shows the latest tag and the correct image ref
	require.NoError(t, err, `expected no error running [%v]`, cmd)

	require.Contains(t, stdout.String(), fmt.Sprintf(`Using default tag: %s
Pushed %s
digest: sha256`, defaultTag, imageRef), `unexpected stderr after running [%v]`, cmd)
	require.Empty(t, stderr.String(), `expected no stderr running [%v]`, cmd)
}

func TestGetEnvoyExtensionPushFailsOutsideWorkspaceDirectory(t *testing.T) {
	// Change to a non-workspace dir
	dir, revertWd := extension.RequireChDir(t, relativeWorkspaceDir+"/..")
	defer revertWd()

	// Run "getenvoy extension push localhost:5000/getenvoy/sample"
	cmd, stdout, stderr := extension.NewRootCommand()
	cmd.SetArgs([]string{"extension", "push", localRegistryWasmImageRef})
	err := cmdutil.Execute(cmd)

	// Verify the command failed with the expected error
	expectedErr := "there is no extension directory at or above: " + dir
	require.EqualError(t, err, expectedErr, `expected an error running [%v]`, cmd)
	require.Empty(t, stdout.String(), `expected no stdout running [%v]`, cmd)
	expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy extension push --help' for usage.\n", expectedErr)
	require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, cmd)
}

// TestGetEnvoyExtensionPushWithExplicitFileOption shows
func TestGetEnvoyExtensionPushWithExplicitFileOption(t *testing.T) {
	// Change to a non-workspace dir
	dir, revertWd := extension.RequireChDir(t, relativeWorkspaceDir+"/..")
	defer revertWd()

	// Point to a wasm file explicitly
	wasm := filepath.Join(dir, "workspace", "extension.wasm")

	// Run "getenvoy extension push localhost:5000/getenvoy/sample --extension-file testdata/workspace/extension.wasm"
	cmd, stdout, stderr := extension.NewRootCommand()
	cmd.SetArgs([]string{"extension", "push", localRegistryWasmImageRef, "--extension-file", wasm})
	err := cmdutil.Execute(cmd)

	// Verify the pushed a latest tag to the correct registry
	require.NoError(t, err, `expected no error running [%v]`, cmd)
	require.Contains(t, stdout.String(), fmt.Sprintf(`Using default tag: latest
Pushed %s:latest
digest: sha256`, localRegistryWasmImageRef))
	require.Empty(t, stderr.String(), `expected no stderr running [%v]`, cmd)
}
