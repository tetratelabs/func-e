// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/globals"
)

// TestRuntime_Run_EnvoyError takes care to not duplicate test/e2e/testrun.go, but still give some coverage.
func TestRuntime_Run_EnvoyError(t *testing.T) {
	tempDir := t.TempDir()
	runDir := filepath.Join(tempDir, "runs", "1619574747231823000")
	require.NoError(t, os.MkdirAll(runDir, 0o750))

	// Initialize runtime
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	logToOutput := func(format string, args ...interface{}) {
		stdout.WriteString(fmt.Sprintf(format, args...) + "\n")
	}
	r := NewRuntime(&globals.RunOpts{
		EnvoyPath: fakeEnvoyBin,
		RunDir:    runDir,
	}, logToOutput)
	r.Out, r.Err = stdout, stderr

	// Envoy with invalid config is expected to fail
	err := r.Run(t.Context(), []string{"--config-yaml", "invalid.yaml"})
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode())

	t.Run("command arguments", func(t *testing.T) {
		require.Equal(t, []string{
			fakeEnvoyBin,
			"--config-yaml", "invalid.yaml",
			// test we added additional arguments
			"--admin-address-path", filepath.Join(runDir, "admin-address.txt"),
			"--",
			"--func-e-run-dir", runDir,
		}, r.cmd.Args, "command arguments mismatch")
		require.Empty(t, r.cmd.Dir, "working directory should be empty")
	})

	t.Run("output messages", func(t *testing.T) {
		require.Contains(t, stdout.String(), fmt.Sprintf("starting: %s", fakeEnvoyBin))
		require.Contains(t, stderr.String(), "cannot unmarshal !!str")
	})

	t.Run("run directory exists", func(t *testing.T) {
		info, err := os.Stat(runDir)
		require.NoError(t, err, "run directory should exist")
		require.True(t, info.IsDir(), "run directory should be a directory")
	})
}
