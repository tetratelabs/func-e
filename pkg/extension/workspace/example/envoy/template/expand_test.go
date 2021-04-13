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

package template

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/extension/manager"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
)

func TestExpand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "default value: admin",
			input:    `{{ .GetEnvoy.DefaultValue "admin" }}`,
			expected: `{"accessLogPath":"/dev/null","address":{"socketAddress":{"address":"127.0.0.1","portValue":9901}}}`,
		},
		{
			name:     "default value: admin.access_log_path",
			input:    `{{ .GetEnvoy.DefaultValue "admin.access_log_path" }}`,
			expected: `"/dev/null"`,
		},
		{
			name:     "default value: admin.address",
			input:    `{{ .GetEnvoy.DefaultValue "admin.address" }}`,
			expected: `{"socketAddress":{"address":"127.0.0.1","portValue":9901}}`,
		},
		{
			name:     "default value: admin.address.socket.address",
			input:    `{{ .GetEnvoy.DefaultValue "admin.address.socket.address" }}`,
			expected: `"127.0.0.1"`,
		},
		{
			name:     "default value: admin.address.socket.port",
			input:    `{{ .GetEnvoy.DefaultValue "admin.address.socket.port" }}`,
			expected: `9901`,
		},
		{
			name:     "extension: name",
			input:    `{{ .GetEnvoy.Extension.Name }}`,
			expected: `"mycompany.filters.http.custom_metrics"`,
		},
		{
			name:     "extension: code",
			input:    `{{ .GetEnvoy.Extension.Code }}`,
			expected: `{"local":{"filename":"/path/to/extension.wasm"}}`,
		},
		{
			name:     "extension: config",
			input:    `{{ .GetEnvoy.Extension.Config }}`,
			expected: defaultExtensionConfigJSON(t, `{"key":"value"}`),
		},
		{
			name:     "access to proto message",
			input:    `"{{ (.GetEnvoy.Extension.Code).Message.GetLocal.GetFilename }}"`,
			expected: `"/path/to/extension.wasm"`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
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

			actual, err := Expand([]byte(test.input), ctx)
			require.NoError(t, err)
			require.JSONEq(t, test.expected, string(actual))
		})
	}
}

func TestExpandValidates(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name:        "invalid Golang template",
			input:       `{{{ "hello" }}`,
			expectedErr: `failed to parse Envoy config template: template: :1: unexpected "{" in command`,
		},
		{
			name:        "missing leading '.'",
			input:       `{{ GetEnvoy.DefaultValue "admin" }}`,
			expectedErr: `failed to parse Envoy config template: template: :1: function "GetEnvoy" not defined`,
		},
		{
			name:        "unknown property name",
			input:       `{{ .GetEnvoy.DefaultValue "???" }}`,
			expectedErr: `failed to render Envoy config template: template: :1:12: executing "" at <.GetEnvoy.DefaultValue>: error calling DefaultValue: unknown property "???"`,
		},
		{
			name:        "external Wasm module: name",
			input:       `{{ .GetEnvoy.Extension.Name "org/name:version" }}`,
			expectedErr: `failed to render Envoy config template: template: :1:12: executing "" at <.GetEnvoy.Extension.Name>: error calling Name: unable to resolve Wasm module [org/name:version]: not supported yet`,
		},
		{
			name:        "external Wasm module: code",
			input:       `{{ .GetEnvoy.Extension.Code "org/name:version" }}`,
			expectedErr: `failed to render Envoy config template: template: :1:12: executing "" at <.GetEnvoy.Extension.Code>: error calling Code: unable to resolve Wasm module [org/name:version]: not supported yet`,
		},
		{
			name:        "external Wasm module: config",
			input:       `{{ .GetEnvoy.Extension.Config "another-wasm-module" }}`,
			expectedErr: `failed to render Envoy config template: template: :1:12: executing "" at <.GetEnvoy.Extension.Config>: error calling Config: unable to resolve a named config [another-wasm-module]: not supported yet`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			ctx := &ExpandContext{
				DefaultExtension:       manager.NewLocalExtension(extension.NewExtensionDescriptor(), "/path/to/extension.wasm"),
				DefaultExtensionConfig: ``,
			}

			actual, err := Expand([]byte(test.input), ctx)
			require.Nil(t, actual)
			require.EqualError(t, err, test.expectedErr)
		})
	}
}
