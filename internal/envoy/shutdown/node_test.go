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

	"github.com/tetratelabs/func-e/internal/test/morerequire"
)

func TestEnableNodeCollection(t *testing.T) {
	runDir, removeRunDir := morerequire.RequireNewTempDir(t)
	defer removeRunDir()

	require.NoError(t, runWithShutdownHook(t, runDir, enableNodeCollection))

	files := [...]string{"node/ps.txt", "node/network_interface.json", "node/connections.json"}
	for _, file := range files {
		path := filepath.Join(runDir, file)
		f, err := os.Stat(path)
		require.NoError(t, err, "error stating %v", path)

		// While usually not, ps can be empty due to deadline timeout getting a ps listing. Instead of flakey tests, we
		// don't enforce this.
		if file != "node/ps.txt" {
			require.NotEmpty(t, f.Size(), "file %v was empty", path)
		}

		if strings.HasSuffix(file, ".json") {
			raw, err := os.ReadFile(path)
			require.NoError(t, err, "error to read the file %v", path)
			var is []interface{}

			err = json.Unmarshal(raw, &is)
			require.NoError(t, err, "error to unmarshal json string, %v: \"%v\"", err, raw)
			require.NotEmpty(t, len(is), "unmarshalled content is empty, expected to be a non-empty array: \"%v\"", raw)
		}
	}
}
