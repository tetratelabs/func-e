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

package builtin_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config"
	. "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/toolchain/builtin"
)

func TestToolchainConfigValidate(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "default build container",
			input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
`,
		},
		{
			name: "empty build config",
			input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    build: {}
`,
		},
		{
			name: "'build' config with container",
			input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    build:
                      container:
                        image: build/image
`,
		},
		{
			name: "'build' config with *.wasm file output path",
			input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    build:
                      container:
                        image: build/image
                      output:
                        wasmFile: output/extension.wasm
`,
		},
		{
			name: "'test' config with container",
			input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    test:
                      container:
                        image: test/image
`,
		},
		{
			name: "'clean' config with container",
			input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    clean:
                      container:
                        image: clean/image
`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			var toolchainConfig ToolchainConfig
			err := config.Unmarshal([]byte(test.input), &toolchainConfig)
			require.NoError(t, err)

			err = toolchainConfig.Validate()
			require.NoError(t, err)
		})
	}
}

func TestToolchainConfigValidateError(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name:        "blank",
			input:       ``,
			expectedErr: `configuration of the default build container cannot be empty`,
		},
		{
			name: "no default build container",
			input: `
                    kind: BuiltinToolchain
`,
			expectedErr: `configuration of the default build container cannot be empty`,
		},
		{
			name: "default build container: no image name",
			input: `
                    kind: BuiltinToolchain
                    container: {}
`,
			expectedErr: `configuration of the default build container is not valid: image name cannot be empty`,
		},
		{
			name: "default build container: invalid image name",
			input: `
                    kind: BuiltinToolchain
                    container:
                      image: ???
`,
			expectedErr: `configuration of the default build container is not valid: "???" is not a valid image name: invalid reference format`,
		},
		{
			name: "build tool: no image name",
			input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    build:
                      container: {}
`,
			expectedErr: `'build' tool config is not valid: container configuration is not valid: image name cannot be empty`,
		},
		{
			name: "build tool: no image name",
			input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    build:
                      container:
                        image: ???
`,
			expectedErr: `'build' tool config is not valid: container configuration is not valid: "???" is not a valid image name: invalid reference format`,
		},
		{
			name: "build tool: no *.wasm file output path",
			input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    build:
                      container:
                        image: build/image
                      output: {}
`,
			expectedErr: `'build' tool config is not valid: output configuration is not valid: *.wasm file output path cannot be empty`,
		},
		{
			name: "build tool: *.wasm file absolute output path",
			input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    build:
                      container:
                        image: build/image
                      output:
                        wasmFile: /absolute/path/to/extension.wasm
`,
			expectedErr: `'build' tool config is not valid: output configuration is not valid: *.wasm file output path must be relative to the workspace root`,
		},
		{
			name: "test tool: no image name",
			input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    test:
                      container: {}
`,
			expectedErr: `'test' tool config is not valid: container configuration is not valid: image name cannot be empty`,
		},
		{
			name: "test tool: no image name",
			input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    test:
                      container:
                        image: ???
`,
			expectedErr: `'test' tool config is not valid: container configuration is not valid: "???" is not a valid image name: invalid reference format`,
		},
		{
			name: "clean tool: no image name",
			input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    clean:
                      container: {}
`,
			expectedErr: `'clean' tool config is not valid: container configuration is not valid: image name cannot be empty`,
		},
		{
			name: "clean tool: no image name",
			input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    clean:
                      container:
                        image: ???
`,
			expectedErr: `'clean' tool config is not valid: container configuration is not valid: "???" is not a valid image name: invalid reference format`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			var toolchainConfig ToolchainConfig
			err := config.Unmarshal([]byte(test.input), &toolchainConfig)
			require.NoError(t, err)

			err = toolchainConfig.Validate()
			require.EqualError(t, err, test.expectedErr)
		})
	}
}

func TestToolchainConfigGetBuildOutputWasmFile(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "build: no config",
			input: `
                kind: BuiltinToolchain
                container:
                  image: default/image
`,
			expected: `extension.wasm`,
		},
		{
			name: "build: empty config",
			input: `
                kind: BuiltinToolchain
                container:
                  image: default/image
                build: {}
`,
			expected: `extension.wasm`,
		},
		{
			name: "build: empty output config",
			input: `
                kind: BuiltinToolchain
                container:
                  image: default/image
                build:
                  output: {}
`,
			expected: `extension.wasm`,
		},
		{
			name: "build: empty *.wasm file output path",
			input: `
                kind: BuiltinToolchain
                container:
                  image: default/image
                build:
                  output:
                    wasmFile:
`,
			expected: `extension.wasm`,
		},
		{
			name: "build: non-empty *.wasm file output path",
			input: `
                kind: BuiltinToolchain
                container:
                  image: default/image
                build:
                  output:
                    wasmFile: output/extension.wasm
`,
			expected: `output/extension.wasm`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			var toolchainConfig ToolchainConfig
			err := config.Unmarshal([]byte(test.input), &toolchainConfig)
			require.NoError(t, err)

			actual := toolchainConfig.GetBuildOutputWasmFile()
			require.Equal(t, test.expected, actual)
		})
	}
}
