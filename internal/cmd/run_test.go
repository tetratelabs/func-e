// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/require"

	rootcmd "github.com/tetratelabs/func-e/internal/cmd"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/version"
)

// TestFuncERun takes care to not duplicate test/e2e/testrun.go, but still give some coverage.
func TestFuncERun(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		o := setupTest(t)

		// Use a fake Envoy binary
		o.EnvoyPath = fakeEnvoyBin

		// Create a pipe to capture stderr
		stderrReader, stderrWriter := io.Pipe()
		stdout := new(bytes.Buffer)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		// Start a goroutine to scan stderr until it reaches "starting main dispatch loop" written by envoy
		go func() {
			scanner := bufio.NewScanner(stderrReader)
			for scanner.Scan() {
				if strings.Contains(scanner.Text(), "starting main dispatch loop") {
					cancel() // interrupts the child func-e process
					return
				}
			}
		}()

		o.Out = stdout
		o.EnvoyOut = stdout
		o.EnvoyErr = stderrWriter

		err := rootcmd.DoMain(ctx, stdout, stderrWriter, []string{"run", "--config-yaml", "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}"}, o, "test")
		require.NoError(t, err)

		// TestFuncERun_TeesConsoleToLogs proves we can read Envoy logs
		stderrBytes, err := os.ReadFile(filepath.Join(o.RunDir, "stderr.log"))
		stderr := string(stderrBytes)
		require.NoError(t, err)
		pattern := `(?s).*initializing epoch 0.*admin address:.*starting main dispatch loop.*`
		matched, err := regexp.MatchString(pattern, stderr)
		require.NoError(t, err)
		require.True(t, matched, "Didn't find %s in Envoy stderr: %s", pattern, stderr)
	})
}

func TestFuncERun_TeesConsoleToLogs(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		o := setupTest(t)
		o.Out = io.Discard

		stdout := new(bytes.Buffer)
		stderr := new(bytes.Buffer)
		o.EnvoyOut = stdout
		o.EnvoyErr = stderr

		runWithInvalidConfig(t, o, stdout, stderr)

		stdoutStream := stdout.String()
		require.Empty(t, stdoutStream, "envoy doesn't write to stdout by default")
		stdoutLogBytes, err := os.ReadFile(filepath.Join(o.RunDir, "stdout.log"))
		require.NoError(t, err)
		require.Empty(t, stdoutLogBytes)

		stderrStream := stderr.String()
		require.Contains(t, stderrStream, "At least one of --config-path or --config-yaml or Options::configProto() should be non-empty")
		stderrLogBytes, err := os.ReadFile(filepath.Join(o.RunDir, "stderr.log"))
		require.NoError(t, err)
		require.Equal(t, stderrStream, string(stderrLogBytes))
	})
}

func TestFuncERun_PassesFlagsToEnvoy(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		o := setupTest(t)
		o.EnvoyPath = fakeEnvoyBin

		stdout := new(bytes.Buffer)
		stderr := new(bytes.Buffer)
		err := rootcmd.DoMain(t.Context(), stdout, stderr, []string{"run", "--help"}, o, "test")

		var exitErr *exec.ExitError
		require.ErrorAs(t, err, &exitErr)
		require.Equal(t, 1, exitErr.ExitCode())
		require.Empty(t, stdout.String())
		require.Contains(t, stderr.String(), "At least one of --config-path or --config-yaml or Options::configProto() should be non-empty")
	})
}

func TestFuncERun_ReadsHomeVersionFile(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		o := setupTest(t)
		o.EnvoyVersion = "" // pretend this is an initial setup
		o.Out = new(bytes.Buffer)

		require.NoError(t, os.WriteFile(filepath.Join(o.ConfigHome, "envoy-version"), []byte(version.LastKnownEnvoyMinor), 0o600))

		stdout := new(bytes.Buffer)
		stderr := new(bytes.Buffer)
		runWithInvalidConfig(t, o, stdout, stderr)

		// No implicit lookup
		require.NotContains(t, o.Out.(*bytes.Buffer).String(), "looking up latest version")
		require.Equal(t, version.LastKnownEnvoy, o.EnvoyVersion)

		writtenVersion, err := os.ReadFile(filepath.Join(o.ConfigHome, "envoy-version"))
		require.NoError(t, err)
		require.Equal(t, version.LastKnownEnvoyMinor.String(), string(writtenVersion))
	})
}

