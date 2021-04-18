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

package envoy

import (
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/binary"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

func TestRuntime_RunPath(t *testing.T) {
	tests := []struct {
		name                      string
		args                      []string
		killerFunc                func(*Runtime)
		wantPreTerm, wantPostTerm bool
	}{
		{
			name:         "GetEnvoy shot first",
			killerFunc:   func(r *Runtime) { r.FakeInterrupt() },
			wantPreTerm:  true,
			wantPostTerm: true,
		},
		{
			name:       "Envoy shot first",
			killerFunc: func(r *Runtime) { r.cmd.Process.Signal(syscall.SIGINT) },
		},
		{
			name:       "Envoy simulate error",
			killerFunc: func(r *Runtime) { time.Sleep(time.Millisecond * 100) },
			args:       []string{"error"},
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			// This ensures functions are called in the correct order
			r, preStartCalled, preTerminationCalled, postTerminationCalled := newRuntimeWithMockFunctions(t)
			tempDir, revertTempDir := morerequire.RequireNewTempDir(t)
			defer revertTempDir()
			r.store = tempDir

			wd, err := os.Getwd()
			require.NoError(t, err, "error reading working directory")
			sleep := filepath.Join(wd, "testdata", "sleep.sh")

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				r.RunPath(sleep, tc.args)
			}()

			expectedStatus := binary.StatusInitializing // because we won't have an admin URL to check
			require.Eventually(t, func() bool {
				return r.Status() == expectedStatus || r.Status() == binary.StatusTerminated
			}, 10*time.Second, 100*time.Millisecond, "never achieved StatusStarted or StatusTerminated")
			require.Equal(t, expectedStatus, r.Status(), "never achieved StatusStarted or StatusTerminated")

			tc.killerFunc(r)
			wg.Wait()

			// Assert appropriate functions are called
			require.True(t, *preStartCalled, "preStart was not called")
			require.Equal(t, tc.wantPreTerm, *preTerminationCalled, "expected preTermination execution to be %v", tc.wantPreTerm)
			require.Equal(t, tc.wantPostTerm, *postTerminationCalled, "expected postTermination execution to be %v", tc.wantPostTerm)
		})
	}
}

// This ensures functions are called in the correct order
func newRuntimeWithMockFunctions(t *testing.T) (*Runtime, *bool, *bool, *bool) {
	preStartCalled := false
	preStart := func(r *Runtime) {
		r.RegisterPreStart(func(r binary.Runner) error {
			r, _ = r.(*Runtime)
			if r.Status() > binary.StatusStarting {
				t.Error("preStart was called after process has started")
			}
			preStartCalled = true
			return nil
		})
	}

	preTerminationCalled := false
	preTermination := func(r *Runtime) {
		r.RegisterPreTermination(func(r binary.Runner) error {
			r, _ = r.(*Runtime)
			if r.Status() < binary.StatusStarted {
				t.Error("preTermination was called before process was started")
			}
			if r.Status() > binary.StatusReady {
				t.Error("preTermination was called after process was terminated")
			}
			preTerminationCalled = true
			return nil
		})
	}
	postTerminationCalled := false
	postTermination := func(r *Runtime) {
		r.RegisterPreTermination(func(r binary.Runner) error {
			postTerminationCalled = true
			return nil
		})
	}
	runtime, _ := NewRuntime(preStart, preTermination, postTermination)

	return runtime.(*Runtime), &preStartCalled, &preTerminationCalled, &postTerminationCalled
}
