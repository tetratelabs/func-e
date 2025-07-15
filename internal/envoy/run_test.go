// Copyright 2019 Tetrate
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

package envoy

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/moreos"
)

// TestRuntime_Run_EnvoyError takes care to not duplicate test/e2e/testrun.go, but still give some coverage.
func TestRuntime_Run_EnvoyError(t *testing.T) {
	tempDir := t.TempDir()
	runDir := filepath.Join(tempDir, "runs", "1619574747231823000")
	require.NoError(t, os.MkdirAll(runDir, 0o750))

	// Initialize runtime
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	logToOutput := func(format string, args ...interface{}) {
		stdout.WriteString(moreos.Sprintf(format, args...) + "\n")
	}
	r := NewRuntime(&globals.RunOpts{
		EnvoyPath: fakeEnvoyBin,
		RunDir:    runDir,
	}, logToOutput)
	r.Out, r.Err = stdout, stderr

	// Envoy with invalid config is expected to fail
	err := r.Run(context.Background(), []string{"--config-yaml", "invalid.yaml"})
	require.EqualError(t, err, "envoy exited with status: 1")

	t.Run("shutdown hooks not invoked", func(t *testing.T) {
		// Check that the shutdown hooks log message is NOT present
		require.NotContains(t, stdout.String(), "invoking shutdown hooks with deadline")
	})

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
		require.Contains(t, stdout.String(), moreos.Sprintf("starting: %s", fakeEnvoyBin))
		require.Contains(t, stderr.String(), "cannot unmarshal !!str")
	})

	t.Run("archive created", func(t *testing.T) {
		files, err := os.ReadDir(filepath.Dir(runDir))
		require.NoError(t, err, "failed to read runs directory")
		require.Len(t, files, 1, "expected one archive file")
		require.Equal(t, runDir+".tar.gz", filepath.Join(filepath.Dir(runDir), files[0].Name()))
	})
}
