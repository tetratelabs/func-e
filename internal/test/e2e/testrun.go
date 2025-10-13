// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal"
	internalapi "github.com/tetratelabs/func-e/internal/api"
	"github.com/tetratelabs/func-e/internal/test/morerequire"
)

// RunTestFunc is a test function called once Envoy is started.
type RunTestFunc func(ctx context.Context, interruptFuncE func(context.Context) error, adminClient internalapi.AdminClient)

type RunTestOptions struct {
	ExpectFail bool
	TestFunc   RunTestFunc
	Args       []string
}

// Tests are implemented individually rather than as table-driven tests to facilitate
// easier debugging and selective test execution.

// TestRun tests the basic "func-e run" command with a minimal configuration.
func TestRun(ctx context.Context, t *testing.T, factory FuncEFactory) {
	executeRunTest(ctx, t, factory, RunTestOptions{
		Args: []string{"--config-yaml", "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}"},
	})
}

// TestRun_AdminAddressPath tests we do not assume the admin address path is in the run directory.
func TestRun_AdminAddressPath(ctx context.Context, t *testing.T, factory FuncEFactory) {
	adminAddressPath := path.Join(t.TempDir(), "--admin-address-path")
	executeRunTest(ctx, t, factory, RunTestOptions{
		Args: []string{"--config-yaml", "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}", "--admin-address-path", adminAddressPath},
	})
}

// TestRun_LogWarn tests we do not depend on any log messages written to info level.
func TestRun_LogWarn(ctx context.Context, t *testing.T, factory FuncEFactory) {
	executeRunTest(ctx, t, factory, RunTestOptions{
		Args: []string{"--config-yaml", "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}", "--log-level", "warn"},
	})
}

// TestRun_StaticFile tests "func-e run" runs envoy in the right directory, and can read files in it.
func TestRun_StaticFile(ctx context.Context, t *testing.T, factory FuncEFactory) {
	revertWd := morerequire.RequireChdir(t, t.TempDir())
	defer revertWd()

	expectedResponse := []byte("foo")

	setupTestFiles(t, map[string][]byte{
		"envoy.yaml":   internal.StaticFileYaml,
		"response.txt": expectedResponse,
	})

	executeRunTest(ctx, t, factory, RunTestOptions{
		TestFunc: func(ctx context.Context, interruptFuncE func(context.Context) error, adminClient internalapi.AdminClient) {
			req, err := adminClient.NewListenerRequest(ctx, "main", http.MethodGet, "/", nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, expectedResponse, body)
		},
		Args: []string{"-c", "envoy.yaml"},
	})
}

// TestRun_RunDir tests that runtime-generated files (logs, config_dump.json) are properly
// created in the state directory. This test uses api.StateHome() library option to override
// where state files are written, without affecting the data directory (envoy versions).
// The factory must be configured with the provided stateDir, and the test will verify files
// are created in the expected location (stateDir/envoy-runs/{runID}/).
func TestRun_RunDir(ctx context.Context, t *testing.T, factory FuncEFactory, stateDir string) {
	// Work in a different directory to ensure state files don't pollute working directory
	revertWd := morerequire.RequireChdir(t, t.TempDir())
	defer revertWd()

	inlineString := []byte("Hello, World!")

	setupTestFiles(t, map[string][]byte{
		"envoy.yaml": internal.AccessLogYaml,
	})

	executeRunTest(ctx, t, factory, RunTestOptions{
		TestFunc: func(ctx context.Context, interruptFuncE func(context.Context) error, adminClient internalapi.AdminClient) {
			// Get the listener twice to generate access logs in stdout
			req, err := adminClient.NewListenerRequest(ctx, "main", http.MethodGet, "/", nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck

			responseBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, inlineString, responseBody)

			// Make another request to generate more logs
			req, err = adminClient.NewListenerRequest(ctx, "main", http.MethodGet, "/", nil)
			require.NoError(t, err)

			resp, err = http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck

			// Wait a moment for logs to be flushed
			time.Sleep(100 * time.Millisecond)

			// Now check the state directory before shutting down.
			// Logs are in ${stateDir}/envoy-runs/{runID}/
			// The factory was configured with api.StateHome(stateDir) library option.
			envoyRunsDir := filepath.Join(stateDir, "envoy-runs")
			entries, err := os.ReadDir(envoyRunsDir)
			require.NoError(t, err, "envoy-runs directory should exist")
			require.Len(t, entries, 1, "should have exactly one run subdirectory")
			actualRunDir := filepath.Join(envoyRunsDir, entries[0].Name())
			checkRunDirectoryWithAccessLogs(t, actualRunDir)
		},
		Args: []string{"-c", "envoy.yaml"},
	})
}

