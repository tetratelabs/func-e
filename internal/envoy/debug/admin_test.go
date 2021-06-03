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
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/envoy"
	"github.com/tetratelabs/getenvoy/internal/globals"
	"github.com/tetratelabs/getenvoy/internal/test"
	"github.com/tetratelabs/getenvoy/internal/test/morerequire"
)

func TestEnableEnvoyAdminDataCollection(t *testing.T) {
	runDir, removeRunDir := morerequire.RequireNewTempDir(t)
	defer removeRunDir()

	mockAdmin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("junk"))
	}))
	defer mockAdmin.Close()

	adminPath := filepath.Join(runDir, "admin-address.txt")
	err := os.WriteFile(adminPath, []byte(mockAdmin.Listener.Addr().String()), 0600)
	require.NoError(t, err)

	runAndTerminateWithDebug(t, runDir, enableEnvoyAdminDataCollection, `--admin-address-path`, adminPath)

	for _, filename := range adminAPIPaths {
		path := filepath.Join(runDir, filename)
		f, err := os.Stat(path)
		require.NoError(t, err, "error stating %v", path)
		require.NotEmpty(t, f.Size(), "file %v was empty", path)
	}
}

// runAndTerminateWithDebug is like RequireRunTerminate, except returns a directory populated by the debug plugin.
func runAndTerminateWithDebug(t *testing.T, runDir string, debug func(r *envoy.Runtime) error, args ...string) error {
	fakeEnvoy := filepath.Join(runDir, "envoy")
	morerequire.RequireCaptureScript(t, fakeEnvoy)

	o := &globals.RunOpts{EnvoyPath: fakeEnvoy, RunDir: runDir, DontArchiveRunDir: true}

	stderr := new(bytes.Buffer)
	r := envoy.NewRuntime(o)
	r.Out = io.Discard
	r.Err = stderr
	require.NoError(t, debug(r))

	return test.RequireRunTerminate(t, func() {
		fakeInterrupt := r.FakeInterrupt
		if fakeInterrupt != nil {
			fakeInterrupt()
		}
	}, r, stderr, args...)
}
