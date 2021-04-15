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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoytest"
)

func TestEnableEnvoyLogCollection(t *testing.T) {
	r, err := envoy.NewRuntime(EnableEnvoyLogCollection)
	require.NoError(t, err, "error getting envoy runtime")
	defer os.RemoveAll(r.DebugStore())

	envoytest.RequireRunTerminate(t, r, "")

	// We expect to see logs because when envoy starts up, the status checker will make HTTP requests to /ready
	for _, filename := range []string{"logs/access.log", "logs/error.log"} {
		path := filepath.Join(r.DebugStore(), filename)
		f, err := os.Stat(path)
		require.NoError(t, err, "error stating %v", path)
		require.NotEmpty(t, f.Size(), "file %v was empty", path)
	}
}
