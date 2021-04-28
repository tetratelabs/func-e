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

package push_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/globals"
	rootcmd "github.com/tetratelabs/getenvoy/pkg/cmd"
	"github.com/tetratelabs/getenvoy/pkg/test/cmd"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

// relativeExtensionDir points to a usable pre-initialized workspace
const relativeExtensionDir = "testdata/workspace"

// When unspecified, we default the tag to Docker's default "latest". Note: recent tools enforce qualifying this!
const defaultTag = "latest"

// TestGetEnvoyExtensionPush shows current directory is usable, provided it is a valid workspace.
func TestGetEnvoyExtensionPush(t *testing.T) {
	mock := mockRegistryServer(t)
	defer mock.Close()

	// localhost:5000/getenvoy/sample, not http://localhost:5000/getenvoy/sample
	localRegistryWasmImageRef := fmt.Sprintf(`%s/getenvoy/sample`, mock.Listener.Addr())

	// "getenvoy extension clean" must be in a valid extension directory
	o := &globals.GlobalOpts{ExtensionDir: morerequire.RequireAbs(t, relativeExtensionDir)}

	// Run "getenvoy extension push localhost:5000/getenvoy/sample"
	c, stdout, stderr := cmd.NewRootCommand(o)
	c.SetArgs([]string{"extension", "push", localRegistryWasmImageRef, "--use-http", "true"})
	err := rootcmd.Execute(c)

	// A fully qualified image ref includes the tag
	imageRef := localRegistryWasmImageRef + ":" + defaultTag

	// Verify stdout shows the latest tag and the correct image ref
	require.NoError(t, err, `expected no error running [%v]`, c)
	require.Contains(t, stdout.String(), fmt.Sprintf(`Using default tag: %s
Pushed %s
digest: sha256`, defaultTag, imageRef), `unexpected stderr after running [%v]`, c)
	require.Empty(t, stderr, `expected no stderr running [%v]`, c)
}

func TestGetEnvoyExtensionPushFailsOutsideExtensionDirectory(t *testing.T) {
	mock := mockRegistryServer(t)
	defer mock.Close()

	// Change to a non-workspace dir
	o := &globals.GlobalOpts{ExtensionDir: morerequire.RequireAbs(t, ".")}

	// Run "getenvoy extension push localhost:5000/getenvoy/sample" (not http://localhost:5000/getenvoy/sample)
	c, stdout, stderr := cmd.NewRootCommand(o)
	c.SetArgs([]string{"extension", "push", fmt.Sprintf(`%s/getenvoy/sample`, mock.Listener.Addr())})
	err := rootcmd.Execute(c)

	// Verify the command failed with the expected error
	expectedErr := fmt.Sprintf("not an extension directory %q", o.ExtensionDir)
	require.EqualError(t, err, expectedErr, `expected an error running [%v]`, c)
	require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
	expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy extension push --help' for usage.\n", expectedErr)
	require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, c)
}

// TestGetEnvoyExtensionPushWithExplicitFileOption shows we don't need to be in a extension directory to push a wasm.
func TestGetEnvoyExtensionPushWithExplicitFileOption(t *testing.T) {
	mock := mockRegistryServer(t)
	defer mock.Close()

	// localhost:5000/getenvoy/sample, not http://localhost:5000/getenvoy/sample
	localRegistryWasmImageRef := fmt.Sprintf(`%s/getenvoy/sample`, mock.Listener.Addr())

	// Change to a non-workspace dir
	o := &globals.GlobalOpts{ExtensionDir: morerequire.RequireAbs(t, "testdata")}

	// Point to a wasm file explicitly
	wasm := filepath.Join(o.ExtensionDir, "workspace", "extension.wasm")

	// Run "getenvoy extension push localhost:5000/getenvoy/sample --extension-file testdata/workspace/extension.wasm"
	c, stdout, stderr := cmd.NewRootCommand(o)
	c.SetArgs([]string{"extension", "push", localRegistryWasmImageRef, "--extension-file", wasm, "--use-http", "true"})
	err := rootcmd.Execute(c)

	// Verify the pushed a latest tag to the correct registry
	require.NoError(t, err, `expected no error running [%v]`, c)
	require.Contains(t, stdout.String(), fmt.Sprintf(`Using default tag: latest
Pushed %s:latest
digest: sha256`, localRegistryWasmImageRef))
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
