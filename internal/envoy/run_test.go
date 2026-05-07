// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/admin"
	internalapi "github.com/tetratelabs/func-e/internal/api"
	"github.com/tetratelabs/func-e/internal/globals"
)

// TestRuntime_Run_EnvoyError takes care to not duplicate test/e2e/testrun.go, but still give some coverage.
func TestRuntime_Run_EnvoyError(t *testing.T) {
	tempDir := t.TempDir()
	runDir := filepath.Join(tempDir, "runs", "1619574747231823000")
	require.NoError(t, os.MkdirAll(runDir, 0o750))

	// Initialize runtime
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	logToOutput := func(format string, args ...any) {
		stdout.WriteString(fmt.Sprintf(format, args...) + "\n")
	}
	r := NewRuntime(&globals.RunOpts{
		EnvoyPath:  fakeEnvoyBin,
		HTTPClient: http.DefaultClient,
		RunDir:     runDir,
		TempDir:    runDir,
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
			admin.AddressPathFlag, filepath.Join(runDir, "admin-address.txt"),
		}, r.cmd.Args, "command arguments mismatch")
		require.Empty(t, r.cmd.Dir, "working directory should be empty")
	})

	t.Run("output messages", func(t *testing.T) {
		require.Contains(t, stdout.String(), "starting: "+fakeEnvoyBin)
		require.Contains(t, stderr.String(), "cannot unmarshal !!str")
	})

	t.Run("run directory exists", func(t *testing.T) {
		info, err := os.Stat(runDir)
		require.NoError(t, err, "run directory should exist")
		require.True(t, info.IsDir(), "run directory should be a directory")
	})
}

func TestRuntime_Run_StartupHook(t *testing.T) {
	var logBuf bytes.Buffer
	logToOutput := func(format string, args ...any) {
		logBuf.WriteString(fmt.Sprintf(format, args...) + "\n")
	}

	tests := []struct {
		name        string
		startupHook internalapi.StartupHook
		expectedErr string
		expectLog   string
		envoyArgs   []string
	}{
		{
			name: "startup hook returns error",
			startupHook: func(_ context.Context, _ internalapi.AdminClient, _ string) error {
				return errors.New("database connection failed")
			},
			expectedErr: "database connection failed",
			expectLog:   "database connection failed",
			envoyArgs: []string{
				"--config-yaml", "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}",
			},
		},
		{
			name: "startup hook panics",
			startupHook: func(_ context.Context, _ internalapi.AdminClient, _ string) error {
				panic("nil pointer dereference")
			},
			expectedErr: "startup hook panicked: nil pointer dereference",
			expectLog:   "startup hook panicked: nil pointer dereference",
			envoyArgs: []string{
				"--config-yaml", "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}",
			},
		},
		{
			name: "startup hook succeeds",
			startupHook: func(_ context.Context, _ internalapi.AdminClient, _ string) error {
				logToOutput("startup hook executed successfully")
				return nil
			},
			expectLog: "startup hook executed successfully",
			envoyArgs: []string{
				"--config-yaml", "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the log buffer for each test
			logBuf.Reset()

			tempDir := t.TempDir()
			runDir := filepath.Join(tempDir, "runs", "test")
			require.NoError(t, os.MkdirAll(runDir, 0o750))

			// Create runtime with custom startup hook
			r := NewRuntime(&globals.RunOpts{
				EnvoyPath:  fakeEnvoyBin,
				HTTPClient: http.DefaultClient,
				RunDir:     runDir,
				TempDir:    runDir,
			}, logToOutput)
			r.Out, r.Err = new(bytes.Buffer), new(bytes.Buffer)

			// Wrap the startup hook to cancel context after execution
			// This ensures the test doesn't hang waiting for envoy to exit
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			// Wrap the hook to cancel context after execution
			r.startupHook = func(ctx context.Context, adminClient internalapi.AdminClient, runID string) error {
				defer cancel() // Always cancel, even if hook panics
				return tt.startupHook(ctx, adminClient, runID)
			}

			err := r.Run(ctx, tt.envoyArgs)

			// Check error
			if tt.expectedErr != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}

			// Check log output
			if tt.expectLog != "" {
				require.Contains(t, logBuf.String(), tt.expectLog)
			}

			// Verify the process is dead
			if r.cmd != nil && r.cmd.Process != nil {
				// Give the process a moment to fully exit
				time.Sleep(50 * time.Millisecond)

				p, err := process.NewProcessWithContext(t.Context(), int32(r.cmd.Process.Pid))
				if err != nil {
					require.ErrorIs(t, err, process.ErrorProcessNotRunning, "process should be dead")
					return
				}

				running, err := p.IsRunningWithContext(t.Context())
				require.NoError(t, err)
				require.False(t, running, "process should be dead")
			}
		})
	}
}
