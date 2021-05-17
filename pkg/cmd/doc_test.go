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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/reference"
	"github.com/tetratelabs/getenvoy/pkg/globals"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

func TestGetEnvoyDoc(t *testing.T) {
	tempDir, deleteTempDir := morerequire.RequireNewTempDir(t)
	defer deleteTempDir()

	c, _, _ := newApp(&globals.GlobalOpts{})
	c.SetArgs([]string{"doc", "-o", tempDir, "-l", "/reference/"})
	require.NoError(t, c.Execute())

	for _, file := range []string{"getenvoy.md", "getenvoy_list.md", "getenvoy_fetch.md", "getenvoy_run.md"} {
		file := file // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(file, func(t *testing.T) {
			want, err := os.ReadFile(filepath.Join("testdata", file))
			require.NoError(t, err)
			have, err := os.ReadFile(filepath.Join(tempDir, file))
			require.NoError(t, err)
			require.Equal(t, strings.ReplaceAll(string(want), "VERSION", reference.Latest), string(have))
		})
	}
}
