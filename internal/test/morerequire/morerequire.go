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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

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

// RequireSetMtime sets the mtime of the dir given a string formatted date. Ex "2006-01-02"
func RequireSetMtime(t *testing.T, dir, date string) {
	td, err := time.Parse("2006-01-02", date)
	require.NoError(t, err)
	require.NoError(t, os.Chtimes(dir, td, td))
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

// RequireChdirIntoTemp creates a new temp directory and cleans it up on with the returned function.
func RequireChdirIntoTemp(t *testing.T) func() {
	wd, err := os.Getwd()
	require.NoError(t, err)

	tempDir, removeTempDir := RequireNewTempDir(t)
	if err = os.Chdir(tempDir); err != nil {
		removeTempDir() // don't leak
		require.NoError(t, err)
	}

	return func() {
		require.NoError(t, os.Chdir(wd))
		removeTempDir()
	}
}