// TestRun_InvalidConfig tests "func-e run" with an invalid configuration.
func TestRun_InvalidConfig(ctx context.Context, t *testing.T, factory FuncEFactory) {
	revertWd := morerequire.RequireChdir(t, t.TempDir())
	defer revertWd()

	setupTestFiles(t, map[string][]byte{
		"invalid.yaml": []byte("invalid yaml content"),
	})

	executeRunTest(ctx, t, factory, RunTestOptions{
		ExpectFail: true,
		Args:       []string{"-c", "invalid.yaml"},
	})
}

// TestRun_CtrlCs tests the "Ctrl+C twice" behavior where multiple interrupts
// are handled gracefully without causing issues.
func TestRun_CtrlCs(ctx context.Context, t *testing.T, factory FuncEFactory) {
	executeRunTest(ctx, t, factory, RunTestOptions{
		Args: []string{"--config-yaml", "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}"},
		TestFunc: func(ctx context.Context, interruptFuncE func(context.Context) error, adminClient internalapi.AdminClient) {
			// First interrupt should begin graceful shutdown
			require.NoError(t, interruptFuncE(ctx))

			// Send 5 more interrupts to ensure no special casing
			for i := 0; i < 5; i++ {
				require.NoError(t, interruptFuncE(ctx))
			}
		},
	})
}

// TestRun_LegacyHomeDir tests "func-e run" with FUNC_E_HOME (legacy mode) to ensure:
// 1. Files are created in legacy directory structure (versions/, runs/, version)
// 2. RunID is in epoch format (not ISO timestamp)
// Note: Deprecation warning is tested separately in CLI tests (internal/cmd/app_test.go)
func TestRun_LegacyHomeDir(ctx context.Context, t *testing.T, factory FuncEFactory) {
	homeDir := t.TempDir()
	t.Setenv("FUNC_E_HOME", homeDir)

	revertWd := morerequire.RequireChdir(t, t.TempDir())
	defer revertWd()

	setupTestFiles(t, map[string][]byte{
		"envoy.yaml": internal.AccessLogYaml,
	})

	executeRunTest(ctx, t, factory, RunTestOptions{
		TestFunc: func(ctx context.Context, interruptFuncE func(context.Context) error, adminClient internalapi.AdminClient) {
			// Get the listener twice to generate access logs in stdout
			req, err := adminClient.NewListenerRequest(ctx, "main", http.MethodGet, "/", nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck

			responseBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, []byte("Hello, World!"), responseBody)

			// Make another request to generate more logs
			req, err = adminClient.NewListenerRequest(ctx, "main", http.MethodGet, "/", nil)
			require.NoError(t, err)

			resp, err = http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck

			// Wait a moment for logs to be flushed
			time.Sleep(100 * time.Millisecond)

			// Verify legacy structure
			// Legacy mode: homeDir/runs/{epochRunID}/
			runsDir := filepath.Join(homeDir, "runs")
			entries, err := os.ReadDir(runsDir)
			require.NoError(t, err, "runs directory should exist in legacy mode")
			require.Len(t, entries, 1, "should have exactly one run subdirectory")

			runID := entries[0].Name()
			// Verify runID is epoch format (numeric string, not ISO timestamp)
			require.Regexp(t, `^\d{19}$`, runID)

			actualRunDir := filepath.Join(runsDir, runID)
			checkRunDirectoryWithAccessLogs(t, actualRunDir)
		},
		Args: []string{"-c", "envoy.yaml"},
	})
}

// setupTestFiles writes the given files to the current directory for test setup.
func setupTestFiles(t *testing.T, files map[string][]byte) {
	for name, content := range files {
		require.NoError(t, os.WriteFile(name, content, 0o600))
	}
}

