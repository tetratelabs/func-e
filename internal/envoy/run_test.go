// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

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

func TestRuntime_Run_StartupHook(t *testing.T) {
	var logBuf bytes.Buffer
	logToOutput := func(format string, args ...interface{}) {
		logBuf.WriteString(fmt.Sprintf(format, args...) + "\n")
	}

	tests := []struct {
		name        string
		startupHook StartupHook
		expectError string
		expectLog   string
		envoyArgs   []string
	}{
		{
			name: "startup hook returns error",
			startupHook: func(ctx context.Context, runDir, adminAddress string) error {
				return errors.New("database connection failed")
			},
			expectError: "database connection failed",
			expectLog:   "database connection failed",
			envoyArgs: []string{
				"--config-yaml", "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}",
			},
		},
		{
			name: "startup hook panics",
			startupHook: func(ctx context.Context, runDir, adminAddress string) error {
				panic("nil pointer dereference")
			},
			expectError: "processStderr panicked: nil pointer dereference",
			expectLog:   "processStderr panicked: nil pointer dereference",
			envoyArgs: []string{
				"--config-yaml", "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}",
			},
		},
		{
			name: "startup hook succeeds",
			startupHook: func(ctx context.Context, runDir, adminAddress string) error {
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
				EnvoyPath: fakeEnvoyBin,
				RunDir:    runDir,
			}, logToOutput)
			r.Out, r.Err = new(bytes.Buffer), new(bytes.Buffer)

			// Wrap the startup hook to cancel context after execution
			// This ensures the test doesn't hang waiting for envoy to exit
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			// Wrap the hook to cancel context after execution
			r.startupHook = func(ctx context.Context, runDir, adminAddress string) error {
				defer cancel() // Always cancel, even if hook panics
				return tt.startupHook(ctx, runDir, adminAddress)
			}

			err := r.Run(ctx, tt.envoyArgs)

			// Check error
			if tt.expectError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectError)
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

				// Check if process is still alive by sending signal 0
				err := r.cmd.Process.Signal(syscall.Signal(0))
				require.Error(t, err, "process should be dead")
				require.Contains(t, err.Error(), "process already finished")
			}
		})
	}
}

func TestRuntime_processStderr(t *testing.T) {
	tests := []struct {
		name        string
		stderrLines []string
		startupHook func(context.CancelFunc) StartupHook
		expectError string
	}{
		{
			name: "normal operation",
			stderrLines: []string{
				"[info] initializing epoch 0",
				"[info] admin address: 127.0.0.1:9901",
				"[info] starting main dispatch loop",
				"[info] all clusters initialized",
			},
			startupHook: func(cancelFunc context.CancelFunc) StartupHook {
				return func(ctx context.Context, runDir, adminAddress string) error {
					return nil
				}
			},
		},
		{
			name: "startup hook error stops processing",
			stderrLines: []string{
				"[info] starting main dispatch loop",
			},
			startupHook: func(cancelFunc context.CancelFunc) StartupHook {
				return func(ctx context.Context, runDir, adminAddress string) error {
					return errors.New("hook failed")
				}
			},
			expectError: "hook failed",
		},
		{
			name: "panic in startup hook",
			stderrLines: []string{
				"[info] starting main dispatch loop",
			},
			startupHook: func(cancelFunc context.CancelFunc) StartupHook {
				return func(ctx context.Context, runDir, adminAddress string) error {
					panic("test panic")
				}
			},
			expectError: "processStderr panicked: test panic",
		},
		{
			name: "context cancelled during processing",
			stderrLines: []string{
				"[info] some log",
				"[info] another log",
			},
			startupHook: func(cancelFunc context.CancelFunc) StartupHook {
				return func(ctx context.Context, runDir, adminAddress string) error {
					cancelFunc()
					return nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test environment
			tempDir := t.TempDir()
			adminAddressPath := filepath.Join(tempDir, "admin-address.txt")
			require.NoError(t, os.WriteFile(adminAddressPath, []byte("127.0.0.1:9901"), 0o600))

			// Setup runtime
			var logBuf bytes.Buffer
			logf := func(format string, args ...interface{}) {
				logBuf.WriteString(fmt.Sprintf(format, args...) + "\n")
			}

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			r := &Runtime{
				o: &globals.RunOpts{
					RunDir: tempDir,
				},
				logf:             logf,
				Err:              new(bytes.Buffer),
				adminAddressPath: adminAddressPath,
				startupHook:      tt.startupHook(cancel),
			}

			// Create a buffer with test data
			var stderrData bytes.Buffer
			for _, line := range tt.stderrLines {
				stderrData.WriteString(line + "\n")
			}

			errCh := make(chan error, 1)

			// Process stderr directly - no goroutine needed
			r.processStderr(ctx, &stderrData, errCh)

			// Get the error
			err := <-errCh

			// Check error
			if tt.expectError != "" {
				require.Error(t, err)
				require.EqualError(t, err, tt.expectError)
			} else {
				require.NoError(t, err)
			}

			// Verify stderr was copied
			stderrOutput := r.Err.(*bytes.Buffer).String()
			for _, line := range tt.stderrLines {
				require.Contains(t, stderrOutput, line)
			}
		})
	}
}
