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

package extension_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	reference "github.com/tetratelabs/getenvoy/pkg"
	. "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
)

func TestDescriptorValidate(t *testing.T) {
	input := fmt.Sprintf(`#
# Envoy Wasm extension created with getenvoy toolkit.
#
kind: Extension

name: mycompany.filters.http.custom_metrics

category: envoy.filters.http
language: rust

# Runtime the extension is being developed against.
runtime:
  envoy:
    version: %[1]s
`, reference.Latest)
	var descriptor Descriptor
	err := yaml.Unmarshal([]byte(input), &descriptor)
	require.NoError(t, err)

	err = descriptor.Validate()
	require.NoError(t, err)
}

func TestDescriptorValidateError(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name:        "empty",
			input:       ``,
			expectedErr: `4 errors occurred: extension name cannot be empty; extension category cannot be empty; programming language cannot be empty; runtime description is not valid: envoy version cannot be empty`,
		},
		{
			name: "invalid envoy version",
			input: `#
# Envoy Wasm extension created with getenvoy toolkit.
#
kind: Extension

name: mycompany.filters.http.custom_metrics

language: rust
category: envoy.filters.http

# Runtime the extension is being developed against.
runtime:
  envoy:
    version: invalid value
`,
			expectedErr: `runtime description is not valid: envoy version is not valid: "invalid value" is not a valid GetEnvoy reference. Expected format: <flavor>:<version>[/<platform>]`,
		},
		{
			name: "missing extension name",
			input: fmt.Sprintf(`#
# Envoy Wasm extension created with getenvoy toolkit.
#
kind: Extension

category: envoy.filters.http
language: rust

# Runtime the extension is being developed against.
runtime:
  envoy:
    version: %[1]s
`, reference.Latest),
			expectedErr: `extension name cannot be empty`,
		}, {
			name: "invalid extension name",
			input: fmt.Sprintf(`#
# Envoy Wasm extension created with getenvoy toolkit.
#
kind: Extension

name: ?!@#$%%

category: envoy.filters.http
language: rust

# Runtime the extension is being developed against.
runtime:
  envoy:
    version: %[1]s
`, reference.Latest),
			expectedErr: `"?!@#$%" is not a valid extension name. Extension name must match the format "^[a-z0-9_]+(\\.[a-z0-9_]+)*$". E.g., 'mycompany.filters.http.custom_metrics'`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			var descriptor Descriptor
			err := yaml.Unmarshal([]byte(test.input), &descriptor)
			require.NoError(t, err)

			err = descriptor.Validate()
			require.EqualError(t, err, test.expectedErr)
		})
	}
}

func TestValidateExtensionName(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Envoy-like name", "mycompany.filters.http.custom_metrics"},
		{"no segments", "myextension"},
		{"numbers", "911.i18n.v2"},
		{"'_'", "_._"},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			err := ValidateExtensionName(test.input)
			require.NoError(t, err)
		})
	}
}

func TestValidateExtensionNameError(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"trailing '.'", "myextension."},
		{"upper-case", "MYEXTENSION"},
		{"'-'", "-.-"},
		{"non alpha-num characters", `!@#$%^&*()-+<>?~:;"'\[]{}`},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			err := ValidateExtensionName(test.input)
			expectedErr := fmt.Sprintf(`%q is not a valid extension name. Extension name must match the format "^[a-z0-9_]+(\\.[a-z0-9_]+)*$". E.g., 'mycompany.filters.http.custom_metrics'`, test.input) //nolint:lll
			require.EqualError(t, err, expectedErr, test.input)
		})
	}
}

func TestSanitizeExtensionName(t *testing.T) {
	actual := SanitizeExtensionName(`My-C0mpany.com`, ``, `e!x@t#`)
	require.Equal(t, `my_c0mpany_com.e_x_t_`, actual)
}

func TestSanitizeExtensionNameSegment(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`My-C0mpany.com`, `my_c0mpany_com`},
		{`!@#$%^&*()-+<>?~:;"'\[]{}`, `_________________________`},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.input, func(t *testing.T) {
			actual := SanitizeExtensionNameSegment(test.input)
			require.Equal(t, test.expected, actual)
		})
	}
}
