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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tetratelabs/getenvoy/pkg/binary"
)

func TestRuntime_RunPath(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		killerFunc  func(*Runtime)
		wantPreTerm bool
	}{
		{
			name:        "GetEnvoy shot first",
			killerFunc:  func(r *Runtime) { r.SendSignal(syscall.SIGINT) },
			wantPreTerm: true,
		},
		{
			name:        "Envoy shot first",
			killerFunc:  func(r *Runtime) { r.cmd.Process.Signal(syscall.SIGINT) },
			wantPreTerm: false,
		},
		{
			name:        "Envoy simulate error",
			killerFunc:  func(r *Runtime) { time.Sleep(time.Millisecond * 100) },
			args:        []string{"error"},
			wantPreTerm: false,
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			// This ensures functions are called in the correct order
			r, preStartCalled, preTerminationCalled := newRuntimeWithMockFunctions(t)
			tmpDir, _ := ioutil.TempDir("", "getenvoy-test-")
			defer os.RemoveAll(tmpDir)
			r.store = tmpDir

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				r.RunPath(filepath.Join("testdata", "sleep.sh"), tc.args)
			}()

			r.Wait(binary.StatusStarted)
			tc.killerFunc(r)
			wg.Wait()

			// Assert appropriate functions are called
			assert.True(t, *preStartCalled, "preStart was not called")
			assert.Equal(t, tc.wantPreTerm, *preTerminationCalled, fmt.Sprintf("expected preTermination execution to be %v", tc.wantPreTerm))
		})
	}
}

// This ensures functions are called in the correct order
func newRuntimeWithMockFunctions(t *testing.T) (*Runtime, *bool, *bool) {
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
	runtime, _ := NewRuntime(preStart, preTermination)

	return runtime.(*Runtime), &preStartCalled, &preTerminationCalled
}
