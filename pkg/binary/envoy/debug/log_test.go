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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/binary/envoytest"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

func TestEnableEnvoyLogCollection(t *testing.T) {
	debugDir, removeDebugDir := morerequire.RequireNewTempDir(t)
	defer removeDebugDir()

	workingDir := envoytest.RunAndTerminateWithDebug(t, debugDir, EnableEnvoyLogCollection)

	for _, filename := range []string{"logs/access.log", "logs/error.log"} {
		require.FileExists(t, filepath.Join(workingDir, filename))
	}
}
