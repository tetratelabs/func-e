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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/globals"
)

func TestFuncEHelp(t *testing.T) {
	for _, command := range []string{"", "use", "versions", "run"} {
		command := command
		t.Run(command, func(t *testing.T) {
			c, stdout, _ := newApp(&globals.GlobalOpts{Version: "1.0", EnvoyVersion: "1.99.0"})
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
			require.Equal(t, string(bytes), stdout.String())
		})
	}
}
