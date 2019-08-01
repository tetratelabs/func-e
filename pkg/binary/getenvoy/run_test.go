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

package getenvoy

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
			killerFunc:  func(r *Runtime) { r.signals <- syscall.SIGINT },
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
			r, _ := New()
			tmpDir, _ := ioutil.TempDir("", "getenvoy-test-")
			defer os.RemoveAll(tmpDir)
			r.local = tmpDir

			// These will t.Error if not called in correct order
			preStartCalled := registerPreStart(t, r)
			preTerminationCalled := registerPreTermination(t, r)

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				r.RunPath(filepath.Join("testdata", "sleep.sh"), tc.args)
			}()

			waitForProcessStart(r)
			tc.killerFunc(r)
			wg.Wait()

			// Assert appropriate functions are called
			assert.True(t, *preStartCalled, "preStart was not called")
			assert.Equal(t, tc.wantPreTerm, *preTerminationCalled, fmt.Sprintf("expected preTermination execution to be %v", tc.wantPreTerm))
		})
	}
}

func waitForProcessStart(r *Runtime) {
	for r.cmd == nil || r.cmd.Process == nil {
		time.Sleep(time.Millisecond)
	}
}

func registerPreStart(t *testing.T, r *Runtime) *bool {
	called := false
	r.registerPreStart(func() error {
		if r.cmd != nil && r.cmd.Process != nil {
			t.Error("preStart was called after process has started")
		}
		called = true
		return nil
	})
	return &called
}

func registerPreTermination(t *testing.T, r *Runtime) *bool {
	called := false
	r.registerPreTermination(func() error {
		if r.cmd != nil && r.cmd.Process == nil {
			t.Error("preTermination was called before process was started")
		}
		if r.cmd != nil && r.cmd.ProcessState != nil {
			t.Error("preTermination was called after process was terminated")
		}
		called = true
		return nil
	})
	return &called
}
