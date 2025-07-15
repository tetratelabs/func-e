// Copyright 2021 Tetrate
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

package e2e

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/tetratelabs/func-e/internal/moreos"
)

const (
	envoyVersionsURLEnvKey = "ENVOY_VERSIONS_URL"
	envoyVersionsJSON      = "envoy-versions.json"
)

var expectedMockHeaders = map[string]string{"User-Agent": "func-e/dev"}

// TestMain ensures the "func-e" binary is valid.
func TestMain(m *testing.M) {
	// As this is an e2e test, we execute all tests with a binary compiled earlier.
	if err := readOrBuildFuncEBin(); err != nil {
		exitOnInvalidBinary(err)
	}

	// pre-flight check the binary is usable
	versionLine, _, err := funcEExec("--version")
	if err != nil {
		exitOnInvalidBinary(err)
	}

	// Allow local file override when a SNAPSHOT version
	if _, err := os.Stat(envoyVersionsJSON); err == nil && strings.Contains(versionLine, "SNAPSHOT") {
		s, err := mockEnvoyVersionsServer() // no defer s.Close() because os.Exit() subverts it
		if err != nil {
			moreos.Fprintf(os.Stderr, "failed to serve %s: %v\n", envoyVersionsJSON, err)
			os.Exit(1)
		}
		os.Setenv(envoyVersionsURLEnvKey, s.URL) //nolint:errcheck
	}
	os.Exit(m.Run())
}

func exitOnInvalidBinary(err error) {
	moreos.Fprintf(os.Stderr, `failed to start e2e tests due to an invalid "func-e" binary: %v\n`, err)
	os.Exit(1)
}

// mockEnvoyVersionsServer ensures envoyVersionsURLEnvKey is set appropriately, so that non-release versions can see
// changes to local envoyVersionsJSON.
func mockEnvoyVersionsServer() (*httptest.Server, error) {
	f, err := os.Open(envoyVersionsJSON)
	if err != nil {
		return nil, err
	}

	defer f.Close() //nolint:errcheck
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ensure e2e tests won't eventually interfere with analytics when run against a release version
		for k, v := range expectedMockHeaders {
			h := r.Header.Get(k)
			if h != v {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(moreos.Sprintf("invalid %q: %s != %s\n", k, h, v))) //nolint
				return
			}
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(b) //nolint
	}))
	return ts, nil
}
