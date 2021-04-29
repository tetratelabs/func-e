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
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/binary/envoytest"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

func TestEnableEnvoyAdminDataCollection(t *testing.T) {
	debugDir, removeDebugDir := morerequire.RequireNewTempDir(t)
	defer removeDebugDir()

	mockAdmin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("junk"))
	}))
	defer mockAdmin.Close()

	adminPath := filepath.Join(debugDir, "admin-address.txt")
	err := os.WriteFile(adminPath, []byte(mockAdmin.Listener.Addr().String()), 0600)
	require.NoError(t, err)

	workingDir := envoytest.RunAndTerminateWithDebug(t, debugDir,
		EnableEnvoyAdminDataCollection, `--admin-address-path`, adminPath)

	for _, filename := range adminAPIPaths {
		path := filepath.Join(workingDir, filename)
		f, err := os.Stat(path)
		require.NoError(t, err, "error stating %v", path)
		require.NotEmpty(t, f.Size(), "file %v was empty", path)
	}
}
