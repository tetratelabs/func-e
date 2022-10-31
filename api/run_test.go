// Copyright 2022 Tetrate
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

package api

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/test"
	"github.com/tetratelabs/func-e/internal/version"
)

var (
	// minRunArgs is the minimal config needed to run Envoy 1.18+, non-windows <1.18 need access_log_path: '/dev/stdout'
	minRunArgs = []string{"--config-yaml", "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}"}
)

func TestRun(t *testing.T) {

	tmpDir := t.TempDir()
	envoyVersion := version.LastKnownEnvoy
	versionsServer := test.RequireEnvoyVersionsTestServer(t, envoyVersion)
	envoyVersionsURL := versionsServer.URL + "/envoy-versions.json"
	b := bytes.NewBufferString("")

	require.Equal(t, 0, b.Len())

	// 1. Pass a context with a timeout to Run to start running Envoy
	// 2. Wait until the context is done
	// 3. Ensure that the error is nil
	var err error
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	go func() {
		err = Run(ctx, minRunArgs, Out(b), HomeDir(tmpDir), EnvoyVersionsURL(envoyVersionsURL))
	}()
	<-ctx.Done()
	require.NoError(t, err)

	require.NotEqual(t, 0, b.Len())
	_, err = os.Stat(filepath.Join(tmpDir, "versions"))
	require.NoError(t, err)
	t.Cleanup(versionsServer.Close)
}
