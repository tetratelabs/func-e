// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import "syscall"

// processGroupAttr ensures graceful shutdown (shutdown hooks) is attempted:
//   - Setpgid: Creates a new process group to isolate Envoy from parent signals, so that shutdown hooks can be applied
//     before propagating the signal to Envoy.
//   - Pdeathsig: Ensures Envoy terminates when func-e exits, preventing orphaned processes and resource leaks.
func processGroupAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGKILL}
}
