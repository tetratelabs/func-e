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

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config"

	. "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/toolchain/builtin"
)

var _ = Describe("ToolchainConfig", func() {
	Describe("Validate()", func() {
		Describe("in case of valid input", func() {
			type testCase struct {
				input string
			}
			DescribeTable("should not return any error",
				func(given testCase) {
					var toolchain ToolchainConfig
					err := config.Unmarshal([]byte(given.input), &toolchain)
					Expect(err).NotTo(HaveOccurred())

					err = toolchain.Validate()
					Expect(err).ToNot(HaveOccurred())

					actual, err := config.Marshal(toolchain)
					Expect(err).NotTo(HaveOccurred())

					Expect(actual).To(MatchYAML(given.input))
				},
				Entry("default build container", testCase{
					input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
`,
				}),
				Entry("empty build config", testCase{
					input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    build: {}
`,
				}),
				Entry("'build' config with container", testCase{
					input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    build:
                      container:
                        image: build/image
`,
				}),
				Entry("'build' config with *.wasm file output path", testCase{
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
				}),
				Entry("'test' config with container", testCase{
					input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    test:
                      container:
                        image: test/image
`,
				}),
				Entry("'clean' config with container", testCase{
					input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    clean:
                      container:
                        image: clean/image
`,
				}),
			)
		})

		Describe("in case of invalid input", func() {
			type testCase struct {
				input       string
				expectedErr string
			}
			DescribeTable("should fail with a proper error",
				func(given testCase) {
					var toolchain ToolchainConfig
					err := config.Unmarshal([]byte(given.input), &toolchain)
					Expect(err).NotTo(HaveOccurred())

					err = toolchain.Validate()
					Expect(err).To(MatchError(given.expectedErr))
				},
				Entry("blank", testCase{
					input:       ``,
					expectedErr: `configuration of the default build container cannot be empty`,
				}),
				Entry("no default build container", testCase{
					input: `
                    kind: BuiltinToolchain
`,
					expectedErr: `configuration of the default build container cannot be empty`,
				}),
				Entry("default build container: no image name", testCase{
					input: `
                    kind: BuiltinToolchain
                    container: {}
`,
					expectedErr: `configuration of the default build container is not valid: image name cannot be empty`,
				}),
				Entry("default build container: invalid image name", testCase{
					input: `
                    kind: BuiltinToolchain
                    container:
                      image: ???
`,
					expectedErr: `configuration of the default build container is not valid: "???" is not a valid image name: invalid reference format`,
				}),
				Entry("build tool: no image name", testCase{
					input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    build:
                      container: {}
`,
					expectedErr: `'build' tool config is not valid: container configuration is not valid: image name cannot be empty`,
				}),
				Entry("build tool: no image name", testCase{
					input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    build:
                      container:
                        image: ???
`,
					expectedErr: `'build' tool config is not valid: container configuration is not valid: "???" is not a valid image name: invalid reference format`,
				}),
				Entry("build tool: no *.wasm file output path", testCase{
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
				}),
				Entry("build tool: *.wasm file absolute output path", testCase{
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
				}),
				Entry("test tool: no image name", testCase{
					input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    test:
                      container: {}
`,
					expectedErr: `'test' tool config is not valid: container configuration is not valid: image name cannot be empty`,
				}),
				Entry("test tool: no image name", testCase{
					input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    test:
                      container:
                        image: ???
`,
					expectedErr: `'test' tool config is not valid: container configuration is not valid: "???" is not a valid image name: invalid reference format`,
				}),
				Entry("clean tool: no image name", testCase{
					input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    clean:
                      container: {}
`,
					expectedErr: `'clean' tool config is not valid: container configuration is not valid: image name cannot be empty`,
				}),
				Entry("clean tool: no image name", testCase{
					input: `
                    kind: BuiltinToolchain
                    container:
                      image: default/image
                    clean:
                      container:
                        image: ???
`,
					expectedErr: `'clean' tool config is not valid: container configuration is not valid: "???" is not a valid image name: invalid reference format`,
				}),
			)
		})
	})

	Describe("GetBuildOutputWasmFile()", func() {
		type testCase struct {
			input    string
			expected string
		}
		DescribeTable("should return proper output path",
			func(given testCase) {
				var toolchain ToolchainConfig
				err := config.Unmarshal([]byte(given.input), &toolchain)
				Expect(err).NotTo(HaveOccurred())

				Expect(toolchain.GetBuildOutputWasmFile()).To(Equal(given.expected))
			},
			Entry("build: no config", testCase{
				input: `
                kind: BuiltinToolchain
                container:
                  image: default/image
`,
				expected: `extension.wasm`,
			}),
			Entry("build: empty config", testCase{
				input: `
                kind: BuiltinToolchain
                container:
                  image: default/image
                build: {}
`,
				expected: `extension.wasm`,
			}),
			Entry("build: empty output config", testCase{
				input: `
                kind: BuiltinToolchain
                container:
                  image: default/image
                build:
                  output: {}
`,
				expected: `extension.wasm`,
			}),
			Entry("build: empty *.wasm file output path", testCase{
				input: `
                kind: BuiltinToolchain
                container:
                  image: default/image
                build:
                  output:
                    wasmFile:
`,
				expected: `extension.wasm`,
			}),
			Entry("build: non-empty *.wasm file output path", testCase{
				input: `
                kind: BuiltinToolchain
                container:
                  image: default/image
                build:
                  output:
                    wasmFile: output/extension.wasm
`,
				expected: `output/extension.wasm`,
			}),
		)
	})
})
