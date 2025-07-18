// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package shutdown

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/envoy"
	"github.com/tetratelabs/func-e/internal/globals"
)

func TestEnvoyAdminDataCollection(t *testing.T) {
	runDir := t.TempDir()

	stderr, err := runWithShutdownHook(t, runDir, enableAdminDataCollection)
	require.NoError(t, err)

	for _, filename := range adminAPIPaths {
		path := filepath.Join(runDir, filename)
		f, err := os.Stat(path)
		require.NoError(t, err, "error stating %v: %s", path, stderr)
		require.NotEmpty(t, f.Size(), "file %v was empty: %s", path, stderr)
	}
}

// TODO: consider re-using logic from test/e2e/testrun.go
// runWithShutdownHook is like RunEnvoy, except invokes the hook on shutdown
func runWithShutdownHook(t *testing.T, runDir string, hook EnableHook) (stderr *bytes.Buffer, err error) {
	o := &globals.RunOpts{EnvoyPath: fakeEnvoyBin, RunDir: runDir, DontArchiveRunDir: true}

	// Use a temporary file for stderr
	stderrFile, err := os.CreateTemp(runDir, "stderr.log")
	if err != nil {
		t.Fatalf("failed to create temp stderr file: %v", err)
	}
	defer stderrFile.Close() //nolint:errcheck

	r := envoy.NewRuntime(o, t.Logf)
	r.Out = io.Discard
	r.Err = stderrFile
	require.NoError(t, hook(r))

	// Prepare args
	args := []string{"--config-yaml", "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}"}
	adminAddressPath := filepath.Join(runDir, "admin-address.txt")
	args = append(args, "--admin-address-path", adminAddressPath)

	// Use a cancellable context with timeout as backup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start the runtime in a goroutine
	done := make(chan error, 1)
	go func() {
		done <- r.Run(ctx, args)
	}()

	// Wait for readiness by monitoring stderr
	readiness := make(chan bool, 1)
	go func() {
		defer close(readiness)
		for {
			if _, err := stderrFile.Seek(0, io.SeekStart); err != nil {
				return
			}
			stderrBytes, err := io.ReadAll(stderrFile)
			if err != nil {
				return
			}
			if strings.Contains(string(stderrBytes), "starting main dispatch loop") {
				readiness <- true
				return
			}
		}
	}()

	// Wait for readiness or timeout
	select {
	case <-readiness:
		// Envoy is ready, cancel to trigger shutdown
		cancel()
		err = <-done
	case err = <-done:
		// Process completed before readiness
	case <-time.After(5 * time.Second):
		// Timeout waiting for readiness
		cancel()
		err = <-done
	}

	// Context cancellation after readiness is expected and should be treated as success
	if errors.Is(err, context.Canceled) {
		err = nil
	}

	// Read stderr for inspection
	stderrBytes, _ := os.ReadFile(stderrFile.Name())
	stderr = bytes.NewBuffer(stderrBytes)
	return stderr, err
}
