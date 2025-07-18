// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal"
	"github.com/tetratelabs/func-e/internal/test/morerequire"
)

var envoyStartedLine = "starting main dispatch loop"

// RunTestFunc is a test function called once Envoy is started.
type RunTestFunc func(ctx context.Context, runDir string, envoyPid int32, interruptFuncE func(context.Context) error, adminClient *AdminClient)

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
		TestFunc: func(ctx context.Context, runDir string, envoyPid int32, interruptFuncE func(context.Context) error, adminClient *AdminClient) {
			mainURL, err := adminClient.GetListenerBaseURL(ctx, "main")
			require.NoError(t, err)

			body, err := httpGet(ctx, mainURL)
			require.NoError(t, err)
			require.Equal(t, expectedResponse, body)
		},
		Args: []string{"-c", "envoy.yaml"},
	})
}

// TestRun_RunDirectory tests that the run directory is properly created with expected files.
func TestRun_RunDirectory(ctx context.Context, t *testing.T, factory FuncEFactory) {
	revertWd := morerequire.RequireChdir(t, t.TempDir())
	defer revertWd()

	inlineString := []byte("Hello, World!")

	setupTestFiles(t, map[string][]byte{
		"envoy.yaml": internal.AccessLogYaml,
	})

	executeRunTest(ctx, t, factory, RunTestOptions{
		TestFunc: func(ctx context.Context, runDir string, envoyPid int32, interruptFuncE func(context.Context) error, adminClient *AdminClient) {
			mainURL, err := adminClient.GetListenerBaseURL(ctx, "main")
			require.NoError(t, err)

			// Get the listener twice to generate access logs in stdout
			responseBody, err := httpGet(ctx, mainURL)
			require.NoError(t, err)
			require.Equal(t, inlineString, responseBody)

			// Make another request to generate more logs
			_, err = httpGet(ctx, mainURL)
			require.NoError(t, err)

			// Wait a moment for logs to be flushed
			time.Sleep(100 * time.Millisecond)

			// Now check the run directory before shutting down
			checkRunDirectoryWithAccessLogs(t, runDir)
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
		TestFunc: func(ctx context.Context, runDir string, envoyPid int32, interruptFuncE func(context.Context) error, adminClient *AdminClient) {
			// First interrupt should begin graceful shutdown
			require.NoError(t, interruptFuncE(ctx))

			// Send 5 more interrupts to ensure no special casing
			for i := 0; i < 5; i++ {
				require.NoError(t, interruptFuncE(ctx))
			}
		},
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
	var stdoutBuf strings.Builder
	stderrReader, stderrWriter := io.Pipe()
	var stderrBuf strings.Builder

	t.Cleanup(func() {
		if t.Failed() {
			_, _ = os.Stdout.WriteString(stdoutBuf.String())
			_, _ = os.Stderr.WriteString(stderrBuf.String())
		}
	})

	funcE, err := factory.New(ctx, t, &stdoutBuf, stderrWriter)
	require.NoError(t, err)

	// Note: We can't check the func-e process because in the case of api-mode it is the test runner!
	// So, we use the func-e goroutine exit as a proxy of the func-e process exit.
	runErrCh := make(chan error, 1)
	go func() {
		runErrCh <- funcE.Run(ctx, opts.Args)
	}()

	// Test failures won't fail fast unless the calling goroutine checks the status.
	testFailed := make(chan bool, 1)

	// We don't need to process stdout because Envoy writes nothing notable to it.
	// We scan until the process is started, which is the barrier for running any test function.
	var envoyPid int32
	var runDir string
	stderrScanner := bufio.NewScanner(stderrReader)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic in stderrScanner goroutine: %v", r)
				testFailed <- true // Indicate that the test failed due to a panic
			} else {
				testFailed <- t.Failed() // Send test failure status after processing.
			}
		}()

		// Goroutine lifecycle: Scans stderr lines and handles Envoy start detection.
		// - Scans each line, checking for Envoy start indicator.
		// - Upon detecting start, gets PID, runs test if provided, interrupts Envoy.
		// - Exits the loop early after interrupting to allow process termination.
		// - Defers sending testFailed to ensure cleanup even on early return or panic.
		// Continues scanning until EOF or error.
		for stderrScanner.Scan() {
			line := stderrScanner.Text()
			stderrBuf.WriteString(line + "\n")
			if !strings.Contains(line, envoyStartedLine) {
				continue
			}

			// We shouldn't get here because on failure, the above loop should exit when envoy fails to start.
			if opts.ExpectFail {
				t.Errorf("Envoy started unexpectedly")
				break // break the loop to allow troubleshooting.
			}

			// Envoy started successfully when we want it to. Even if there is no TestFunc, validate there is a
			// run directory and Envoy PID.
			runDir, envoyPid, err = funcE.OnStart(ctx)
			require.NoError(t, err)

			// If a test function is provided, run it and then interrupt func-e. These are different branches
			// to ensure a test failure doesn't leak the interrupt call.
			if opts.TestFunc != nil {
				runTestAndInterruptFuncE(ctx, t, funcE, runDir, envoyPid, opts.TestFunc)
			} else {
				require.NoError(t, funcE.Interrupt(ctx))
			}
			break // break the loop to allow the process to terminate.
		}
	}()

	// This block waits for the result of the func-e run. There are three main
	// outcomes to wait for, handled by the select statement:
	// 1. The func-e run completes, sending an error (or nil) to `runErrCh`.
	// 2. The test fails within the stderr scanning goroutine, which sends a
	//    signal to `testFailed`.
	// 3. The test context is canceled (e.g., due to a timeout).
	var runErr error
	select {
	case <-ctx.Done():
		// The test timed out or was canceled before func-e finished.
		t.Fatalf("Context done before func-e finished: %v", ctx.Err())
	case runErr = <-runErrCh:
		// `funcE.Run` finished. `runErr` now holds its return value.
	case failed := <-testFailed:
		// The stderr scanner goroutine reported a test failure.
		if failed {
			t.FailNow() // Fail the test immediately.
		}
		// If the test didn't fail, we still need to wait for funcE.Run to finish.
		runErr = <-runErrCh
	}

	// Close the stderr pipe writer to signal the scanner to stop.
	_ = stderrWriter.Close()
	// Ensure the scanner didn't encounter any errors.
	require.NoError(t, stderrScanner.Err())

	// Update envoyPid to reflect current state - if process doesn't exist or isn't running, it's 0
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
func runTestAndInterruptFuncE(ctx context.Context, t *testing.T, funcE FuncE, runDir string, envoyPid int32, callback func(context.Context, string, int32, func(context.Context) error, *AdminClient)) {
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
	adminClient := requireAdminReady(ctx, t, runDir)
	callback(ctx, runDir, envoyPid, func(ctx context.Context) error { return funcE.Interrupt(ctx) }, adminClient)
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
	require.Contains(t, stdoutStr, "GET / HTTP/1.1", "should contain GET request")
	require.Contains(t, stdoutStr, "200", "should contain 200 status code")
	require.Contains(t, stdoutStr, "13", "should contain response body size (13 bytes for 'Hello, World!')")
}
