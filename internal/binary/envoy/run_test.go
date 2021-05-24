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

	"github.com/tetratelabs/getenvoy/internal/binary/envoy"
	"github.com/tetratelabs/getenvoy/internal/binary/envoytest"
	"github.com/tetratelabs/getenvoy/internal/globals"
	"github.com/tetratelabs/getenvoy/internal/test/morerequire"
)

func TestRuntime_Run(t *testing.T) {
	tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
	defer removeTempDir()

	debugDir := filepath.Join(tempDir, "debug")
	fakeTimestamp := "1619574747231823000"
	workingDir := filepath.Join(debugDir, fakeTimestamp)

	// "quiet" as we aren't testing the environment envoy runs in
	fakeEnvoy := filepath.Join(tempDir, "quiet")
	morerequire.RequireCaptureScript(t, fakeEnvoy)

	tests := []struct {
		name                           string
		args                           []string
		terminate                      func(r *envoy.Runtime)
		expectedStdout, expectedStderr string
		expectedErr                    string
		expectedHooks                  []string
	}{
		{
			name:      "GetEnvoy Ctrl-C",
			terminate: nil,
			// Don't warn the user when they exited the process
			expectedStdout: fmt.Sprintf(`starting: %s
working directory: %s
`, fakeEnvoy, workingDir),
			expectedStderr: "started\ncaught SIGINT\n",
			expectedHooks:  []string{"preStart", "preTermination", "postTermination"},
		},
		// We don't test envoy dying from an external signal as it isn't reported back to the getenvoy process and
		// Envoy returns exit status zero on anything except kill -9. We can't test kill -9 with a fake shell script.
		{
			name:      "Envoy exited with error",
			terminate: func(r *envoy.Runtime) { time.Sleep(time.Millisecond * 100) },
			args:      []string{"quiet_exit=3"},
			expectedStdout: fmt.Sprintf(`starting: %s quiet_exit=3
working directory: %s
`, fakeEnvoy, workingDir),
			expectedStderr: "started\n",
			expectedErr:    "envoy exited with status: 3",
			expectedHooks:  []string{"preStart"},
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			o := &globals.RunOpts{EnvoyPath: fakeEnvoy, WorkingDir: workingDir}
			require.NoError(t, os.MkdirAll(workingDir, 0750))

			stdout := new(bytes.Buffer)
			stderr := new(bytes.Buffer)
			r, hooksCalled := newRuntimeWithMockHooks(t, stdout, stderr, o)

			err := envoytest.RequireRunTerminate(t, tc.terminate, r, tc.args...)
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

			// Ensure the working directory was deleted, and the debug directory only contains the archive
			files, err := os.ReadDir(debugDir)
			require.NoError(t, err)
			require.Equal(t, 1, len(files))
			archive := filepath.Join(debugDir, files[0].Name())
			require.Equal(t, filepath.Join(debugDir, fakeTimestamp+".tar.gz"), archive)

			// Cleanup for the next run
			require.NoError(t, os.Remove(archive))
		})
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
