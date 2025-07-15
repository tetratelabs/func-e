// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package moreos

import (
	"errors"
	"os"
	"syscall"
)

func interrupt(p *os.Process) error {
	// Send SIGINT to the child PID directly
	if err := p.Signal(syscall.SIGINT); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}
	return nil
}

func ensureProcessDone(p *os.Process) error {
	// Send SIGKILL to the child PID directly
	if err := p.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}
	return nil
}

func isExecutable(f os.FileInfo) bool {
	return f.Mode()&0o111 != 0
}
