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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/cmd/extension/globals"
	cmdtest "github.com/tetratelabs/getenvoy/pkg/test/cmd"
	cmdutil "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

func TestGetEnvoyExtensionGlobalFlags(t *testing.T) {
	type testCase struct {
		flag     string
		value    *bool
		expected bool
	}
	tests := []testCase{ // we don't test default as that depends on the runtime env
		{
			flag:     "--no-prompt",
			value:    &globals.NoPrompt,
			expected: true,
		},
		{
			flag:     "--no-prompt=true",
			value:    &globals.NoPrompt,
			expected: true,
		},
		{
			flag:     "--no-prompt=false",
			value:    &globals.NoPrompt,
			expected: false,
		},
		{
			flag:     "--no-colors",
			value:    &globals.NoColors,
			expected: true,
		},
		{
			flag:     "--no-colors=true",
			value:    &globals.NoColors,
			expected: true,
		},
		{
			flag:     "--no-colors=false",
			value:    &globals.NoColors,
			expected: false,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.flag, func(t *testing.T) {
			// Run "getenvoy extension"
			c, stdout, stderr := cmdtest.NewRootCommand()
			c.SetArgs(append([]string{"extension"}, test.flag))
			err := cmdutil.Execute(c)

			require.NoError(t, err, `expected no error running [%v]`, c)
			require.NotEmpty(t, stdout.String(), `expected stdout running [%v]`, c)
			require.Empty(t, stderr.String(), `expected no stderr running [%v]`, c)

			require.Equal(t, test.expected, *test.value)
		})
	}
}
