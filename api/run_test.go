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

func TestRunWithCtxDone(t *testing.T) {
	tmpDir := t.TempDir()
	envoyVersion := version.LastKnownEnvoy
	versionsServer := test.RequireEnvoyVersionsTestServer(t, envoyVersion)
	defer versionsServer.Close()
	envoyVersionsURL := versionsServer.URL + "/envoy-versions.json"
	// Run the same test multiple times to ensure that the Envoy process is cleaned up properly with the context cancellation
	// in conjunction with the exit channel.
	for i := range 5 {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			exitCh := make(chan struct{}, 1)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			out := &bytes.Buffer{}
			// This will return right after the context is done, but the Envoy process itself is running another goroutine,
			// so without using the exit channel, we might end up existing the test (or main program) before the Envoy process receives
			// the signal to exit, hence it might end up being a zombie process.
			err := Run(ctx, []string{
				"--config-yaml", "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 9901}}}",
			}, Out(out), HomeDir(tmpDir), EnvoyVersionsURL(envoyVersionsURL), ExitChannel(exitCh))
			require.NoError(t, err)
			require.NotContains(t, out.String(), "Address already in use")
			<-exitCh // Wait for the Envoy process to completely exit.
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

	err := Run(ctx, []string{"--version"}, Out(b), HomeDir(tmpDir), EnvoyVersionsURL(envoyVersionsURL))
	require.NoError(t, err)

	require.NotEqual(t, 0, b.Len())
	_, err = os.Stat(filepath.Join(tmpDir, "versions"))
	require.NoError(t, err)

}
