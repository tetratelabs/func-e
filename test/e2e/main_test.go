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

package e2e_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	e2e "github.com/tetratelabs/getenvoy/test/e2e/util"
)

// getEnvoy is the absolute path to the "getenvoy" binary used in all tests.
var getEnvoy = e2e.GetEnvoy

//nolint:golint
const E2E_GETENVOY_BINARY = "E2E_GETENVOY_BINARY"

// TestMain ensures the "getenvoy" binary is valid.
func TestMain(m *testing.M) {
	// As this is an e2e test, we execute all tests with a binary compiled earlier.
	path, err := getEnvoyPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, `failed to start e2e tests due to an invalid "getenvoy" binary: %v`, err)
		os.Exit(1)
	}
	e2e.GetEnvoyPath = path
	os.Exit(m.Run())
}

// getEnvoyPath reads E2E_GETENVOY_BINARY or defaults to "$PWD/build/bin/$GOOS/$GOARCH/getenvoy"
// An error is returned if the value isn't an executable file.
func getEnvoyPath() (string, error) {
	path := os.Getenv(E2E_GETENVOY_BINARY)
	if path == "" {
		// Assemble the default created by "make bin"
		relativePath := filepath.Join("..", "..", "build", "bin", runtime.GOOS, runtime.GOARCH, "getenvoy")
		abs, err := filepath.Abs(relativePath)
		if err != nil {
			return "", fmt.Errorf("%s didn't resolve to a valid path. Correct environment variable %s", path, E2E_GETENVOY_BINARY)
		}
		path = abs
	}
	stat, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return "", fmt.Errorf("%s doesn't exist. Correct environment variable %s", path, E2E_GETENVOY_BINARY)
	}
	if stat.IsDir() {
		return "", fmt.Errorf("%s is not a file. Correct environment variable %s", path, E2E_GETENVOY_BINARY)
	}
	// While "make bin" should result in correct permissions, double-check as some tools lose them, such as
	// https://github.com/actions/upload-artifact#maintaining-file-permissions-and-case-sensitive-files
	if stat.Mode()&0111 == 0 {
		return "", fmt.Errorf("%s is not executable. Correct environment variable %s", path, E2E_GETENVOY_BINARY)
	}
	return path, nil
}
