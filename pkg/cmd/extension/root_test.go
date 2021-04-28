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

	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/globals"
	rootcmd "github.com/tetratelabs/getenvoy/pkg/cmd"
	cmdtest "github.com/tetratelabs/getenvoy/pkg/test/cmd"
)

func TestGetEnvoyExtensionGlobalFlags(t *testing.T) {
	type testCase struct {
		flag     string
		value    func(o *globals.GlobalOpts) bool
		expected bool
	}
	tests := []testCase{ // we don't test default as that depends on the runtime env
		{
			flag:     "--no-prompt",
			value:    func(o *globals.GlobalOpts) bool { return o.NoWizard },
			expected: true,
		},
		{
			flag:     "--no-prompt=true",
			value:    func(o *globals.GlobalOpts) bool { return o.NoWizard },
			expected: true,
		},
		{
			flag:     "--no-prompt=false",
			value:    func(o *globals.GlobalOpts) bool { return o.NoWizard },
			expected: false,
		},
		{
			flag:     "--no-colors",
			value:    func(o *globals.GlobalOpts) bool { return o.NoColors },
			expected: true,
		},
		{
			flag:     "--no-colors=true",
			value:    func(o *globals.GlobalOpts) bool { return o.NoColors },
			expected: true,
		},
		{
			flag:     "--no-colors=false",
			value:    func(o *globals.GlobalOpts) bool { return o.NoColors },
			expected: false,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.flag, func(t *testing.T) {
			// Run "getenvoy extension"
			o := &globals.GlobalOpts{}
			c, stdout, stderr := cmdtest.NewRootCommand(o)
			c.SetArgs([]string{"extension", test.flag})
			err := rootcmd.Execute(c)

			require.NoError(t, err, `expected no error running [%v]`, c)
			require.NotEmpty(t, stdout.String(), `expected stdout running [%v]`, c)
			require.Empty(t, stderr.String(), `expected no stderr running [%v]`, c)

			require.Equal(t, test.expected, test.value(o))
		})
	}
}
