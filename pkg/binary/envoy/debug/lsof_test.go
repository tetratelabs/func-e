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
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/binary/envoytest"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

func TestEnableOpenFilesDataCollection(t *testing.T) {
	debugDir, removeDebugDir := morerequire.RequireNewTempDir(t)
	defer removeDebugDir()

	workingDir := envoytest.RunAndTerminateWithDebug(t, debugDir, EnableOpenFilesDataCollection)

	file := "lsof/lsof.json"
	path := filepath.Join(workingDir, file)
	f, err := os.Stat(path)
	require.NoError(t, err, "error stating %v", path)

	if runtime.GOOS == `darwin` { // process.OpenFiles in unsupported, so this feature won't work
		require.Empty(t, f.Size(), "file %v was not empty", path)
	} else {
		require.NotEmpty(t, f.Size(), "file %v was empty", path)
		raw, err := os.ReadFile(path)
		require.NoError(t, err, "error reading file %v", path)

		var is []interface{}
		err = json.Unmarshal(raw, &is)
		require.NoError(t, err, "error to unmarshal json string, %v: \"%v\"", err, raw)
		require.NotEmpty(t, len(is), "unmarshalled content is empty, expected to be a non-empty array: \"%v\"", raw)
	}
}
