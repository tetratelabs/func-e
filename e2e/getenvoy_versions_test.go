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

package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/test/morerequire"
	"github.com/tetratelabs/getenvoy/internal/version"
)

func TestGetEnvoyVersions_NothingYet(t *testing.T) {
	homeDir, removeHomeDir := morerequire.RequireNewTempDir(t)
	defer removeHomeDir()

	stdout, stderr, err := getEnvoy("--home-dir", homeDir, "versions").exec()

	require.NoError(t, err)
	require.Equal(t, `No Envoy versions, yet
`, stdout)
	require.Empty(t, stderr)
}

func TestGetEnvoyVersions_All(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := getEnvoy("versions", "-a").exec()

	require.Regexp(t, "^VERSION\tRELEASE_DATE\n", stdout)
	require.Regexp(t, fmt.Sprintf("%s\t202[1-9]-[01][0-9]-[0-3][0-9]\n", version.LastKnownEnvoy), stdout)
	require.Empty(t, stderr)
	require.NoError(t, err)
}

// TestGetEnvoyVersions verifies output is sorted
func TestGetEnvoyVersions(t *testing.T) {
	homeDir, removeHomeDir := morerequire.RequireNewTempDir(t)
	defer removeHomeDir()

	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, "versions", "1.16.1"), 0700))
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, "versions", "1.17.2"), 0700))

	stdout, stderr, err := getEnvoy("--home-dir", homeDir, "versions").exec()

	require.NoError(t, err)
	require.Equal(t, `VERSION
1.17.2
1.16.1
`, stdout)
	require.Empty(t, stderr)
}