func TestFuncERun_CreatesHomeVersionFile(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		o := setupTest(t)
		o.EnvoyVersion = "" // pretend this is an initial setup
		o.Out = new(bytes.Buffer)

		// make sure first run where the home doesn't exist yet, works!
		require.NoError(t, os.RemoveAll(o.DataHome))

		stdout := new(bytes.Buffer)
		stderr := new(bytes.Buffer)
		runWithInvalidConfig(t, o, stdout, stderr)

		// We logged the implicit lookup
		require.Contains(t, o.Out.(*bytes.Buffer).String(), "looking up the latest Envoy version")
		require.FileExists(t, filepath.Join(o.ConfigHome, "envoy-version"))
		require.Equal(t, version.LastKnownEnvoy, o.EnvoyVersion)

		writtenVersion, err := os.ReadFile(filepath.Join(o.ConfigHome, "envoy-version"))
		require.NoError(t, err)
		require.Equal(t, version.LastKnownEnvoyMinor.String(), string(writtenVersion))
	})
}

// runWithInvalidConfig intentionally has envoy quit.
func runWithInvalidConfig(t *testing.T, o *globals.GlobalOpts, stdout, stderr *bytes.Buffer) {
	t.Helper()
	err := rootcmd.DoMain(t.Context(), stdout, stderr, []string{"run"}, o, "test")
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode())
}

func TestFuncERun_ValidatesHomeVersion(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		o := setupTest(t)
		o.Out = new(bytes.Buffer)

		o.EnvoyVersion = ""
		require.NoError(t, os.WriteFile(filepath.Join(o.ConfigHome, "envoy-version"), []byte("a.a.a"), 0o600))

		stdout := new(bytes.Buffer)
		stderr := new(bytes.Buffer)
		err := rootcmd.DoMain(t.Context(), stdout, stderr, []string{"run"}, o, "test")

		expectedErr := fmt.Sprintf(`invalid version in "$FUNC_E_CONFIG_HOME/envoy-version": "a.a.a" should look like %q or %q`, version.LastKnownEnvoy, version.LastKnownEnvoyMinor)
		require.EqualError(t, err, expectedErr)
	})
}

// TestFuncERun_ValidatesWorkingVersion duplicates logic in version_test.go to ensure a non-home version validates.
func TestFuncERun_ValidatesWorkingVersion(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		o := setupTest(t)
		o.Out = new(bytes.Buffer)
		o.EnvoyVersion = ""

		t.Chdir(t.TempDir())
		require.NoError(t, os.WriteFile(".envoy-version", []byte("b.b.b"), 0o600))

		stdout := new(bytes.Buffer)
		stderr := new(bytes.Buffer)
		err := rootcmd.DoMain(t.Context(), stdout, stderr, []string{"run"}, o, "test")

		expectedErr := fmt.Sprintf(`invalid version in "$PWD/.envoy-version": "b.b.b" should look like %q or %q`, version.LastKnownEnvoy, version.LastKnownEnvoyMinor)
		require.EqualError(t, err, expectedErr)
	})
}

func TestFuncERun_ErrsWhenVersionsServerDown(t *testing.T) {
	tempDir := t.TempDir()

	o := &globals.GlobalOpts{
		EnvoyVersionsURL: "https://127.0.0.1:9999",
		ConfigHome:       tempDir,
		DataHome:         tempDir,
		StateHome:        tempDir,
		RuntimeDir:       tempDir,
		Out:              new(bytes.Buffer),
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	err := rootcmd.DoMain(t.Context(), stdout, stderr, []string{"run"}, o, "test")

	require.Contains(t, o.Out.(*bytes.Buffer).String(), "looking up the latest Envoy version")
	require.ErrorContains(t, err, `couldn't lookup the latest Envoy version from `+o.EnvoyVersionsURL)
}
