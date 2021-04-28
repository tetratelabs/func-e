// +build !windows

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

package exec

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
	ioutil "github.com/tetratelabs/getenvoy/pkg/util/io"
)

func TestRunPipesStdoutAndStderr(t *testing.T) {
	stdin, stdout, stderr := new(bytes.Buffer), new(bytes.Buffer), new(bytes.Buffer)
	stdin.WriteString("input to stdin\n")

	testScript, removeTestScript := morerequire.RequireCaptureScript(t, "test")
	defer removeTestScript()

	cmd := exec.Command(testScript, "text_exit=0")
	err := Run(cmd, ioutil.StdStreams{In: stdin, Out: stdout, Err: stderr})

	require.NoError(t, err, `error running [%v]`, cmd)
	expectedStdout := fmt.Sprintf("test wd: %s\ntest bin: %s\ntest args: text_exit=0\n", morerequire.RequireAbs(t, "."), testScript)
	require.Equal(t, expectedStdout, stdout.String(), `invalid stdout running [%v]`, cmd)
	require.Equal(t, "test stderr\n", stderr.String(), `invalid stderr running [%v]`, cmd)
}

func TestRunErrorWrapsCause(t *testing.T) {
	testScript, removeTestScript := morerequire.RequireCaptureScript(t, "test")
	defer removeTestScript()

	tests := []struct {
		name              string
		path              string
		expectedErr       string
		expectedErrTarget error
	}{
		{
			name:              "invalid path",
			path:              "/invalid/path",
			expectedErr:       `failed to execute an external command "/invalid/path test_exit=123": fork/exec /invalid/path: no such file or directory`,
			expectedErrTarget: new(os.PathError),
		},
		{
			name:              "exit status",
			path:              testScript,
			expectedErr:       fmt.Sprintf(`failed to execute an external command "%s test_exit=123": exit status 123`, testScript),
			expectedErrTarget: new(exec.ExitError),
		},
	}

	for _, test := range tests {
		tt := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tt.name, func(t *testing.T) {
			stdin, stdout, stderr := new(bytes.Buffer), new(bytes.Buffer), new(bytes.Buffer)

			cmd := exec.Command(tt.path, "test_exit=123")
			err := Run(cmd, ioutil.StdStreams{In: stdin, Out: stdout, Err: stderr})

			// Verify the command failed with the expected error.
			require.EqualError(t, err, tt.expectedErr, `expected an error running [%v]`, cmd)

			var runErr *RunError
			require.ErrorAs(t, err, &runErr, `expected a RunError running [%v]`, cmd)
			require.Equal(t, tt.path+" test_exit=123", runErr.Cmd(), `expected RunError.Cmd() to contain path and args`)
			require.ErrorAs(t, runErr.Cause(), &tt.expectedErrTarget, `expected RunError.Cause() to wrap the original error`)
		})
	}
}

func TestRunShutdownError(t *testing.T) {
	for _, s := range []os.Signal{syscall.SIGINT, syscall.SIGTERM} {
		s := s // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(s.String(), func(t *testing.T) {
			cmd := exec.Command("sleep", "1")

			// Fake the actual signal handler. Signals sent to stopCh aren't actually sent to the process.
			stopCh := make(chan os.Signal, 1)
			revertSetupSignalHandler := overrideSetupSignalHandler(func(ctx context.Context) <-chan os.Signal {
				return stopCh
			})
			defer revertSetupSignalHandler()

			// Since we are only testing signals manifest in errors, kill quickly.
			revertKillTimeout := overrideKillTimeout(1 * time.Millisecond)
			defer revertKillTimeout()

			// We run the command in a goroutine as we need to signal the process, which results in an error.
			errCh := make(chan error)
			go func() {
				defer close(errCh)
				errCh <- Run(cmd, ioutil.StdStreams{In: new(bytes.Buffer), Out: new(bytes.Buffer), Err: new(bytes.Buffer)})
			}()

			stopCh <- s // Trigger exec.terminate()
			close(stopCh)

			// Verify the process shutdown and raised an error due to the signal we caught
			for err := range errCh {
				require.Equal(t, newShutdownError(s), err)
			}
		})
	}
}

func TestRunSendsSIGTERMIfProcessStillRunningAfterStopSignal(t *testing.T) {
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	stdio := ioutil.StdStreams{In: new(bytes.Buffer), Out: stdout, Err: stderr}

	sleepScript, removeSleepScript := morerequire.RequireCaptureScript(t, "sleep")
	defer removeSleepScript()
	cmd := exec.Command(sleepScript)

	// Fake the actual signal handler. Signals sent to stopCh aren't actually sent to the process.
	stopCh := make(chan os.Signal, 1)
	revertSetupSignalHandler := overrideSetupSignalHandler(func(ctx context.Context) <-chan os.Signal {
		return stopCh
	})
	defer revertSetupSignalHandler()

	// We run the command in a goroutine as we need to SIGTERM the process, which results in an error.
	errCh := make(chan error)
	go func() {
		errCh <- Run(cmd, stdio)
		close(errCh)
	}()

	// Wait for shell script to start
	require.Eventually(t, func() bool {
		return strings.Contains(stdout.String(), "sleep wd")
	}, 1*time.Second, 10*time.Millisecond)

	// Send a fake signal to our signal handler which invokes the stop channel
	fakeSignal := syscall.Signal(0)
	stopCh <- fakeSignal
	close(stopCh)

	// Wait for the err channel which means the process shutdown.
	for err := range errCh {
		require.Equal(t, newShutdownError(fakeSignal), err)
	}

	// Verify the program received a SIGTERM because the process was still alive upon exec.terminate()
	require.Equal(t, "sleep stderr\nSIGTERM caught!\n", stderr.String(), `invalid stderr running [%v]`, cmd)
}

// overrideSetupSignalHandler sets setupSignalHandler to a function. Doing so prevents the process from seeing signals.
// The function returned reverts the original value.
func overrideSetupSignalHandler(override func(ctx context.Context) <-chan os.Signal) func() {
	previous := setupSignalHandler
	setupSignalHandler = override
	return func() {
		setupSignalHandler = previous
	}
}

// overrideKillTimeout sets killTimeout to a short value as when we override setupSignalHandler, we need to kill faster.
// The function returned reverts the original value.
func overrideKillTimeout(override time.Duration) func() {
	previous := killTimeout
	killTimeout = override
	return func() {
		killTimeout = previous
	}
}
