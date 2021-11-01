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

package cmd_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestFuncEHelp(t *testing.T) {
	for _, command := range []string{"", "use", "versions", "run", "which"} {
		command := command
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
			expectedStdout := moreos.Sprintf(string(bytes))
			expectedStdout = strings.ReplaceAll(expectedStdout, "1.99.0", version.LastKnownEnvoy.String())
			expectedStdout = strings.ReplaceAll(expectedStdout, "1.99", version.LastKnownEnvoyMinor.String())
			if runtime.GOOS == moreos.OSWindows {
				expectedStdout = strings.ReplaceAll(expectedStdout, "/", "\\")
				// As most maintainers don't use Windows, it is easier to revert piece-wise
				for _, original := range []string{
					globals.DefaultEnvoyVersionsURL,
					globals.DefaultEnvoyVersionsSchemaURL,
					"darwin/arm64",
					"$GOOS/$GOARCH",
				} {
					toRevert := strings.ReplaceAll(original, "/", "\\")
					expectedStdout = strings.ReplaceAll(expectedStdout, toRevert, original)
				}
			}
			require.Equal(t, expectedStdout, stdout.String())
		})
	}
}
