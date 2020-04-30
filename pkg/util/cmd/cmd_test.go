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
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	. "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

func Test_unknown_command(t *testing.T) {
	rootCmd := &cobra.Command{
		Use: "getenvoy",
	}
	rootCmd.AddCommand(&cobra.Command{
		Use: "init",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("unexpected error")
		},
	})
	out := new(bytes.Buffer)
	rootCmd.SetOut(out)
	rootCmd.SetArgs([]string{"other", "command"})

	err := Execute(rootCmd)

	if assert.Error(t, err) {
		assert.Equal(t, `Error: unknown command "other" for "getenvoy"

Run 'getenvoy --help' for usage.
`, out.String())
	}
}

func Test_unknown_flag(t *testing.T) {
	rootCmd := &cobra.Command{
		Use: "getenvoy",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("unexpected error")
		},
	}
	out := new(bytes.Buffer)
	rootCmd.SetOut(out)
	rootCmd.SetArgs([]string{"--xyz"})

	err := Execute(rootCmd)

	if assert.Error(t, err) {
		assert.Equal(t, `Error: unknown flag: --xyz

Usage:
  getenvoy [flags]

Flags:
  -h, --help   help for getenvoy

`, out.String())
	}
}

func Test_command_returns_error(t *testing.T) {
	rootCmd := &cobra.Command{
		Use: "getenvoy",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("expected error")
		},
	}
	out := new(bytes.Buffer)
	rootCmd.SetOut(out)
	rootCmd.SetArgs([]string{})

	err := Execute(rootCmd)

	if assert.Error(t, err) {
		assert.Equal(t, `Error: expected error

Usage:
  getenvoy [flags]

Flags:
  -h, --help   help for getenvoy

`, out.String())
	}
}

func Test_successful_command(t *testing.T) {
	rootCmd := &cobra.Command{
		Use: "getenvoy",
		Run: func(_ *cobra.Command, _ []string) {},
	}
	out := new(bytes.Buffer)
	rootCmd.SetOut(out)
	rootCmd.SetArgs([]string{})

	err := Execute(rootCmd)

	if assert.NoError(t, err) {
		assert.Equal(t, ``, out.String())
	}
}
