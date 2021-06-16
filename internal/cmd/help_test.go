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

	"github.com/tetratelabs/getenvoy/internal/globals"
	"github.com/tetratelabs/getenvoy/internal/version"
)

func TestGetEnvoyHelp(t *testing.T) {
	for _, command := range []string{"", "use", "versions", "run"} {
		command := command
		t.Run(command, func(t *testing.T) {
			c, stdout, _ := newApp(&globals.GlobalOpts{})
			args := []string{"getenvoy"}
			if command != "" {
				args = []string{"getenvoy", "help", command}
			}
			require.NoError(t, c.Run(args))

			expected := "getenvoy_help.txt"
			if command != "" {
				expected = fmt.Sprintf("getenvoy_%s_help.txt", command)
			}

			bytes, err := os.ReadFile(filepath.Join("testdata", expected))
			want := strings.ReplaceAll(string(bytes), "{VERSION}", string(version.GetEnvoy))
			want = strings.ReplaceAll(want, "{ENVOY_VERSION}", string(version.LastKnownEnvoy))

			require.NoError(t, err)
			require.Equal(t, want, stdout.String())
		})
	}
}
