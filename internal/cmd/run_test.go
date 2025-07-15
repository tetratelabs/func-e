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

package cmd_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	rootcmd "github.com/tetratelabs/func-e/internal/cmd"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/test/morerequire"
	"github.com/tetratelabs/func-e/internal/version"
)

// TestFuncERun takes care to not duplicate test/e2e/testrun.go, but still give some coverage.
func TestFuncERun(t *testing.T) {
	o := setupTest(t)

	// Override the Envoy path to use fake Envoy
	o.EnvoyPath = fakeEnvoyBin

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	c := rootcmd.NewApp(o)
	c.Writer = stdout
	c.ErrWriter = stderr

	args := []string{"func-e", "run", "--config-yaml", "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}"}

	// Create a context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure context is canceled when test completes

	// Create a buffer-based reader that implements io.ReadSeeker
	stderrBuf := new(bytes.Buffer)
	c.ErrWriter = io.MultiWriter(stderr, stderrBuf)

	// Run Envoy in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- c.RunContext(ctx, args)
	}()

	// Wait for Envoy to output that it's started
	require.Eventually(t, func() bool {
		return strings.Contains(stderr.String(), "starting main dispatch loop")
	}, 5*time.Second, 100*time.Millisecond, "Envoy didn't start within the expected time")

	// Cancel the context to stop Envoy
	cancel()

	// Wait for the command to complete
	err := <-errCh
	require.NoError(t, err)

	// Verify all key messages from fake Envoy appear in the correct order using regex
	stderrOutput := stderr.String()
	pattern := `(?s).*initializing epoch 0.*admin address:.*starting main dispatch loop.*`
	matched, err := regexp.MatchString(pattern, stderrOutput)
	require.NoError(t, err)
	require.True(t, matched, "Expected fake Envoy output sequence not found in stderr")
}

func TestFuncERun_TeesConsoleToLogs(t *testing.T) {
	o := setupTest(t)

	c, stdout, stderr := newApp(o)
	// ignore messages from func-e we only care about envoy
	o.Out = io.Discard
	o.DontArchiveRunDir = true // we need to read-back the log files
	runWithInvalidConfig(t, c)

	actual, err := os.ReadFile(filepath.Join(o.RunDir, "stdout.log"))
	require.NoError(t, err)
	require.Contains(t, stdout.String(), string(actual))

	actual, err = os.ReadFile(filepath.Join(o.RunDir, "stderr.log"))
	require.NoError(t, err)
	require.NotEmpty(t, stderr.String()) // sanity check
	require.Equal(t, stderr.String(), string(actual))
}

func TestFuncERun_ReadsHomeVersionFile(t *testing.T) {
	o := setupTest(t)
	o.EnvoyVersion = "" // pretend this is an initial setup
	o.Out = new(bytes.Buffer)

	require.NoError(t, os.WriteFile(filepath.Join(o.HomeDir, "version"), []byte(version.LastKnownEnvoyMinor), 0o600))

	c, _, _ := newApp(o)
	runWithInvalidConfig(t, c)

	// No implicit lookup
	require.NotContains(t, o.Out.(*bytes.Buffer).String(), moreos.Sprintf("looking up latest version"))
	require.Equal(t, version.LastKnownEnvoy, o.EnvoyVersion)

	writtenVersion, err := os.ReadFile(filepath.Join(o.HomeDir, "version"))
	require.NoError(t, err)
	require.Equal(t, version.LastKnownEnvoyMinor.String(), string(writtenVersion))
}

func TestFuncERun_CreatesHomeVersionFile(t *testing.T) {
	o := setupTest(t)
	o.EnvoyVersion = "" // pretend this is an initial setup
	o.Out = new(bytes.Buffer)

	// make sure first run where the home doesn't exist yet, works!
	require.NoError(t, os.RemoveAll(o.HomeDir))

	c, _, _ := newApp(o)
	runWithInvalidConfig(t, c)

	// We logged the implicit lookup
	require.Contains(t, o.Out.(*bytes.Buffer).String(), moreos.Sprintf("looking up the latest Envoy version"))
	require.FileExists(t, filepath.Join(o.HomeDir, "version"))
	require.Equal(t, version.LastKnownEnvoy, o.EnvoyVersion)

	writtenVersion, err := os.ReadFile(filepath.Join(o.HomeDir, "version"))
	require.NoError(t, err)
	require.Equal(t, version.LastKnownEnvoyMinor.String(), string(writtenVersion))
}

// runWithInvalidConfig intentionally has envoy quit. This allows tests to not have to interrupt envoy to proceed
func runWithInvalidConfig(t *testing.T, c *cli.App) {
	require.EqualError(t, c.Run([]string{"func-e", "run"}), "envoy exited with status: 1")
}

func TestFuncERun_ValidatesHomeVersion(t *testing.T) {
	o := setupTest(t)
	o.Out = new(bytes.Buffer)

	o.EnvoyVersion = ""
	require.NoError(t, os.WriteFile(filepath.Join(o.HomeDir, "version"), []byte("a.a.a"), 0o600))

	c, _, _ := newApp(o)
	err := c.Run([]string{"func-e", "run"})

	// Verify the command failed with the expected error
	expectedErr := fmt.Sprintf(`invalid version in "$FUNC_E_HOME/version": "a.a.a" should look like %q or %q`, version.LastKnownEnvoy, version.LastKnownEnvoyMinor)
	require.EqualError(t, err, moreos.ReplacePathSeparator(expectedErr))
}

// TestFuncERun_ValidatesWorkingVersion duplicates logic in version_test.go to ensure a non-home version validates.
func TestFuncERun_ValidatesWorkingVersion(t *testing.T) {
	o := setupTest(t)
	o.Out = new(bytes.Buffer)
	o.EnvoyVersion = ""

	revertWd := morerequire.RequireChdir(t, t.TempDir())
	defer revertWd()
	require.NoError(t, os.WriteFile(".envoy-version", []byte("b.b.b"), 0o600))

	c, _, _ := newApp(o)
	err := c.Run([]string{"func-e", "run"})

	// Verify the command failed with the expected error
	expectedErr := fmt.Sprintf(`invalid version in "$PWD/.envoy-version": "b.b.b" should look like %q or %q`, version.LastKnownEnvoy, version.LastKnownEnvoyMinor)
	require.EqualError(t, err, moreos.ReplacePathSeparator(expectedErr))
}

func TestFuncERun_ErrsWhenVersionsServerDown(t *testing.T) {
	tempDir := t.TempDir()

	o := &globals.GlobalOpts{
		EnvoyVersionsURL: "https://127.0.0.1:9999",
		HomeDir:          tempDir,
		Out:              new(bytes.Buffer),
	}
	c, _, _ := newApp(o)
	err := c.Run([]string{"func-e", "run"})

	require.Contains(t, o.Out.(*bytes.Buffer).String(), moreos.Sprintf("looking up the latest Envoy version"))
	require.Contains(t, err.Error(), fmt.Sprintf(`couldn't lookup the latest Envoy version from %s`, o.EnvoyVersionsURL))
}
