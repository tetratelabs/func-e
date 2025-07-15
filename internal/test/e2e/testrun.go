// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"bufio"
	"context"
	_ "embed"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal"
	"github.com/tetratelabs/func-e/internal/tar"
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
		ExpectFail: false,
		Args:       []string{"--config-yaml", "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}"},
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
		ExpectFail: false,
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

// TestRun_MinimalListener tests "func-e run" with a minimal listener configuration without an admin server.
// It verifies that the api archive is created correctly.
func TestRun_MinimalListener(ctx context.Context, t *testing.T, factory FuncEFactory) {
	revertWd := morerequire.RequireChdir(t, t.TempDir())
	defer revertWd()

	inlineString := []byte("Hello, World!")

	setupTestFiles(t, map[string][]byte{
		"envoy.yaml": internal.MinimalYaml,
	})

	var capturedRunDir string
	executeRunTest(ctx, t, factory, RunTestOptions{
		ExpectFail: false,
		TestFunc: func(ctx context.Context, runDir string, envoyPid int32, interruptFuncE func(context.Context) error, adminClient *AdminClient) {
			mainURL, err := adminClient.GetListenerBaseURL(ctx, "main")
			require.NoError(t, err)

			responseBody, err := httpGet(ctx, mainURL)
			require.NoError(t, err)
			require.Equal(t, inlineString, responseBody)

			capturedRunDir = runDir
		},
		Args: []string{"-c", "envoy.yaml"},
	})
	if capturedRunDir != "" {
		verifyRunArchive(t, capturedRunDir)
	}
}

// TestRun_InvalidConfig tests "func-e run" with an invalid configuration.
// It verifies that Envoy fails to start, the api archive is created, and an error is logged.
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
			testFailed <- t.Failed() // Send test failure status after processing.
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

	// There's a race between the run goroutine finishing and the context canceling (e.g. due to timeout).
	var runErr error
	select {
	case e := <-runErrCh:
		runErr = e // Receive error from Run goroutine.
	case <-ctx.Done():
		t.Fatalf("Context done before func-e %v", ctx.Err())
	}

	// The below order is critical to ensure the stderrScanner goroutine exits.
	// 1. Closing the stderrWriter will signal the stderrScanner goroutine to exit.
	_ = stderrWriter.Close()
	// 2. Wait for the stderrScanner goroutine to finish processing.
	failed := <-testFailed
	// 3. We can now safely check if there is any scanning error.
	require.NoError(t, stderrScanner.Err())

	if failed {
		t.FailNow()
	}

	if !opts.ExpectFail {
		require.True(t, envoyPid != 0, "expected Envoy to start")
		require.NoError(t, runErr)
	} else {
		require.True(t, envoyPid == 0, "expected Envoy not to start")
		require.Error(t, runErr)
		require.NotContains(t, stdoutBuf.String(), "invoking shutdown hooks with deadline")
	}

	// After shutdown, the Envoy process should not exist. Otherwise, there's a leak issue.
	if envoyPid != 0 {
		_, err = process.NewProcessWithContext(ctx, envoyPid)
		require.Error(t, err, "expected Envoy process to be gone after shutdown")
	}
}

// runTestAndInterruptFuncE runs the test function and ensures func-e is interrupted afterward.
func runTestAndInterruptFuncE(ctx context.Context, t *testing.T, funcE FuncE, runDir string, envoyPid int32, callback func(context.Context, string, int32, func(context.Context) error, *AdminClient)) {
	defer func() {
		require.NoError(t, funcE.Interrupt(ctx))
		if ctx.Err() != nil {
			t.Logf("context canceled during interrupt: %v", ctx.Err())
		}
	}() // Defer ensures interrupt is called even if callback panics.
	adminClient := requireAdminReady(ctx, t, runDir)
	callback(ctx, runDir, envoyPid, func(ctx context.Context) error { return funcE.Interrupt(ctx) }, adminClient)
}

// verifyRunArchive checks the contents of the api archive, ensuring expected files exist.
func verifyRunArchive(t *testing.T, runDir string) {
	runArchive := runDir + ".tar.gz"

	src, err := os.Open(runArchive)
	require.NoError(t, err)

	err = tar.Untar(runDir, src)
	require.NoError(t, err)

	expectedFiles := []string{"stdout.log", "stderr.log", "config_dump.json", "stats.json"}
	for _, filename := range expectedFiles {
		path := filepath.Join(runDir, filename)
		f, err := os.Stat(path)
		require.NoError(t, err)
		require.Positive(t, f.Size())
	}
}
