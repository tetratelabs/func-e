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
	"os/signal"
	"syscall"
)

var (
	// TODO(yskopets): at the moment, we only support Linux and Mac
	shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

	terminate = func() {
		os.Exit(1)
	}
)

// SetupSignalHandler registers for SIGTERM and SIGINT signals, and returns a
// channel that will be closed as soon as one of those signals is received.
// If a signal is received for the second time, the program will be terminated
// with exit code 1.
//
// The provided context is used to undo registration for signals if the context
// becomes done before a signal is received for the second time.
// Consequently, the returned channel might never get closed if the context
// becomes doneo prior to the first signal.
func SetupSignalHandler(ctx context.Context) <-chan struct{} {
	stopCh := make(chan struct{})

	signalCh := make(chan os.Signal, 2)
	signal.Notify(signalCh, shutdownSignals...)
	go func() {
		defer signal.Stop(signalCh)

		// on first signal, close the stop channel
		select {
		case <-ctx.Done():
			return
		case <-signalCh:
			close(stopCh)
		}

		// on second signal, terminate the program
		select {
		case <-ctx.Done():
			return
		case <-signalCh:
			terminate()
		}
	}()

	return stopCh
}
