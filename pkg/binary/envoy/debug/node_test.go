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

package debug

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoytest"
)

func TestEnableNodeCollection(t *testing.T) {
	r, err := envoy.NewRuntime(EnableNodeCollection)
	require.NoError(t, err, "error creating envoy runtime")
	defer os.RemoveAll(r.DebugStore())

	envoytest.RequireRunTerminate(t, r, envoytest.RunKillOptions{})

	files := [...]string{"node/ps.txt", "node/network_interface.json", "node/connections.json"}
	for _, file := range files {
		path := filepath.Join(r.DebugStore(), file)
		f, err := os.Stat(path)
		require.NoError(t, err, "error stating %v", path)
		require.NotEmpty(t, f.Size(), "file %v was empty", path)

		if strings.HasSuffix(file, ".json") {
			raw, err := os.ReadFile(path)
			require.NoError(t, err, "error to read the file %v", path)
			var is []interface{}

			err = json.Unmarshal(raw, &is)
			require.NoError(t, err, "error to unmarshal json string, %v: \"%v\"", err, raw)
			require.NotEmpty(t, len(is), "unmarshaled content is empty, expected to be a non-empty array: \"%v\"", raw)
		}
	}
}
