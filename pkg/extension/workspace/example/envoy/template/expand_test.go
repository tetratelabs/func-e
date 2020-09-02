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

package template_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/tetratelabs/getenvoy/pkg/extension/workspace/example/envoy/template"

	"github.com/tetratelabs/getenvoy/pkg/extension/manager"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
)

var _ = Describe("Expand()", func() {
	Describe("in case of invalid input", func() {
		type testCase struct {
			input       string
			expectedErr string
		}
		//nolint:lll
		DescribeTable("should fail with a proper error",
			func(given testCase) {
				ctx := &ExpandContext{
					DefaultExtension:       manager.NewLocalExtension(extension.NewExtensionDescriptor(), "/path/to/extension.wasm"),
					DefaultExtensionConfig: ``,
				}

				_, err := Expand([]byte(given.input), ctx)
				Expect(err).To(MatchError(given.expectedErr))
			},
			Entry("invalid Golang template", testCase{
				input:       `{{{ "hello" }}`,
				expectedErr: `failed to parse Envoy config template: template: :1: unexpected "{" in command`,
			}),
			Entry("missing leading '.'", testCase{
				input:       `{{ GetEnvoy.DefaultValue "admin" }}`,
				expectedErr: `failed to parse Envoy config template: template: :1: function "GetEnvoy" not defined`,
			}),
			Entry("unknown property name", testCase{
				input:       `{{ .GetEnvoy.DefaultValue "???" }}`,
				expectedErr: `failed to render Envoy config template: template: :1:12: executing "" at <.GetEnvoy.DefaultValue>: error calling DefaultValue: unknown property "???"`,
			}),
			Entry("external Wasm module: name", testCase{
				input:       `{{ .GetEnvoy.Extension.Name "org/name:version" }}`,
				expectedErr: `failed to render Envoy config template: template: :1:12: executing "" at <.GetEnvoy.Extension.Name>: error calling Name: unable to resolve Wasm module [org/name:version]: not supported yet`,
			}),
			Entry("external Wasm module: code", testCase{
				input:       `{{ .GetEnvoy.Extension.Code "org/name:version" }}`,
				expectedErr: `failed to render Envoy config template: template: :1:12: executing "" at <.GetEnvoy.Extension.Code>: error calling Code: unable to resolve Wasm module [org/name:version]: not supported yet`,
			}),
			Entry("external Wasm module: config", testCase{
				input:       `{{ .GetEnvoy.Extension.Config "another-wasm-module" }}`,
				expectedErr: `failed to render Envoy config template: template: :1:12: executing "" at <.GetEnvoy.Extension.Config>: error calling Config: unable to resolve a named config [another-wasm-module]: not supported yet`,
			}),
		)
	})

	Describe("in case of valid input", func() {
		type testCase struct {
			input    string
			expected string
		}
		DescribeTable("should fail with a proper error",
			func(given testCase) {
				ctx := &ExpandContext{
					DefaultExtension: manager.NewLocalExtension(
						&extension.Descriptor{
							Name:     "mycompany.filters.http.custom_metrics",
							Category: extension.EnvoyHTTPFilter,
							Language: extension.LanguageRust,
						},
						"/path/to/extension.wasm",
					),
					DefaultExtensionConfig: `{"key":"value"}`,
				}

				actual, err := Expand([]byte(given.input), ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(actual)).To(Equal(given.expected))
			},
			Entry("default value: admin", testCase{
				input:    `{{ .GetEnvoy.DefaultValue "admin" }}`,
				expected: `{"accessLogPath":"/dev/null","address":{"socketAddress":{"address":"127.0.0.1","portValue":9901}}}`,
			}),
			Entry("default value: admin.access_log_path", testCase{
				input:    `{{ .GetEnvoy.DefaultValue "admin.access_log_path" }}`,
				expected: `"/dev/null"`,
			}),
			Entry("default value: admin.address", testCase{
				input:    `{{ .GetEnvoy.DefaultValue "admin.address" }}`,
				expected: `{"socketAddress":{"address":"127.0.0.1","portValue":9901}}`,
			}),
			Entry("default value: admin.address.socket.address", testCase{
				input:    `{{ .GetEnvoy.DefaultValue "admin.address.socket.address" }}`,
				expected: `"127.0.0.1"`,
			}),
			Entry("default value: admin.address.socket.port", testCase{
				input:    `{{ .GetEnvoy.DefaultValue "admin.address.socket.port" }}`,
				expected: `9901`,
			}),
			Entry("extension: name", testCase{
				input:    `{{ .GetEnvoy.Extension.Name }}`,
				expected: `"mycompany.filters.http.custom_metrics"`,
			}),
			Entry("extension: code", testCase{
				input:    `{{ .GetEnvoy.Extension.Code }}`,
				expected: `{"local":{"filename":"/path/to/extension.wasm"}}`,
			}),
			Entry("extension: config", testCase{
				input:    `{{ .GetEnvoy.Extension.Config }}`,
				expected: `{"@type":"type.googleapis.com/google.protobuf.StringValue","value":"{\"key\":\"value\"}"}`,
			}),
			Entry("access to proto message", testCase{
				input:    `{{ (.GetEnvoy.Extension.Code).Message.GetLocal.GetFilename }}`,
				expected: `/path/to/extension.wasm`,
			}),
		)
	})
})
