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
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/test"
	"github.com/tetratelabs/func-e/internal/version"
)

var (
	runArgs = []string{"--version"}
)

func TestRunWithCtxDone(t *testing.T) {
	tmpDir := t.TempDir()
	envoyVersion := version.LastKnownEnvoy
	versionsServer := test.RequireEnvoyVersionsTestServer(t, envoyVersion)
	defer versionsServer.Close()
	envoyVersionsURL := versionsServer.URL + "/envoy-versions.json"

	err := Run(t.Context(), runArgs, HomeDir(tmpDir), EnvoyVersionsURL(envoyVersionsURL))
	require.NoError(t, err)
	// Run the same test multiple times to ensure that the Envoy process is cleaned up properly with the context cancellation.
	for i := range 10 {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			start := time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			// This will return right after the context is done.
			err := Run(ctx, []string{
				"--log-level", "info",
				"--config-yaml", "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 9901}}}",
			}, HomeDir(tmpDir), EnvoyVersionsURL(envoyVersionsURL))
			require.Greater(t, time.Since(start).Seconds(), 2.0)
			require.NoError(t, err) // If the address is already in use, the exit code will be 1.
		})
	}
}

func TestRunToCompletion(t *testing.T) {

	tmpDir := t.TempDir()
	envoyVersion := version.LastKnownEnvoy
	versionsServer := test.RequireEnvoyVersionsTestServer(t, envoyVersion)
	defer versionsServer.Close()
	envoyVersionsURL := versionsServer.URL + "/envoy-versions.json"
	b := bytes.NewBufferString("")

	require.Equal(t, 0, b.Len())

	ctx := context.Background()
	// Set a large ctx timeout value
	ctx, cancel := context.WithTimeout(ctx, 1000*time.Minute)
	defer cancel()

	err := Run(ctx, runArgs, Out(b), HomeDir(tmpDir), EnvoyVersionsURL(envoyVersionsURL))
	require.NoError(t, err)

	require.NotEqual(t, 0, b.Len())
	_, err = os.Stat(filepath.Join(tmpDir, "versions"))
	require.NoError(t, err)

}
