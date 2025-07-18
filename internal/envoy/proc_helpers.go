// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"errors"
	"os"
	"syscall"
)

// interrupt attempts to interrupt the process. It doesn't necessarily kill it.
func interrupt(p *os.Process) error {
	// Send SIGINT to the child PID directly
	if err := p.Signal(syscall.SIGINT); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}
	return nil
}

// ensureProcessDone makes sure the process has exited completely.
func ensureProcessDone(p *os.Process) error {
	// Send SIGKILL to the child PID directly
	if err := p.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}
	return nil
}
