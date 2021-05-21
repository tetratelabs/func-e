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
	if err != nil {
		require.NoError(t, err, `ioutil.TempDir("", "") erred`)
	}
	d, err = filepath.EvalSymlinks(d)
	require.NoError(t, err, `filepath.EvalSymlinks(%s) erred`, d)
	require.NotEmpty(t, d, `filepath.EvalSymlinks(%s) returned ""`)
	return d, func() {
		e := os.RemoveAll(d)
		require.NoError(t, e, `error removing directory: %v`, d)
	}
}

var (
	// captureScript is a test script used for capturing arguments and signals. If it receives an argument "exit=N",
	// where N is a code number, the script exits with that status. Otherwise it sleeps until a signal interrupts it.
	//go:embed testdata/capture.sh
	captureScript []byte
)

// RequireCaptureScript creates a copy of cmd.CaptureScript with the given basename. The function returned cleans it up.
func RequireCaptureScript(t *testing.T, name string) (string, func()) {
	d, cleanup := RequireNewTempDir(t)
	path := filepath.Join(d, name)
	err := ioutil.WriteFile(path, captureScript, 0700) //nolint:gosec
	if err != nil {
		cleanup()
		t.Fatalf(`expected no creating capture script %q: %v`, path, err)
	}
	return path, cleanup
}
