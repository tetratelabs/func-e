// Copyright 2021 Tetrate
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

package version

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewVersion(t *testing.T) {
	tests := []struct {
		input       string
		expected    Version
		expectedErr string
	}{
		{
			input:    "1.19",
			expected: MinorVersion("1.19"),
		},
		{
			input:    "1.19_debug",
			expected: MinorVersion("1.19_debug"),
		},
		{
			input:    "2.0",
			expected: MinorVersion("2.0"),
		},
		{
			input:    "2.0_debug",
			expected: MinorVersion("2.0_debug"),
		},
		{
			input:    "1.19.1",
			expected: PatchVersion("1.19.1"),
		},
		{
			input:    "1.19.1_debug",
			expected: PatchVersion("1.19.1_debug"),
		},
		{
			input:    "2.0.0",
			expected: PatchVersion("2.0.0"),
		},
		{
			input:    "2.0.0_debug",
			expected: PatchVersion("2.0.0_debug"),
		},
		{
			input:       "",
			expectedErr: "missing [version] argument",
		},
		{
			input:       "a.b.c",
			expectedErr: fmt.Sprintf(`invalid [version] argument: "a.b.c" should look like %q or %q`, LastKnownEnvoy, LastKnownEnvoyMinor),
		},
	}

	for _, tt := range tests {
		tc := tt // pin! see https://github.com/kyoh86/scopelint for why
		t.Run(tc.input, func(t *testing.T) {
			actual, err := NewVersion("[version] argument", tc.input)
			require.Equal(t, tc.expected, actual)
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}

func TestNewMinorVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected MinorVersion
	}{
		{
			input:    "1.19",
			expected: MinorVersion("1.19"),
		},
		{
			input:    "1.19_debug",
			expected: MinorVersion("1.19_debug"),
		},
		{
			input:    "2.0",
			expected: MinorVersion("2.0"),
		},
		{
			input:    "2.0_debug",
			expected: MinorVersion("2.0_debug"),
		},
		{input: "1.19.1"},
		{input: "1.19.1_debug"},
		{input: "1.19.1.2"},
		{input: "1.19."},
		{input: "1.19-debug"},
		{input: "1."},
		{input: "1"},
		{},
	}

	for _, tt := range tests {
		tc := tt // pin! see https://github.com/kyoh86/scopelint for why
		t.Run(tc.input, func(t *testing.T) {
			actual := NewMinorVersion(tc.input)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestNewPatchVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected PatchVersion
	}{
		{
			input:    "1.19.1",
			expected: PatchVersion("1.19.1"),
		},
		{
			input:    "1.19.1_debug",
			expected: PatchVersion("1.19.1_debug"),
		},
		{
			input:    "2.0.0",
			expected: PatchVersion("2.0.0"),
		},
		{
			input:    "2.0.0_debug",
			expected: PatchVersion("2.0.0_debug"),
		},
		{input: "1.19"},
		{input: "1.19_debug"},
		{input: "1.19.1.2"},
		{input: "1.19.1."},
		{input: "1.19.1-debug"},
		{input: "1."},
		{input: "1"},
		{},
	}

	for _, tt := range tests {
		tc := tt // pin! see https://github.com/kyoh86/scopelint for why
		t.Run(tc.input, func(t *testing.T) {
			actual := NewPatchVersion(tc.input)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestPatchVersion_Patch(t *testing.T) {
	tests := []struct {
		input    PatchVersion
		expected int
	}{
		{
			input:    PatchVersion("1.19.0"),
			expected: 0,
		},
		{
			input:    PatchVersion("1.19.0_debug"),
			expected: 0,
		},
		{
			input:    PatchVersion("1.19.1"),
			expected: 1,
		},
		{
			input:    PatchVersion("1.19.1_debug"),
			expected: 1,
		},
		{
			input:    PatchVersion("1.19.10"),
			expected: 10,
		},
		{
			input:    PatchVersion("1.19.10_debug"),
			expected: 10,
		},
		{ // bad data which is impossible if instantiated properly
			input:    PatchVersion("ice cream"),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input.String(), func(t *testing.T) {
			actual := tt.input.Patch()
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestVersion_String(t *testing.T) {
	tests := []struct {
		input    Version
		expected string
	}{
		{
			input:    MinorVersion("1.19"),
			expected: "1.19",
		},
		{
			input:    MinorVersion("1.19_debug"),
			expected: "1.19_debug",
		},
		{
			input:    PatchVersion("1.19.1"),
			expected: "1.19.1",
		},
		{
			input:    PatchVersion("1.19.1_debug"),
			expected: "1.19.1_debug",
		},
	}

	for _, tc := range tests {
		t.Run(tc.input.String(), func(t *testing.T) {
			actual := tc.input.String()
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestVersion_ToMinor(t *testing.T) {
	tests := []struct {
		input    Version
		expected MinorVersion
	}{
		{
			input:    MinorVersion("1.19"),
			expected: MinorVersion("1.19"),
		},
		{
			input:    MinorVersion("1.19_debug"),
			expected: MinorVersion("1.19_debug"),
		},
		{
			input:    PatchVersion("1.19.1"),
			expected: MinorVersion("1.19"),
		},
		{
			input:    PatchVersion("1.19.1_debug"),
			expected: MinorVersion("1.19_debug"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.input.String(), func(t *testing.T) {
			actual := tt.input.ToMinor()
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestFindLatestPatch(t *testing.T) {
	type testCase struct {
		name          string
		patchVersions []PatchVersion
		minorVersion  MinorVersion
		expected      PatchVersion
	}

	tests := []testCase{
		{
			name: "zero",
			patchVersions: []PatchVersion{
				PatchVersion("1.20.0_debug"), // mixed debug and not is unlikely, but possible
				PatchVersion("1.20.0"),
			},
			minorVersion: MinorVersion("1.20"),
			expected:     PatchVersion("1.20.0"),
		},
		{
			name: "upgradable",
			patchVersions: []PatchVersion{
				PatchVersion("1.18.3"),
				PatchVersion("1.18.14"),
				PatchVersion("1.18.4"),
				PatchVersion("1.18.4_debug"),
			},
			minorVersion: MinorVersion("1.18"),
			expected:     PatchVersion("1.18.14"),
		},
		{
			name: "notfound",
			patchVersions: []PatchVersion{
				PatchVersion("1.20.0"),
				PatchVersion("1.1_debug"),
			},
			minorVersion: MinorVersion("1.1"),
		},
		{
			name: "debug",
			patchVersions: []PatchVersion{
				PatchVersion("1.19.10_debug"),
				PatchVersion("1.19.2_debug"),
				PatchVersion("1.19.1"),
			},
			minorVersion: MinorVersion("1.19_debug"),
			expected:     PatchVersion("1.19.10_debug"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := FindLatestPatchVersion(tt.patchVersions, tt.minorVersion)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestFindLatestVersion(t *testing.T) {
	type testCase struct {
		name          string
		patchVersions []PatchVersion
		expected      PatchVersion
	}

	tests := []testCase{
		{
			name:          "empty",
			patchVersions: []PatchVersion{},
		},
		{
			name: "all debug is empty",
			patchVersions: []PatchVersion{
				PatchVersion("1.19.0_debug"),
				PatchVersion("1.20.0_debug"),
			},
		},
		{
			name: "ignores debug",
			patchVersions: []PatchVersion{
				PatchVersion("1.20.0_debug"), // mixed debug and not is unlikely, but possible
				PatchVersion("1.20.0"),
			},
			expected: PatchVersion("1.20.0"),
		},
		{
			name: "latest patch",
			patchVersions: []PatchVersion{
				PatchVersion("1.18.1"),
				PatchVersion("1.18.14"),
				PatchVersion("1.18.4"),
			},
			expected: PatchVersion("1.18.14"),
		},
		{
			name: "latest version",
			patchVersions: []PatchVersion{
				PatchVersion("1.20.1"),
				PatchVersion("1.18.2"),
			},
			expected: PatchVersion("1.20.1"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := FindLatestVersion(tt.patchVersions)
			require.Equal(t, tt.expected, actual)
		})
	}
}
