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
	"io"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoytest"
	"github.com/tetratelabs/getenvoy/pkg/globals"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

func TestRuntime_Run(t *testing.T) {
	tests := []struct {
		name                      string
		args                      []string
		terminate                 func(r *envoy.Runtime)
		wantPreTerm, wantPostTerm bool
	}{
		{
			name:         "GetEnvoy Ctrl-C",
			terminate:    nil,
			wantPreTerm:  true,
			wantPostTerm: true,
		},
		{
			name: "Envoy interrupted externally",
			terminate: func(r *envoy.Runtime) {
				pid, e := r.GetPid()
				require.NoError(t, e)
				proc, e := os.FindProcess(pid)
				require.NoError(t, e)
				e = proc.Signal(syscall.SIGINT)
				require.NoError(t, e)
			},
		},
		{
			name:      "Envoy exited with error",
			terminate: func(r *envoy.Runtime) { time.Sleep(time.Millisecond * 100) },
			args:      []string{"envoy_exit=3"},
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			debugDir, removeDebugDir := morerequire.RequireNewTempDir(t)
			defer removeDebugDir()

			fakeEnvoy, removeFakeEnvoy := morerequire.RequireCaptureScript(t, "envoy")
			defer removeFakeEnvoy()

			fakeTimestamp := "1619574747231823000"
			workingDir := filepath.Join(debugDir, fakeTimestamp)
			fakeLogger := log.New(io.Discard, "", 0)
			o := &globals.RunOpts{EnvoyPath: fakeEnvoy, WorkingDir: workingDir, Log: fakeLogger, DebugLog: fakeLogger}
			require.NoError(t, os.MkdirAll(workingDir, 0750))

			// This ensures functions are called in the correct order
			r, preStartCalled, preTerminationCalled, postTerminationCalled := newRuntimeWithMockFunctions(t, o)

			envoytest.RequireRunTerminate(t, tc.terminate, r, tc.args...)

			// Assert appropriate functions are called
			require.True(t, *preStartCalled, "preStart was not called")
			require.Equal(t, tc.wantPreTerm, *preTerminationCalled, "expected preTermination execution to be %v", tc.wantPreTerm)
			require.Equal(t, tc.wantPostTerm, *postTerminationCalled, "expected postTermination execution to be %v", tc.wantPostTerm)

			// Ensure the debug directory was created
			files, err := os.ReadDir(debugDir)
			require.NoError(t, err)
			require.Equal(t, 1, len(files))

			// Ensure a debug archive was created
			require.Contains(t, files[0].Name(), ".tar.gz")
		})
	}
}

// This ensures functions are called in the correct order
func newRuntimeWithMockFunctions(t *testing.T, o *globals.RunOpts) (*envoy.Runtime, *bool, *bool, *bool) {
	r := envoy.NewRuntime(o)
	preStartCalled := false
	r.RegisterPreStart(func() error {
		_, err := r.GetPid()
		require.Error(t, err, "preTermination was called after process was started")
		preStartCalled = true
		return nil
	})

	preTerminationCalled := false
	r.RegisterPreTermination(func() error {
		pid, err := r.GetPid()
		require.NoError(t, err, "preTermination was called before process was started")
		_, err = os.FindProcess(pid)
		require.NoError(t, err, "preTermination was called after process was terminated")
		preTerminationCalled = true
		return nil
	})

	postTerminationCalled := false
	r.RegisterPreTermination(func() error {
		postTerminationCalled = true
		return nil
	})

	return r, &preStartCalled, &preTerminationCalled, &postTerminationCalled
}
