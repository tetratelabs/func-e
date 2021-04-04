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
	"strings"
	"testing"

	"github.com/otiai10/copy"
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

// RequireSetenv will os.Setenv the given key and value. The function returned reverts to the original.
func RequireSetenv(t *testing.T, key, value string) func() {
	previous := os.Getenv(key)
	err := os.Setenv(key, value)
	require.NoError(t, err, `error setting env variable %s=%s`, key, value)
	return func() {
		e := os.Setenv(key, previous)
		require.NoError(t, e, `error reverting env variable %s=%s`, key, previous)
	}
}

// RequireChDir will os.Chdir into the indicated dir, panicing on any problem.
// The string returned is the absolute path corresponding to the input. The function returned reverts to the original.
func RequireChDir(t *testing.T, d string) (string, func()) {
	dir := RequireAbsDir(t, d)

	// Save previous working directory to that it can be reverted later.
	previous, err := os.Getwd()
	require.NoError(t, err, `error determining current directory`)

	// Now, actually change to the directory.
	err = os.Chdir(d)
	require.NoError(t, err, `error changing to directory: %v`, d)
	return dir, func() {
		e := os.Chdir(previous)
		require.NoError(t, e, `error changing to directory: %v`, previous)
	}
}

// RequireCopyOfDir creates a new directory which is a copy of the template. The function returned cleans it up.
func RequireCopyOfDir(t *testing.T, templateDir string) (string, func()) {
	d, cleanup := RequireNewTempDir(t)
	err := copy.Copy(templateDir, d)
	require.NoError(t, err, `expected no error copying %s to %s`, templateDir, d)
	return d, cleanup
}

// RequireAbsDir runs filepath.Abs and ensures there are no errors and the input is a directory.
func RequireAbsDir(t *testing.T, d string) string {
	dir, err := filepath.Abs(d)
	require.NoError(t, err, `error determining absolute directory: %v`, d)
	require.DirExists(t, dir, `directory doesn't exist': %v`, dir)
	return dir
}

// RequireOverridePath will prefix os.Setenv with the indicated dir, panicing on any problem.
// The string returned is the absolute path corresponding to the input. The function returned reverts to the original.
func RequireOverridePath(t *testing.T, d string) (string, func()) {
	dir := RequireAbsDir(t, d)

	// Save previous path to that it can be reverted later.
	previous := os.Getenv("PATH")

	// Place the resolved directory in from of the previous path
	path := strings.Join([]string{dir, previous}, string(filepath.ListSeparator))

	// Now, actually change the PATH env
	err := os.Setenv("PATH", path)
	require.NoError(t, err, `error setting PATH to: %v`, path)
	return dir, func() {
		e := os.Setenv("PATH", previous)
		require.NoError(t, e, `error reverting to PATH: %v`, previous)
	}
}
