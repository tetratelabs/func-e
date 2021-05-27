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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/test/morerequire"
	"github.com/tetratelabs/getenvoy/internal/version"
)

// TestGetEnvoyInstall needs to always execute, so we run it in a separate home directory
func TestGetEnvoyInstall(t *testing.T) {
	homeDir, removeHomeDir := morerequire.RequireNewTempDir(t)
	defer removeHomeDir()

	t.Run("not yet installed", func(t *testing.T) {
		stdout, stderr, err := getEnvoy("--home-dir", homeDir, "install", version.LastKnownEnvoy).exec()

		require.NoError(t, err)
		require.Regexp(t, `downloading https:.*tar.*`, stdout)
		require.Empty(t, stderr)

		require.FileExists(t, filepath.Join(homeDir, "versions", version.LastKnownEnvoy, "bin", "envoy"))
	})

	t.Run("already installed", func(t *testing.T) {
		stdout, stderr, err := getEnvoy("--home-dir", homeDir, "install", version.LastKnownEnvoy).exec()

		require.NoError(t, err)
		require.Equal(t, version.LastKnownEnvoy+" is already downloaded\n", stdout)
		require.Empty(t, stderr)
	})
}
