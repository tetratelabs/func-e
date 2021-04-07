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

package cmd_test

import (
	"bytes"
	"fmt"
	"syscall"
	"testing"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	commonerrors "github.com/tetratelabs/getenvoy/pkg/errors"
	. "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

func TestExecute(t *testing.T) {
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)

	c := &cobra.Command{
		Use: "getenvoy",
		Run: func(_ *cobra.Command, _ []string) {},
	}
	c.SetOut(stdout)
	c.SetErr(stderr)
	c.SetArgs([]string{})

	err := Execute(c)
	require.NoError(t, err)
	require.Empty(t, stdout)
	require.Empty(t, stderr)
}

func TestExecuteValidatesArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedErr string
	}{
		{
			name:        "unknown flag",
			args:        []string{"--xyz"},
			expectedErr: `unknown flag: --xyz`,
		},
		{
			name:        "unknown command",
			args:        []string{"other command"},
			expectedErr: `unknown command "other command" for "getenvoy"`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)

			c := &cobra.Command{
				Use: "getenvoy",
				RunE: func(_ *cobra.Command, _ []string) error {
					return errors.New("unexpected root error")
				},
			}
			c.AddCommand(&cobra.Command{
				Use: "init",
				RunE: func(_ *cobra.Command, _ []string) error {
					return errors.New("unexpected subcommand error")
				},
			})
			c.SetOut(stdout)
			c.SetErr(stderr)
			c.SetArgs(test.args)

			err := Execute(c)
			require.EqualError(t, err, test.expectedErr, `expected an error running [%v]`, c)
			require.Empty(t, stdout)
			expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy --help' for usage.\n", test.expectedErr)
			require.Equal(t, expectedStderr, stderr.String(), `unexpected stderr running [%v]`, c)
		})
	}
}

func TestExecuteApplicationSpecificError(t *testing.T) {
	tests := []struct {
		name           string
		expectedErr    error
		expectedStderr string
	}{
		{
			name:        "arbitrary error",
			expectedErr: errors.New("expected error"),
			expectedStderr: `Error: expected error

Run 'getenvoy --help' for usage.
`,
		},
		{
			name:        "shutdown error",
			expectedErr: commonerrors.NewShutdownError(syscall.SIGINT),
			expectedStderr: `NOTE: Shutting down early because a Ctrl-C ("interrupt") was received.
`,
		},
		{
			name:        "wrapped shutdown error",
			expectedErr: errors.Wrap(commonerrors.NewShutdownError(syscall.SIGINT), "wrapped"),
			expectedStderr: `NOTE: Shutting down early because a Ctrl-C ("interrupt") was received.
`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)

			c := &cobra.Command{
				Use: "getenvoy",
				RunE: func(_ *cobra.Command, _ []string) error {
					return test.expectedErr
				},
			}
			c.SetOut(stdout)
			c.SetErr(stderr)
			c.SetArgs([]string{})

			err := Execute(c)
			require.Equal(t, test.expectedErr, err)
			require.Empty(t, stdout)
			require.Equal(t, test.expectedStderr, stderr.String())
		})
	}
}
