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

package args

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplitCommandLine(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "nil",
			input:    nil,
			expected: []string{},
		},
		{
			name:     "empty",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "already split",
			input:    []string{"-e", "VAR=VALUE"},
			expected: []string{"-e", "VAR=VALUE"},
		},
		{
			name:  "command line",
			input: []string{"-e VAR=VALUE -v /host/path:/container/path"},
			expected: []string{
				"-e",
				"VAR=VALUE",
				"-v",
				"/host/path:/container/path",
			},
		},
		{
			name:  "quoted command line",
			input: []string{`'-e VAR=VALUE' "-v /host/path:/container/path"`},
			expected: []string{
				"-e VAR=VALUE",
				"-v /host/path:/container/path",
			},
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			actual, err := SplitCommandLine(test.input...)
			require.NoError(t, err)
			require.Equal(t, test.expected, actual)
		})
	}
}
