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

package manifest_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/manifest"
)

func TestParseReference(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected manifest.Reference
	}{
		{
			name:     "1.17.1",
			input:    `1.17.1`,
			expected: manifest.Reference{Flavor: "standard", Version: "1.17.1", Platform: manifest.CurrentPlatform()},
		},
		{
			name:     "standard:1.17.1",
			input:    `standard:1.17.1`,
			expected: manifest.Reference{Flavor: "standard", Version: "1.17.1", Platform: manifest.CurrentPlatform()},
		},
		{
			name:     "standard:1.17.1/darwin",
			input:    `standard:1.17.1/darwin`,
			expected: manifest.Reference{Flavor: "standard", Version: "1.17.1", Platform: "darwin"},
		},
		{
			name:     "standard:1.17.1/linux-glibc",
			input:    `standard:1.17.1/linux-glibc`,
			expected: manifest.Reference{Flavor: "standard", Version: "1.17.1", Platform: "linux-glibc"},
		},
		{
			name:     "coerces to lower-hyphen case format",
			input:    `stAndArd:VeRsIoN/LINUX_GLIBC`,
			expected: manifest.Reference{Flavor: "standard", Version: "version", Platform: "linux-glibc"},
		},
		{
			name:     "trailing slash",
			input:    `standard:1.17.1/`,
			expected: manifest.Reference{Flavor: "standard", Version: "1.17.1", Platform: manifest.CurrentPlatform()},
		},
		{
			name:     "non-standard",
			input:    `experiment:latest`,
			expected: manifest.Reference{Flavor: "experiment", Version: "latest", Platform: manifest.CurrentPlatform()},
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			ref, err := manifest.ParseReference(test.input)
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
			expectedErr: `"" is not a valid GetEnvoy reference. Expected format: [<flavor>:]<version>[/<platform>]`,
		},
		{
			name:        "empty components",
			input:       `:/`,
			expectedErr: `":/" is not a valid GetEnvoy reference. Expected format: [<flavor>:]<version>[/<platform>]`,
		},
		{
			name:        "no flavor",
			input:       `:1.17.1/darwin`,
			expectedErr: `":1.17.1/darwin" is not a valid GetEnvoy reference. Expected format: [<flavor>:]<version>[/<platform>]`,
		},
		{
			name:        "no version",
			input:       `standard:/darwin`,
			expectedErr: `"standard:/darwin" is not a valid GetEnvoy reference. Expected format: [<flavor>:]<version>[/<platform>]`,
		},
		{
			name:        "extra platform",
			input:       `wasm:1.17.1/darwin/darwin`,
			expectedErr: `"wasm:1.17.1/darwin/darwin" is not a valid GetEnvoy reference. Expected format: [<flavor>:]<version>[/<platform>]`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			actual, err := manifest.ParseReference(test.input)
			require.EqualError(t, err, test.expectedErr)
			require.Nil(t, actual)
		})
	}
}
