// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package moreos

import (
	"fmt"
	"io"
	"os"
	"syscall"
)

const (
	// Exe is the runtime.GOOS-specific suffix for executables.
	Exe = ""
	// OSDarwin is a Platform.OS a.k.a. "macOS"
	OSDarwin = "darwin"
	// OSLinux is a Platform.OS
	OSLinux = "linux"
)

// Errorf is like fmt.Errorf. TODO: remove this as this only exists due to Windows builds support before.
func Errorf(format string, a ...interface{}) error {
	err := fmt.Errorf(format, a...)
	return err
}

// ReplacePathSeparator returns the input.
// TODO: remove this as this only exists due to Windows builds support before.
func ReplacePathSeparator(input string) string {
	return input
}

// Sprintf is like Fprintf is like fmt.Sprintf.
// TODO: remove this as this only exists due to Windows builds support before.
func Sprintf(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}

// Fprintf is like fmt.Fprintf, but handles EOL according runtime.GOOS. See Sprintf for notes.
// TODO: remove this as this only exists due to Windows builds support before.
func Fprintf(w io.Writer, format string, a ...interface{}) {
	_, _ = fmt.Fprintf(w, format, a...)
}

// ProcessGroupAttr sets attributes that ensure exec.Cmd doesn't propagate signals from func-e by default.
// This is used to ensure shutdown hooks can apply
func ProcessGroupAttr() *syscall.SysProcAttr {
	return processGroupAttr() // un-exported to prevent godoc drift
}

// Interrupt attempts to interrupt the process. It doesn't necessarily kill it.
func Interrupt(p *os.Process) error {
	return interrupt(p) // un-exported to prevent godoc drift
}

// EnsureProcessDone makes sure the process has exited completely.
func EnsureProcessDone(p *os.Process) error {
	return ensureProcessDone(p) // un-exported to prevent godoc drift
}

// IsExecutable returns true if the input can be run as an exec.Cmd
func IsExecutable(f os.FileInfo) bool {
	return isExecutable(f)
}
