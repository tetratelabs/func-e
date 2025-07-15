// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

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
