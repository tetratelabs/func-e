// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	// testProcessTimeout is how long we wait for a process to exit in tests.
	// Set to 40% of shutdownTimeout since the process should die quickly after
	// handleShutdown completes (which already waited up to shutdownTimeout).
	testProcessTimeout = shutdownTimeout * 2 / 5

	// testSlowHookDuration simulates a hook that would exceed shutdownTimeout.
	// Set to 6x shutdownTimeout to ensure it clearly exceeds the timeout.
	testSlowHookDuration = shutdownTimeout * 6

	// testShutdownBuffer is extra time beyond shutdownTimeout to account for
	// goroutine scheduling and process termination overhead.
	// Set to 40% of shutdownTimeout to be proportional but reasonable.
	testShutdownBuffer = shutdownTimeout * 2 / 5
)

// TestHandleShutdown_PanicInHook tests that a panic in a shutdown hook
// doesn't prevent Envoy from being terminated
func TestHandleShutdown_PanicInHook(t *testing.T) {
	// Create a test command that sleeps (simulating Envoy)
	cmd := exec.Command("sleep", fmt.Sprintf("%d", int(testSlowHookDuration.Seconds())))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	require.NoError(t, cmd.Start())

	r := &Runtime{
		cmd: cmd,
		Out: os.Stdout,
		logf: func(format string, args ...interface{}) {
			t.Logf(format, args...)
		},
	}

	// Register a hook that panics
	r.RegisterShutdownHook(func(ctx context.Context) error {
		panic("shutdown hook panic!")
	})

	// Register another hook to verify execution continues
	hookExecuted := false
	r.RegisterShutdownHook(func(ctx context.Context) error {
		hookExecuted = true
		return nil
	})

	// Call handleShutdown - it should recover from panic and still kill the process
	r.handleShutdown()

	// Wait for process to exit
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Process exited as expected
	case <-time.After(testProcessTimeout):
		t.Fatal("Process didn't exit within timeout")
	}

	require.True(t, hookExecuted, "Other hooks should still execute despite panic")
}

// TestHandleShutdown_MultipleHooksPanic tests that multiple panicking hooks
// don't prevent termination
func TestHandleShutdown_MultipleHooksPanic(t *testing.T) {
	cmd := exec.Command("sleep", fmt.Sprintf("%d", int(testSlowHookDuration.Seconds())))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	require.NoError(t, cmd.Start())

	r := &Runtime{
		cmd: cmd,
		Out: os.Stdout,
		logf: func(format string, args ...interface{}) {
			t.Logf(format, args...)
		},
	}

	// Register multiple panicking hooks
	for i := 0; i < 3; i++ {
		idx := i
		r.RegisterShutdownHook(func(ctx context.Context) error {
			panic(fmt.Sprintf("hook %d panic!", idx))
		})
	}

	// Call handleShutdown
	r.handleShutdown()

	// Wait for process to exit
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Process exited as expected
	case <-time.After(testProcessTimeout):
		t.Fatal("Process didn't exit within timeout despite multiple panics")
	}
}

// TestHandleShutdown_SlowHook tests that slow hooks don't prevent termination
func TestHandleShutdown_SlowHook(t *testing.T) {
	cmd := exec.Command("sleep", fmt.Sprintf("%d", int(testSlowHookDuration.Seconds())))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	require.NoError(t, cmd.Start())

	r := &Runtime{
		cmd: cmd,
		Out: os.Stdout,
		logf: func(format string, args ...interface{}) {
			t.Logf(format, args...)
		},
	}

	// Register a slow hook
	r.RegisterShutdownHook(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(testSlowHookDuration):
			return nil
		}
	})

	start := time.Now()
	r.handleShutdown()
	elapsed := time.Since(start)

	// Should complete within shutdown timeout + buffer for overhead
	require.Less(t, elapsed, shutdownTimeout+testShutdownBuffer, "Shutdown should respect timeout")

	// Wait for process to exit
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Process exited as expected
	case <-time.After(testProcessTimeout):
		t.Fatal("Process didn't exit within timeout even with slow hooks")
	}
}

// TestEnsureProcessDone tests that EnsureProcessDone kills the process
func TestEnsureProcessDone(t *testing.T) {
	cmd := exec.Command("sleep", fmt.Sprintf("%d", int(testSlowHookDuration.Seconds())))
	require.NoError(t, cmd.Start())

	// Ensure process is running
	require.NoError(t, cmd.Process.Signal(syscall.Signal(0)))

	// Call ensureProcessDone
	err := ensureProcessDone(cmd.Process)
	require.NoError(t, err)

	// Wait for process to exit
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Process exited as expected
	case <-time.After(testProcessTimeout / 2):
		// Use half of testProcessTimeout since EnsureProcessDone sends SIGKILL
		// which should terminate the process immediately
		t.Fatal("Process didn't exit after EnsureProcessDone")
	}
}
