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

package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/tetratelabs/getenvoy/pkg/types"
)

func TestParseReference(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Reference
	}{
		{
			name:     "standard:1.15.3",
			input:    `standard:1.15.3`,
			expected: Reference{Flavor: "standard", Version: "1.15.3", Platform: ""},
		},
		{
			name:     "standard:1.15.3/darwin",
			input:    `standard:1.15.3/darwin`,
			expected: Reference{Flavor: "standard", Version: "1.15.3", Platform: "darwin"},
		},
		{
			name:     "standard:1.15.3/linux-glibc",
			input:    `standard:1.15.3/linux-glibc`,
			expected: Reference{Flavor: "standard", Version: "1.15.3", Platform: "linux-glibc"},
		},
		{
			name:     "wasm:1.15",
			input:    `wasm:1.15`,
			expected: Reference{Flavor: "wasm", Version: "1.15", Platform: ""},
		},
		{
			name:     "wasm:1.15/darwin",
			input:    `wasm:1.15/darwin`,
			expected: Reference{Flavor: "wasm", Version: "1.15", Platform: "darwin"},
		},
		{
			name:     "wasm:1.15/linux-glibc",
			input:    `wasm:1.15/linux-glib`,
			expected: Reference{Flavor: "wasm", Version: "1.15", Platform: "linux-glib"},
		},
		{
			name:     "mixed case",
			input:    `Wasm:NightlY/LINUX-GLIBC`,
			expected: Reference{Flavor: "wasm", Version: "nightly", Platform: "linux-glibc"},
		},
		{
			name:     "trailing slash",
			input:    `standard:1.15.3/`,
			expected: Reference{Flavor: "standard", Version: "1.15.3", Platform: ""},
		},
		{
			name:     "special characters",
			input:    `abcd-EFGH.01234_:-56789.XYZ_/`,
			expected: Reference{Flavor: "abcd-efgh.01234_", Version: "-56789.xyz_", Platform: ""},
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			ref, err := ParseReference(test.input)
			require.NoError(t, err)
			require.Equal(t, &test.expected, ref)
		})
	}
}

func TestParseReferenceValidatesInput(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name:        "empty string",
			input:       ``,
			expectedErr: `"" is not a valid GetEnvoy reference. Expected format: <flavor>:<version>[/<platform>]`,
		},
		{
			name:        "empty components",
			input:       `:/`,
			expectedErr: `":/" is not a valid GetEnvoy reference. Expected format: <flavor>:<version>[/<platform>]`,
		},
		{
			name:        "no flavor",
			input:       `:1.15.3/darwin`,
			expectedErr: `":1.15.3/darwin" is not a valid GetEnvoy reference. Expected format: <flavor>:<version>[/<platform>]`,
		},
		{
			name:        "no version",
			input:       `standard:/darwin`,
			expectedErr: `"standard:/darwin" is not a valid GetEnvoy reference. Expected format: <flavor>:<version>[/<platform>]`,
		},
		{
			name:        "invalid character in flavor",
			input:       `stan dard:1.15.3/darwin`,
			expectedErr: `"stan dard:1.15.3/darwin" is not a valid GetEnvoy reference. Expected format: <flavor>:<version>[/<platform>]`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			actual, err := ParseReference(test.input)
			require.EqualError(t, err, test.expectedErr)
			require.Nil(t, actual)
		})
	}
}

func TestParseReferenceNormalizesStringForm(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    `standard:1.15.3`,
			expected: `standard:1.15.3`,
		},
		{
			input:    `standard:1.15.3/darwin`,
			expected: `standard:1.15.3/darwin`,
		},
		{
			input:    `standard:1.15.3/`,
			expected: `standard:1.15.3`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.input, func(t *testing.T) {
			ref, err := ParseReference(test.input)
			require.NoError(t, err)

			// Ensure ref.String() normalizes the input
			require.Equal(t, test.expected, ref.String())

			// Round-trip that the ultimate string form makes the same reference
			ref2, err := ParseReference(ref.String())
			require.NoError(t, err)
			require.Equal(t, ref, ref2)
		})
	}
}
