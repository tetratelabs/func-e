// Copyright 2021 Tetrate
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
	"testing"
)

// RequireEnvoyPid returns the pid of the child process
func RequireEnvoyPid(t *testing.T, r *Runtime) int {
	if r.cmd == nil || r.cmd.Process == nil {
		t.Fatal("envoy process not yet started")
	}
	return r.cmd.Process.Pid
}
