// Copyright 2019 Tetrate
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

package shutdown

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodeCollection(t *testing.T) {
	runDir := t.TempDir()

	stderr, err := runWithShutdownHook(t, runDir, enableNodeCollection)
	require.NoError(t, err)

	files := [...]string{"node/ps.txt", "node/network_interface.json", "node/connections.json"}
	for _, file := range files {
		path := filepath.Join(runDir, file)
		f, err := os.Stat(path)
		require.NoError(t, err, "error stating %v: %s", path, stderr)

		// While usually not, ps can be empty due to deadline timeout getting a ps listing. Instead of flakey tests, we
		// don't enforce this.
		if file != "node/ps.txt" {
			require.NotEmpty(t, f.Size(), "file %v was empty: %s", path, stderr)
		}

		if strings.HasSuffix(file, ".json") {
			raw, err := os.ReadFile(path)
			require.NoError(t, err, "error to read the file %v: %s", path, stderr)
			var is []interface{}

			err = json.Unmarshal(raw, &is)
			require.NoError(t, err, "error to unmarshal json string, %v: \"%v\": %s", err, raw, stderr)
			require.NotEmpty(t, is, "unmarshalled content is empty, expected to be a non-empty array: \"%v\": %s", raw, stderr)
		}
	}
}
