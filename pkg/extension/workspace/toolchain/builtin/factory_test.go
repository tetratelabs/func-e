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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain/builtin"

	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config"
	extensionconfig "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
	builtinconfig "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/toolchain/builtin"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain/registry"
)

var _ = Describe("built-in toolchain factory", func() {

	var workspace model.Workspace

	BeforeEach(func() {
		w, err := workspaces.GetWorkspaceAt("testdata/workspace")
		Expect(err).ToNot(HaveOccurred())
		workspace = w
	})

	type testCase struct {
		config   string
		expected string
	}

	DescribeTable("should load toolchain config and apply defaults",
		func(given testCase) {
			By("verifying built-in toolchain is registered")
			factory, exists := registry.Get(builtinconfig.Kind)
			Expect(exists).To(BeTrue())

			By("loading toolchain config")
			builder, err := factory.LoadConfig(registry.LoadConfigArgs{
				Workspace: workspace,
				Toolchain: registry.ToolchainConfig{
					Name: "example",
					Config: &model.File{
						Source:  "<memory>",
						Content: []byte(given.config),
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			By("verifying that defaults get applied to the toolchain config")
			actual, err := config.Marshal(builder.GetConfig())
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))

			By("creating a toolchain")
			toolchain, err := builder.Build()
			Expect(err).ToNot(HaveOccurred())
			Expect(toolchain).ToNot(BeNil())
		},
		Entry("empty config", testCase{
			config: `kind: BuiltinToolchain`,
			expected: `
            kind: BuiltinToolchain
            container:
              image: getenvoy/extension-rust-builder:latest
            build:
              output:
                wasmFile: target/getenvoy/extension.wasm
`,
		}),
		Entry("example config", testCase{
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
		}),
		Entry("full config", testCase{
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
		}),
	)
})
