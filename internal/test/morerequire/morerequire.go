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

// Package morerequire includes more require functions than "github.com/stretchr/testify/require"
// Do not add dependencies on any main code as it will cause cycles.
package morerequire

import (
	// Embedding the capture script is easier than file I/O each time it is used.
	_ "embed"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// RequireNewTempDir creates a new directory. The function returned cleans it up.
func RequireNewTempDir(t *testing.T) (string, func()) {
	d, err := ioutil.TempDir("", "")
	require.NoError(t, err, `ioutil.TempDir("", "") erred`)
	d, err = filepath.EvalSymlinks(d)
	require.NoError(t, err, `filepath.EvalSymlinks(%s) erred`, d)
	return d, func() {
		os.RemoveAll(d) //nolint
	}
}

var (
	// captureScript is a test script used for capturing arguments and signals. If it receives an argument "exit=N",
	// where N is a code number, the script exits with that status. Otherwise it sleeps until a signal interrupts it.
	//go:embed testdata/capture.sh
	captureScript []byte
)

// RequireCaptureScript writes captureScript to the given path
func RequireCaptureScript(t *testing.T, path string) {
	require.NoError(t, os.WriteFile(path, captureScript, 0700)) //nolint:gosec
}

// RequireSetenv will os.Setenv the given key and value. The function returned reverts to the original.
func RequireSetenv(t *testing.T, key, value string) func() {
	previous := os.Getenv(key)
	err := os.Setenv(key, value)
	require.NoError(t, err, `error setting env variable %s=%s`, key, value)
	return func() {
		if previous != "" {
			err = os.Setenv(key, previous)
		} else {
			err = os.Unsetenv(key)
		}
		require.NoError(t, err, `error reverting env variable %s=%s`, key, previous)
	}
}
