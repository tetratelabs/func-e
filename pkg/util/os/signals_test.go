// Copyright 2020 Tetrate
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

package os

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestShutdownSignals(t *testing.T) {
	// This is a base-case, just verifying the default values
	require.Equal(t, []os.Signal{syscall.SIGINT, syscall.SIGTERM}, shutdownSignals)
}

func TestSetupSignalHandlerCatchesRelevantSignalAndClosesTheChannel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopCh := SetupSignalHandler(ctx)

	// send a relevant signal to the process
	err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	require.NoError(t, err)

	requireSignal(t, syscall.SIGINT, stopCh)
	requireChannelClosed(t, stopCh)
}

func TestSetupSignalHandlerIgnoresWhenContextCanceledBeforeSignal(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopCh := SetupSignalHandler(ctx)

	cancel() // context canceled

	// send a relevant signal to the process
	err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	require.NoError(t, err)

	// The signal is ignored because the context closed prior to receiving it
	requireNoSignal(t, stopCh)
}

func TestSetupSignalHandlerIgnoresIrrelevantSignal(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopCh := SetupSignalHandler(ctx)

	// send an irrelevant signal to the process
	err := syscall.Kill(syscall.Getpid(), syscall.SIGUSR2)
	require.NoError(t, err)

	requireNoSignal(t, stopCh)
}

func requireSignal(t *testing.T, expected os.Signal, ch <-chan os.Signal) {
	require.Eventually(t, func() bool {
		select {
		case s, ok := <-ch:
			return s == expected && ok
		default:
			return false
		}
	}, 50*time.Millisecond, 10*time.Millisecond)
}

func requireChannelClosed(t *testing.T, ch <-chan os.Signal) {
	select {
	case _, ok := <-ch:
		require.False(t, ok)
	default:
		t.Fatal()
	}
}

// requireNoSignal ensures the channel is open, but there are no signals in it.
func requireNoSignal(t *testing.T, ch <-chan os.Signal) {
	require.Never(t, func() bool {
		select {
		case v, ok := <-ch:
			return v != nil || ok
		default:
			return false
		}
	}, 100*time.Millisecond, 10*time.Millisecond)
}

// overrideTerminateWithBool returns a boolean made true on terminate. The function returned reverts the original.
func overrideTerminateWithBool() (*bool, func()) {
	previous := terminate
	closed := false
	terminate = func() {
		closed = true
	}
	return &closed, func() {
		terminate = previous
	}
}

func TestSetupSignalHandlerTerminatesOnSecondRelevantSignal(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	terminated, revertTerminate := overrideTerminateWithBool()
	defer revertTerminate()

	stopCh := SetupSignalHandler(ctx)

	// send a relevant signal to the process
	err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	require.NoError(t, err)

	// First relevant signal closes the channel, but doesn't terminate
	requireSignal(t, syscall.SIGINT, stopCh)
	requireChannelClosed(t, stopCh)
	require.False(t, *terminated)

	// send another relevant signal to the process
	err = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	require.NoError(t, err)

	// Second relevant signal terminates the process
	require.Eventually(t, func() bool {
		return *terminated
	}, 50*time.Millisecond, 10*time.Millisecond)
}

func TestSetupSignalHandlerDoesntTerminateWhenContextCanceledBeforeSecondRelevantSignal(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	terminated, revertTerminate := overrideTerminateWithBool()
	defer revertTerminate()

	stopCh := SetupSignalHandler(ctx)

	// send a relevant signal to the process
	err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	require.NoError(t, err)

	// First relevant signal closes the channel, but doesn't terminate
	requireSignal(t, syscall.SIGINT, stopCh)
	requireChannelClosed(t, stopCh)
	require.False(t, *terminated)

	cancel() // context canceled

	// send another relevant signal to the process
	err = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	require.NoError(t, err)

	// Second relevant signal doesn't terminate the process
	require.Never(t, func() bool {
		return *terminated
	}, 50*time.Millisecond, 10*time.Millisecond)
}
