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

package toolchain_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	reference "github.com/tetratelabs/getenvoy/pkg"
	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/fs"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
	toolchains "github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

func TestLoadToolchainFailsIfNotDefault(t *testing.T) {
	workspace, err := workspaces.GetWorkspaceAt("testdata/workspace1")
	require.NoError(t, err)

	_, err = toolchains.LoadToolchain("non-existing", workspace)
	require.EqualError(t, err, `unknown toolchain "non-existing". At the moment, only "default" toolchain is supported`) //nolint:lll
}

func TestLoadToolchainFailsIfUnknown(t *testing.T) {
	workspace, err := workspaces.GetWorkspaceAt("testdata/workspace1")
	require.NoError(t, err)

	_, err = toolchains.LoadToolchain(toolchains.Default, workspace)
	require.EqualError(t, err, fmt.Sprintf(`toolchain "default" has invalid configuration coming from "%s/.getenvoy/extension/toolchains/default.yaml": `+ //nolint:lll
		`unknown toolchain kind "UnknownToolchain"`, workspace.GetDir().GetRootDir()))
}

//nolint:lll
func TestLoadToolchainFailsOnInvalidConfig(t *testing.T) {
	workspace, err := workspaces.GetWorkspaceAt("testdata/workspace2")
	require.NoError(t, err)

	_, err = toolchains.LoadToolchain(toolchains.Default, workspace)
	require.EqualError(t, err, fmt.Sprintf(`toolchain "default" has invalid configuration coming from "%s/.getenvoy/extension/toolchains/default.yaml": `+
		`'build' tool config is not valid: container configuration is not valid: "?invalid value?" is not a valid image name: invalid reference format`, workspace.GetDir().GetRootDir()))
}

func TestLoadToolchain(t *testing.T) {
	workspace, err := workspaces.GetWorkspaceAt("testdata/workspace3")
	require.NoError(t, err)

	builder, err := toolchains.LoadToolchain(toolchains.Default, workspace)
	require.NoError(t, err)

	toolchain, err := builder.Build()
	require.NoError(t, err)
	require.NotNil(t, toolchain)
}

func TestLoadToolchainCreatesDefaultWhenMissing(t *testing.T) {
	extensionDir, removeExtensionDir := morerequire.RequireNewTempDir(t)
	defer removeExtensionDir()

	dir, err := fs.CreateExtensionDir(extensionDir)
	require.NoError(t, err)

	err = dir.WriteFile(model.DescriptorFile, []byte(fmt.Sprintf(`
kind: Extension

name: mycompany.filters.http.custom_metrics

category: envoy.filters.http
language: rust

# Runtime the extension is being developed against.
runtime:
  envoy:
    version: %s
`, reference.Latest)))
	require.NoError(t, err)

	workspace, err := workspaces.GetWorkspaceAt(extensionDir)
	require.NoError(t, err)

	builder, err := toolchains.LoadToolchain(toolchains.Default, workspace)
	require.NoError(t, err)

	toolchain, err := builder.Build()
	require.NoError(t, err)
	require.NotNil(t, toolchain)

	hasDefaultToolchain, err := workspace.HasToolchain(toolchains.Default)
	require.NoError(t, err)
	require.True(t, hasDefaultToolchain)

	file, err := workspace.GetToolchainConfig(toolchains.Default)
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file.Content)
}
