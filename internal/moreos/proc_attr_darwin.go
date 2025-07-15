// Copyright 2022 Tetrate
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

package moreos

import "syscall"

// processGroupAttr returns nil on Darwin (macOS) to keep child processes in the same process group as the parent.
// This differs from Linux where we use:
// - Setpgid: true - Creates a separate process group to allow graceful shutdown hooks
// - Pdeathsig: SIGKILL - Ensures child termination when parent dies
//
// Darwin lacks Pdeathsig support, so separate process groups would create orphaned processes
// that become zombies when func-e exits. While this approach prevents graceful shutdown
// handling (signals propagate immediately to all processes), it avoids resource leaks.
func processGroupAttr() *syscall.SysProcAttr {
	return nil
}