// executeRunTest executes the given func-e arguments and runs the provided test function once Envoy is available.
func executeRunTest(ctx context.Context, t *testing.T, factory FuncEFactory, opts RunTestOptions) {
	var stdoutBuf, stderr strings.Builder

	t.Cleanup(func() {
		if t.Failed() {
			_, _ = os.Stdout.WriteString(stdoutBuf.String())
			_, _ = os.Stderr.WriteString(stderr.String())
		}
	})

	funcE, err := factory.New(ctx, t, &stdoutBuf, &stderr)
	require.NoError(t, err)

	// Note: We can't check the func-e process because in the case of api-mode it is the test runner!
	// So, we use the func-e goroutine exit as a proxy of the func-e process exit.
	runErrCh := make(chan error, 1)
	go func() {
		runErrCh <- funcE.Run(ctx, opts.Args)
	}()

	// Poll synchronously for Envoy to start and become ready via admin API
	var adminClient internalapi.AdminClient
	var runErr error

	// When ExpectFail is true, we don't poll for Envoy to start - just wait for it to fail
	if opts.ExpectFail {
		runErr = <-runErrCh
	} else {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		var lastErr error

	polling:
		for {
			select {
			case runErr = <-runErrCh:
				// funcE.Run completed
				break polling

			case <-ctx.Done():
				// Test context timed out
				if lastErr != nil {
					t.Errorf("timeout waiting for Envoy to start: %v", lastErr)
				} else {
					t.Errorf("timeout waiting for Envoy to start")
				}
				runErr = <-runErrCh
				break polling

			case <-ticker.C:
				// Try to get Envoy process info and wait for admin API to be ready
				adminClient, lastErr = funcE.OnStart(ctx)
				if lastErr != nil {
					continue // Keep polling
				}

				// Envoy is ready! Run test or just interrupt
				if opts.TestFunc != nil {
					runTestAndInterruptFuncE(ctx, t, funcE, adminClient, opts.TestFunc)
				} else {
					require.NoError(t, funcE.Interrupt(ctx))
				}

				// Wait for funcE.Run to complete after test
				runErr = <-runErrCh
				break polling
			}
		}
	}

	// Update envoyPid to reflect current state - if process doesn't exist or isn't running, it's 0
	var envoyPid int32
	if funcE != nil {
		envoyPid = int32(funcE.EnvoyPid())
		if envoyPid != 0 {
			envoyProcess, err := process.NewProcessWithContext(ctx, envoyPid)
			if err != nil {
				envoyPid = 0 // Process doesn't exist
			} else {
				isRunning, _ := envoyProcess.IsRunning()
				if !isRunning {
					envoyPid = 0 // Process exists but not running
				}
			}
		}
	}

	// Normal shutdown (including interrupts) should return nil error
	// This matches Envoy's behavior of returning exit code 0 on graceful shutdown
	if !opts.ExpectFail {
		require.NoError(t, runErr, "expected func-e to exit cleanly on interrupt")
		require.Equal(t, int32(0), envoyPid, "expected Envoy process to be gone after normal shutdown")
	} else {
		require.True(t, envoyPid == 0, "expected Envoy not to start")
		require.Error(t, runErr)
	}

	// Clean up any leaked process
	if envoyPid != 0 {
		t.Logf("Cleaning up Envoy process %d", envoyPid)
		if p, err := process.NewProcessWithContext(ctx, envoyPid); err == nil {
			_ = p.Kill()
		}
	}
}

// runTestAndInterruptFuncE runs the test function and ensures func-e is interrupted afterward.
func runTestAndInterruptFuncE(ctx context.Context, t *testing.T, funcE FuncE, adminClient internalapi.AdminClient, callback RunTestFunc) {
	defer func() {
		if err := funcE.Interrupt(ctx); err != nil {
			// Only fail if it's not an expected error from killing the process
			if !errors.Is(err, os.ErrProcessDone) && !errors.Is(err, syscall.ESRCH) {
				require.NoError(t, err)
			}
		}
		if ctx.Err() != nil {
			t.Logf("context canceled during interrupt: %v", ctx.Err())
		}
	}() // Defer ensures interrupt is called even if callback panics.

	callback(ctx, func(ctx context.Context) error { return funcE.Interrupt(ctx) }, adminClient)
}

func checkRunDirectoryWithAccessLogs(t *testing.T, runDir string) {
	// Check log files (always created)
	logFiles := []string{"stdout.log", "stderr.log"}
	for _, filename := range logFiles {
		path := filepath.Join(runDir, filename)
		f, err := os.Stat(path)
		require.NoError(t, err, "%s should exist", filename)
		if filename == "stderr.log" {
			require.Positive(t, f.Size(), "%s is empty", filename)
		}
	}

	// Check config_dump.json (created when admin address is available)
	configDumpPath := filepath.Join(runDir, "config_dump.json")
	f, err := os.Stat(configDumpPath)
	require.NoError(t, err, "config_dump.json should exist")
	require.Positive(t, f.Size(), "config_dump.json should not be empty")

	// Check stdout.log contains access logs
	stdoutPath := filepath.Join(runDir, "stdout.log")
	stdoutContent, err := os.ReadFile(stdoutPath)
	require.NoError(t, err, "should be able to read stdout.log")

	// Verify access log entries are present
	stdoutStr := string(stdoutContent)
	require.NotEmpty(t, stdoutStr, "stdout.log should contain access logs")

	// Check for expected access log format elements
	require.Contains(t, stdoutStr, "GET / HTTP/1.1")
	require.Contains(t, stdoutStr, "200")
	require.Contains(t, stdoutStr, "13") // 13 bytes for 'Hello, World!'
}
