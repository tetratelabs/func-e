// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	rootcmd "github.com/tetratelabs/func-e/internal/cmd"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestFuncEHelp(t *testing.T) {
	for _, command := range []string{"", "use", "versions", "run", "which"} {
		t.Run(command, func(t *testing.T) {
			stdout := new(bytes.Buffer)
			stderr := new(bytes.Buffer)
			var args []string
			if command != "" {
				args = []string{"--help", command}
			} else {
				args = []string{"--help"}
			}

			err := rootcmd.DoMain(t.Context(), stdout, stderr, args, nil, "1.0")
			code, isExit := rootcmd.IsExit(err)
			require.True(t, isExit, "expected ExitError from --help, got %v", err)
			require.Equal(t, 0, code)

			expected := "func-e_help.txt"
			if command != "" {
				expected = fmt.Sprintf("func-e_%s_help.txt", command)
			}

			b, err := os.ReadFile(filepath.Join("testdata", expected))
			require.NoError(t, err)
			expectedStdout := string(b)
			expectedStdout = strings.ReplaceAll(expectedStdout, "1.99.0", version.LastKnownEnvoy.String())
			expectedStdout = strings.ReplaceAll(expectedStdout, "1.99", version.LastKnownEnvoyMinor.String())
			require.Equal(t, expectedStdout, stdout.String())
		})
	}
}
