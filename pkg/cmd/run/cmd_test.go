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

package run_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/globals"
	rootcmd "github.com/tetratelabs/getenvoy/pkg/cmd"
	"github.com/tetratelabs/getenvoy/pkg/test/cmd"
)

func TestGetEnvoyRunValidateFlag(t *testing.T) {
	type testCase struct {
		name        string
		args        []string
		expectedErr string
	}
	tests := []testCase{
		{
			name:        "arg[0] with invalid reference",
			args:        []string{"run", "???"},
			expectedErr: `"???" is not a valid GetEnvoy reference. Expected format: <flavor>:<version>[/<platform>]`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			// Run "getenvoy extension run"
			c, stdout, stderr := cmd.NewRootCommand(&globals.GlobalOpts{})
			c.SetArgs(test.args)
			err := rootcmd.Execute(c)

			// Verify the command failed with the expected error
			require.EqualError(t, err, test.expectedErr, `expected an error running [%v]`, c)
			require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
			expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy run --help' for usage.\n", test.expectedErr)
			require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, c)
		})
	}
}
