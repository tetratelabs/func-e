// Copyright 2020 Tetrate
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

package extension_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	rootcmd "github.com/tetratelabs/getenvoy/pkg/cmd"
	"github.com/tetratelabs/getenvoy/pkg/globals"
	"github.com/tetratelabs/getenvoy/pkg/test/cmd"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

// When unspecified, we default the tag to Docker's default "latest". Note: recent tools enforce qualifying this!
const defaultTag = "latest"

func TestGetEnvoyExtensionPushValidateFlag(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedErr string
	}{
		{
			name:        "WASM file doesn't exist",
			args:        []string{"non-existing-file", "localhost:55555/getenvoy/sample"},
			expectedErr: `invalid WASM file "non-existing-file": stat non-existing-file: no such file or directory`,
		},
		{
			name:        "WASM file is a directory",
			args:        []string{".", "localhost:55555/getenvoy/sample"},
			expectedErr: `WASM file argument was a directory "."`,
		},
		{
			name:        "Invalid image ref",
			args:        []string{"push_test.go", "/docker.io/"}, // fake the current test as the wasm file
			expectedErr: `invalid image reference: invalid reference format`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			// Run "getenvoy extension run" with the flags we are testing
			c, stdout, stderr := cmd.NewRootCommand(&globals.GlobalOpts{})
			c.SetArgs(append([]string{"extension", "push"}, test.args...))
			err := rootcmd.Execute(c)

			// Verify the command failed with the expected error
			require.EqualError(t, err, test.expectedErr, `expected an error running [%v]`, c)
			require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
			expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy extension push --help' for usage.\n", test.expectedErr)
			require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, c)
		})
	}
}

// TestGetEnvoyExtensionPush shows current directory is usable, provided it is a valid workspace.
func TestGetEnvoyExtensionPush(t *testing.T) {
	mock := mockRegistryServer(t)
	defer mock.Close()

	// localhost:5000/getenvoy/sample, not http://localhost:5000/getenvoy/sample
	localRegistryWasmImageRef := fmt.Sprintf(`%s/getenvoy/sample`, mock.Listener.Addr())

	tempDir, deleteTempDir := morerequire.RequireNewTempDir(t)
	defer deleteTempDir()

	// Write a fake wasm extension file
	wasmFile := filepath.Join(tempDir, "extension.wasm")
	wasmBytes := []byte{0x00, 0x61, 0x73, 0x6d}        // magic
	err := ioutil.WriteFile(wasmFile, wasmBytes, 0600) //nolint:gosec
	require.NoError(t, err)

	// Run "getenvoy extension push testdata/extension.wasm localhost:5000/getenvoy/sample"
	c, stdout, stderr := cmd.NewRootCommand(&globals.GlobalOpts{})
	c.SetArgs([]string{"extension", "push", wasmFile, localRegistryWasmImageRef, "--use-http", "true"})
	err = rootcmd.Execute(c)

	// A fully qualified image ref includes the tag
	imageRef := localRegistryWasmImageRef + ":" + defaultTag

	// Verify stdout shows the latest tag and the correct image ref
	require.NoError(t, err, `expected no error running [%v]`, c)
	require.Contains(t, stdout.String(), fmt.Sprintf(`Using default tag: %s
Pushed %s
digest: sha256`, defaultTag, imageRef), `unexpected stderr after running [%v]`, c)
	require.Empty(t, stderr, `expected no stderr running [%v]`, c)
}

// The tests above are unit tests, not end-to-end (e2e) tests. Hence, we use a mock registry instead of a real one.
func mockRegistryServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statusCode := 500

		body, err := io.ReadAll(r.Body) // fully read the request
		require.NoError(t, err, "Error reading body of %s %s", r.Method, r.URL.Path)

		switch r.Method {
		case "HEAD":
			if strings.Index(r.URL.Path, "/v2/getenvoy/sample/blobs") == 0 {
				statusCode = 404 // pretend it hasn't been uploaded, yet
			} else if r.URL.Path == "/v2/getenvoy/sample/manifests/latest" {
				statusCode = 404 // pretend there's no manifest either
			}
		case "POST":
			if r.URL.Path == "/v2/getenvoy/sample/blobs/uploads/" {
				statusCode = 202 // pretend we processed the data
				w.Header().Add("Location", "/upload")
			}
		case "PUT":
			if r.URL.Path == "/upload" {
				err := r.ParseForm()
				require.NoError(t, err, "Error parsing PUT %s", r.URL.Path)
				require.NotEmpty(t, r.Form.Get("digest"), `Expected PUT %s to have a query parameter "digest"`, r.URL.Path)

				w.Header().Add("Docker-Content-Digest", r.Form.Get("digest"))
				statusCode = 200 // Pretend we accepted the blob
			} else if r.URL.Path == "/v2/getenvoy/sample/manifests/latest" {
				w.Header().Add("Docker-Content-Digest", "sha256:"+hash(body))
				statusCode = 200 // Pretend we accepted the manifest
			}
		}
		w.WriteHeader(statusCode)
	}))
}

func hash(b []byte) string {
	h := sha256.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}
