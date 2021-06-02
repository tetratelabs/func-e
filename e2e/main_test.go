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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

//nolint:golint
const (
	getenvoyBinaryEnvKey   = "E2E_GETENVOY_BINARY"
	envoyVersionsURLEnvKey = "ENVOY_VERSIONS_URL"
	envoyVersionsJSON      = "../site/envoy-versions.json"
)

var expectedMockHeaders = map[string]string{"User-Agent": "getenvoy/dev"}

// TestMain ensures the "getenvoy" binary is valid.
func TestMain(m *testing.M) {
	// As this is an e2e test, we execute all tests with a binary compiled earlier.
	path, err := readGetEnvoyPath()
	if err != nil {
		exitOnInvalidBinary(err)
	}
	getEnvoyPath = path

	versionLine, _, err := getEnvoy("--version").exec()
	if err != nil {
		exitOnInvalidBinary(err)
	}

	if _, ok := os.LookupEnv(envoyVersionsURLEnvKey); !ok && strings.Contains(versionLine, "SNAPSHOT") {
		s, err := mockEnvoyVersionsServer() // no defer s.Close() because os.Exit() subverts it
		if err != nil {
			fmt.Fprintf(os.Stderr, `failed to serve %s: %v`, envoyVersionsJSON, err)
			os.Exit(1)
		}
		os.Setenv(envoyVersionsURLEnvKey, s.URL)
	}
	os.Exit(m.Run())
}

func exitOnInvalidBinary(err error) {
	fmt.Fprintf(os.Stderr, `failed to start e2e tests due to an invalid "getenvoy" binary: %v`, err)
	os.Exit(1)
}

// mockEnvoyVersionsServer ensures envoyVersionsURLEnvKey is set appropriately, so that non-release versions can see
// changes to local envoyVersionsJSON.
func mockEnvoyVersionsServer() (*httptest.Server, error) {
	f, err := os.Open(envoyVersionsJSON)
	if err != nil {
		return nil, err
	}

	defer f.Close() // nolint
	bytes, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ensure e2e tests won't eventually interfere with analytics when run against a release version
		for k, v := range expectedMockHeaders {
			h := r.Header.Get(k)
			if h != v {
				w.WriteHeader(500)
				w.Write([]byte(fmt.Sprintf(`invalid %q: %s != %s\n`, k, h, v))) //nolint
				return
			}
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(bytes) //nolint
	}))
	return ts, nil
}

// readGetEnvoyPath reads E2E_GETENVOY_BINARY or defaults to "$PWD/dist/getenvoy_$GOOS_$GOARCH/getenvoy"
// An error is returned if the value isn't an executable file.
func readGetEnvoyPath() (string, error) {
	path := os.Getenv(getenvoyBinaryEnvKey)
	if path == "" {
		// Assemble the default created by "make bin"
		relativePath := filepath.Join("..", "dist", fmt.Sprintf("getenvoy_%s_%s", runtime.GOOS, runtime.GOARCH), "getenvoy")
		abs, err := filepath.Abs(relativePath)
		if err != nil {
			return "", fmt.Errorf("%s didn't resolve to a valid path. Correct environment variable %s", path, getenvoyBinaryEnvKey)
		}
		path = abs
	}

	stat, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return "", fmt.Errorf("%s doesn't exist. Correct environment variable %s", path, getenvoyBinaryEnvKey)
	}
	if stat.IsDir() {
		return "", fmt.Errorf("%s is not a file. Correct environment variable %s", path, getenvoyBinaryEnvKey)
	}
	// While "make bin" should result in correct permissions, double-check as some tools lose them, such as
	// https://github.com/actions/upload-artifact#maintaining-file-permissions-and-case-sensitive-files
	if stat.Mode()&0111 == 0 {
		return "", fmt.Errorf("%s is not executable. Correct environment variable %s", path, getenvoyBinaryEnvKey)
	}
	return path, nil
}
