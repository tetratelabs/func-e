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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/reference"
	"github.com/tetratelabs/getenvoy/pkg/globals"
)

func TestGetEnvoyHelp(t *testing.T) {
	for _, command := range []string{"", "fetch", "list", "run"} {
		command := command
		t.Run(command, func(t *testing.T) {
			c, stdout, _ := newApp(&globals.GlobalOpts{})
			args := []string{"help"}
			if command != "" {
				args = append(args, command)
			}
			c.SetArgs(args)
			require.NoError(t, c.Execute())

			expected := "getenvoy_help.txt"
			if command != "" {
				expected = fmt.Sprintf("getenvoy_%s_help.txt", command)
			}

			want, err := os.ReadFile(filepath.Join("testdata", expected))
			require.NoError(t, err)
			require.Equal(t, strings.Replace(string(want), "VERSION", reference.Latest, -1), stdout.String())
		})
	}
}
