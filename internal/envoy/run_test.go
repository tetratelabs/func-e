// Copyright 2019 Tetrate
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

package envoy_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/envoy"
	"github.com/tetratelabs/getenvoy/internal/globals"
	"github.com/tetratelabs/getenvoy/internal/test"
	"github.com/tetratelabs/getenvoy/internal/test/morerequire"
)

func TestRuntime_Run(t *testing.T) {
	tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
	defer removeTempDir()

	runsDir := filepath.Join(tempDir, "runs")
	fakeTimestamp := "1619574747231823000"
	runDir := filepath.Join(runsDir, fakeTimestamp)
	adminFlag := fmt.Sprintf("--admin-address-path %s/admin-address.txt", runDir)

	// "quiet" as we aren't testing the environment envoy runs in
	fakeEnvoy := filepath.Join(tempDir, "quiet")
	morerequire.RequireCaptureScript(t, fakeEnvoy)

	tests := []struct {
		name                           string
		args                           []string
		terminate                      func()
		expectedStdout, expectedStderr string
		expectedErr                    string
		expectedHooks                  []string
	}{
		{
			name: "GetEnvoy Ctrl-C",
			// Don't warn the user when they exited the process
			expectedStdout: fmt.Sprintln("starting:", fakeEnvoy, adminFlag),
			expectedStderr: "started\ncaught SIGINT\n",
			expectedHooks:  []string{"preStart", "preTermination", "postTermination"},
		},
		// We don't test envoy dying from an external signal as it isn't reported back to the getenvoy process and
		// Envoy returns exit status zero on anything except kill -9. We can't test kill -9 with a fake shell script.
		{
			name:           "Envoy exited with error",
			terminate:      func() { time.Sleep(time.Millisecond * 100) },
			args:           []string{"quiet_exit=3"},
			expectedStdout: fmt.Sprintln("starting:", fakeEnvoy, "quiet_exit=3", adminFlag),
			expectedStderr: "started\n",
			expectedErr:    "envoy exited with status: 3",
			expectedHooks:  []string{"preStart"},
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			o := &globals.RunOpts{EnvoyPath: fakeEnvoy, RunDir: runDir}
			require.NoError(t, os.MkdirAll(runDir, 0750))

			stdout := new(bytes.Buffer)
			stderr := new(bytes.Buffer)
			r, hooksCalled := newRuntimeWithMockHooks(t, stdout, stderr, o)

			terminate := tc.terminate
			if terminate == nil {
				terminate = interrupt(r)
			}

			// tee the error stream so we can look for the "started" line without consuming it.
			errCopy := new(bytes.Buffer)
			r.Err = io.MultiWriter(r.Err, errCopy)
			err := test.RequireRunTerminate(t, terminate, r, errCopy, tc.args...)

			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}

			// Assert appropriate hooks are called
			require.Equal(t, tc.expectedHooks, *hooksCalled)

			// Validate we ran what we thought we did
			require.Equal(t, tc.expectedStdout, stdout.String())
			require.Equal(t, tc.expectedStderr, stderr.String())

			// Ensure the working directory was deleted, and the "run" directory only contains the archive
			files, err := os.ReadDir(runsDir)
			require.NoError(t, err)
			require.Equal(t, 1, len(files))
			archive := filepath.Join(runsDir, files[0].Name())
			require.Equal(t, filepath.Join(runsDir, fakeTimestamp+".tar.gz"), archive)

			// Cleanup for the next run
			require.NoError(t, os.Remove(archive))
		})
	}
}

func interrupt(r *envoy.Runtime) func() {
	return func() {
		fakeInterrupt := r.FakeInterrupt
		if fakeInterrupt != nil {
			fakeInterrupt()
		}
	}
}

// This ensures functions are called in the correct order
func newRuntimeWithMockHooks(t *testing.T, stdout, stderr io.Writer, o *globals.RunOpts) (*envoy.Runtime, *[]string) {
	r := envoy.NewRuntime(o)
	r.Out = stdout
	r.Err = stderr
	var hooks []string
	r.RegisterPreStart(func() error {
		_, err := r.GetEnvoyPid()
		require.Error(t, err, "preTermination was called after process was started")
		hooks = append(hooks, "preStart")
		return nil
	})

	r.RegisterPreTermination(func() error {
		pid, err := r.GetEnvoyPid()
		require.NoError(t, err, "preTermination was called before process was started")
		_, err = os.FindProcess(pid)
		require.NoError(t, err, "preTermination was called after process was terminated")
		hooks = append(hooks, "preTermination")
		return nil
	})

	r.RegisterPreTermination(func() error {
		hooks = append(hooks, "postTermination")
		return nil
	})

	return r, &hooks
}
