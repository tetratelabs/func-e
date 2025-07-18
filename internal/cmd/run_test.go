// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	rootcmd "github.com/tetratelabs/func-e/internal/cmd"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/test/morerequire"
	"github.com/tetratelabs/func-e/internal/version"
)

func init() {
	// Don't let urfave quit the current test process on cancel!
	cli.OsExiter = func(code int) { log.Printf("urfave called exit: %d", code) }
}

// TestFuncERun takes care to not duplicate test/e2e/testrun.go, but still give some coverage.
func TestFuncERun(t *testing.T) {
	o := setupTest(t)

	// Use a fake Envoy binary
	o.EnvoyPath = fakeEnvoyBin

	c := rootcmd.NewApp(o)
	c.Name = "func-e"
	// Create a pipe to capture stderr
	stderrReader, stderrWriter := io.Pipe()
	c.ErrWriter = stderrWriter

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

	// When interrupted, func-e should return nil to match Envoy's behavior of exit code 0
	args := []string{"func-e", "run", "--config-yaml", "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}"}
	require.NoError(t, c.RunContext(ctx, args))

	// TestFuncERun_TeesConsoleToLogs proves we can read Envoy logs
	stderrBytes, err := os.ReadFile(filepath.Join(o.RunDir, "stderr.log"))
	stderr := string(stderrBytes)
	require.NoError(t, err)
	pattern := `(?s).*initializing epoch 0.*admin address:.*starting main dispatch loop.*`
	matched, err := regexp.MatchString(pattern, stderr)
	require.NoError(t, err)
	require.True(t, matched, "Didn't find %s in Envoy stderr: %s", pattern, stderr)
}

func TestFuncERun_TeesConsoleToLogs(t *testing.T) {
	o := setupTest(t)

	c, stdout, stderr := newApp(o)
	// ignore messages from func-e we only care about envoy
	o.Out = io.Discard
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
	require.NotContains(t, o.Out.(*bytes.Buffer).String(), "looking up latest version")
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
	require.Contains(t, o.Out.(*bytes.Buffer).String(), "looking up the latest Envoy version")
	require.FileExists(t, filepath.Join(o.HomeDir, "version"))
	require.Equal(t, version.LastKnownEnvoy, o.EnvoyVersion)

	writtenVersion, err := os.ReadFile(filepath.Join(o.HomeDir, "version"))
	require.NoError(t, err)
	require.Equal(t, version.LastKnownEnvoyMinor.String(), string(writtenVersion))
}

// runWithInvalidConfig intentionally has envoy quit. This allows tests to not have to interrupt envoy to proceed
func runWithInvalidConfig(t *testing.T, c *cli.App) {
	err := c.Run([]string{"func-e", "run"})
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode())
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
	require.EqualError(t, err, expectedErr)
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
	require.EqualError(t, err, expectedErr)
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

	require.Contains(t, o.Out.(*bytes.Buffer).String(), "looking up the latest Envoy version")
	require.Contains(t, err.Error(), fmt.Sprintf(`couldn't lookup the latest Envoy version from %s`, o.EnvoyVersionsURL))
}
