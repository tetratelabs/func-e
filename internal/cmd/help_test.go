// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestFuncEHelp(t *testing.T) {
	for _, command := range []string{"", "use", "versions", "run", "which"} {
		t.Run(command, func(t *testing.T) {
			c, stdout, _ := newApp(&globals.GlobalOpts{Version: "1.0"})
			args := []string{"func-e"}
			if command != "" {
				args = []string{"func-e", "help", command}
			}
			require.NoError(t, c.Run(args))

			expected := "func-e_help.txt"
			if command != "" {
				expected = fmt.Sprintf("func-e_%s_help.txt", command)
			}
			bytes, err := os.ReadFile(filepath.Join("testdata", expected))
			require.NoError(t, err)
			expectedStdout := string(bytes)
			expectedStdout = strings.ReplaceAll(expectedStdout, "1.99.0", version.LastKnownEnvoy.String())
			expectedStdout = strings.ReplaceAll(expectedStdout, "1.99", version.LastKnownEnvoyMinor.String())
			require.Equal(t, expectedStdout, stdout.String())
		})
	}
}
