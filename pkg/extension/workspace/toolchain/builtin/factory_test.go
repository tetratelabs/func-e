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
package builtin

import (
	"testing"

	"github.com/stretchr/testify/require"

	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config"
	extensionconfig "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
	builtinconfig "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/toolchain/builtin"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain/registry"
)

func TestBuiltinToolchainLoadConfig(t *testing.T) {
	workspace, err := workspaces.GetWorkspaceAt("testdata/workspace")
	require.NoError(t, err)
	tests := []struct {
		name     string
		config   string
		expected string
	}{
		{
			name:   "empty config",
			config: `kind: BuiltinToolchain`,
			expected: `
            kind: BuiltinToolchain
            container:
              image: getenvoy/extension-rust-builder:latest
            build:
              output:
                wasmFile: target/getenvoy/extension.wasm
`,
		},
		{
			name: "example config",
			config: string(ExampleConfig(&extensionconfig.Descriptor{
				Language: extensionconfig.LanguageRust,
			})),
			expected: `
            kind: BuiltinToolchain
            container:
              image: getenvoy/extension-rust-builder:latest
            build:
              output:
                wasmFile: target/getenvoy/extension.wasm
`,
		},
		{
			name: "full config",
			config: `
            kind: BuiltinToolchain
            container:
              image: tetratelabs/getenvoy-extension-rust-builder:1.2.3
            build:
              container:
                image: tetratelabs/getenvoy-extension-rust-builder:4.5.6
                options:
                - -e
                - VAR=ALUE
              output:
                wasmFile: target/extension.wasm
            test:
              container:
                image: tetratelabs/getenvoy-extension-rust-builder:7.8.9
                options:
                - -v
                - /host:/container
`,
			expected: `
            kind: BuiltinToolchain
            container:
              image: tetratelabs/getenvoy-extension-rust-builder:1.2.3
            build:
              container:
                image: tetratelabs/getenvoy-extension-rust-builder:4.5.6
                options:
                - -e
                - VAR=ALUE
              output:
                wasmFile: target/extension.wasm
            test:
              container:
                image: tetratelabs/getenvoy-extension-rust-builder:7.8.9
                options:
                - -v
                - /host:/container
`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			// verify built-in toolchain is registered
			factory, exists := registry.Get(builtinconfig.Kind)
			require.True(t, exists)

			// load toolchain config
			builder, err := factory.LoadConfig(registry.LoadConfigArgs{
				Workspace: workspace,
				Toolchain: registry.ToolchainConfig{
					Name: "example",
					Config: &model.File{
						Source:  "<memory>",
						Content: []byte(test.config),
					},
				},
			})
			require.NoError(t, err)

			// verify defaults get applied to the toolchain config
			actual, err := config.Marshal(builder.GetConfig())
			require.NoError(t, err)
			require.YAMLEq(t, test.expected, string(actual))

			// create a toolchain
			toolchain, err := builder.Build()
			require.NoError(t, err)
			require.NotNil(t, toolchain)
		})
	}
}
