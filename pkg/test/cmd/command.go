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

package cmd

import (
	"bytes"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/cmd"
	"github.com/tetratelabs/getenvoy/pkg/common"
	builtintoolchain "github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain/builtin"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
	manifesttest "github.com/tetratelabs/getenvoy/pkg/test/manifest"
	"github.com/tetratelabs/getenvoy/pkg/types"
)

// FakeDockerDir includes "docker" which only executes the output. This means it doesn't really invoke docker.
//
// TODO: fake via exec.Run in unit tests because it is less complicated and error-prone than faking via shell scripts.
const FakeDockerDir = "../../../extension/workspace/toolchain/builtin/testdata/toolchain"

// NewRootCommand initializes a command with buffers for stdout and stderr.
func NewRootCommand() (c *cobra.Command, stdout, stderr *bytes.Buffer) {
	stdout = new(bytes.Buffer)
	stderr = new(bytes.Buffer)
	c = cmd.NewRoot()
	c.SetOut(stdout)
	c.SetErr(stderr)
	return c, stdout, stderr
}

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

// OverrideGetCurrentUser sets builtin.GetCurrentUser to return the indicated user.
// The function returned reverts to the original.
func OverrideGetCurrentUser(u *user.User) func() {
	previous := builtintoolchain.GetCurrentUser
	builtintoolchain.GetCurrentUser = func() (*user.User, error) {
		return u, nil
	}
	return func() {
		builtintoolchain.GetCurrentUser = previous
	}
}

// OverrideHomeDir sets common.HomeDir to return the indicated path. The function returned reverts to the original.
func OverrideHomeDir(homeDir string) func() {
	previous := common.HomeDir
	common.HomeDir = homeDir
	return func() {
		common.HomeDir = previous
	}
}

// RequireManifestPlatform returns the current platform as used in manifests.
func RequireManifestPlatform(t *testing.T) string {
	key, err := manifest.NewKey("standard:1.17.0")
	require.NoError(t, err, `error resolving manifest for key: %s`, key)
	return key.Platform
}

// RequireManifestTestServer calls manifest.SetURL to a test new tests server.
// The function returned stops that server and calls manifest.SetURL with the original URL.
func RequireManifestTestServer(t *testing.T, envoySubstituteArchiveDir string) func() {
	testManifest, err := manifesttest.NewSimpleManifest("standard:1.17.0", "wasm:1.15", "wasm:stable")

	require.NoError(t, err, `error creating test manifest`)

	manifestServer := manifesttest.NewServer(&manifesttest.ServerOpts{
		Manifest: testManifest,
		GetArtifactDir: func(uri string) (string, error) {
			ref, e := types.ParseReference(uri)
			if e != nil {
				return "", e
			}
			if ref.Flavor == "wasm" {
				return envoySubstituteArchiveDir, nil
			}
			if ref.Flavor == "standard" {
				ver, e := semver.NewVersion(ref.Version)
				if e == nil && ver.Major() >= 1 && ver.Minor() >= 17 {
					return envoySubstituteArchiveDir, nil
				}
			}
			return "", errors.Errorf("unexpected version of Envoy %q", uri)
		},
		OnError: func(err error) {
			require.NoError(t, err, `unexpected error from test manifest server`)
		},
	})

	// override location of the GetEnvoy manifest
	previous := manifest.GetURL()
	u := manifestServer.GetManifestURL()
	err = manifest.SetURL(u)
	require.NoError(t, err, `error manifest URL to: %s`, u)

	return func() {
		e := manifest.SetURL(previous)
		manifestServer.Close() // before require to ensure this occurs
		require.NoError(t, e, `error reverting manifest URL to: %s`, previous)
	}
}

// InitFakeEnvoyHome creates "$envoyHome/bin/envoy", which echos the commandline, output and stderr. It returns the
// path to the fake envoy script.
//
// "$envoyHome/bin/envoy" also copies any contents in current working directory to "$envoyHome/capture" when invoked.
//
// The capture is necessary because "$envoyHome/bin/envoy" is executed from a getenvoy-managed temp directory, deleted
// on exit. This directory defines how envoy would have run, so we need to save off contents in order to verify them.
//
// TODO: fake via exec.Run in unit tests because it is less complicated and error-prone than faking via shell scripts.
func InitFakeEnvoyHome(t *testing.T, envoyHome string) string {
	// Setup $envoyHome/bin and $envoyHome/capture
	_ = os.Mkdir(envoyHome, fs.ModePerm)
	envoyBin := filepath.Join(envoyHome, "bin")
	envoyCapture := filepath.Join(envoyHome, "capture")
	for _, dir := range []string{envoyBin, envoyCapture} {
		err := os.Mkdir(dir, fs.ModePerm)
		require.NoError(t, err, `couldn't create directory: %s`, dir)
	}

	// Create script literal of $envoyHome/bin/envoy which copies the current directory to $envoyCapture when invoked.
	// stdout and stderr are prefixed "envoy " to differentiate them from other command output, namely docker.
	fakeEnvoyScript := fmt.Sprintf(`#!/bin/sh
set -ue
# Copy all files in the cwd to the capture directory.
cp -r . "%s"

# Echo invocation context to stdout and fake stderr to ensure it is not combined into stdout.
echo envoy pwd: $PWD
echo envoy bin: $0
echo envoy args: $@
echo >&2 envoy stderr
`, envoyCapture)

	// Write $envoyHome/bin/envoy and ensure it is executable
	fakeEnvoyPath := filepath.Join(envoyBin, "envoy")
	err := ioutil.WriteFile(fakeEnvoyPath, []byte(fakeEnvoyScript), 0700) // nolint:gosec
	require.NoError(t, err, `couldn't create fake envoy script: %s`, fakeEnvoyPath)
	return fakeEnvoyPath
}

// ParseEnvoyWorkDirectory returns the CWD captured by the script generated by InitFakeEnvoyHome.
func ParseEnvoyWorkDirectory(stdout *bytes.Buffer) string {
	re := regexp.MustCompile(`.*envoy pwd: (.*)\n.*`)
	envoyWd := re.FindStringSubmatch(stdout.String())[1]
	return envoyWd
}
