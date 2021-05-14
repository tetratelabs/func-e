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
)

func TestGetEnvoyDoc(t *testing.T) {
	c, stdout, stderr := newApp(&globals.GlobalOpts{})
	err := c.Run([]string{"getenvoy", "doc"})
	require.NoError(t, err)
	require.Empty(t, stderr)

	want, err := os.ReadFile(filepath.Join("testdata", "getenvoy.md"))
	require.NoError(t, err)
	require.Equal(t, strings.ReplaceAll(string(want), "ENVOY_VERSION", reference.Latest), stdout.String())
}
