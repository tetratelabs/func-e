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

package moreos

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
)

const (
	exe = ".exe"
	// from WinError.h, but not defined for some reason in types_windows.go
	errorInvalidParameter = 87
)

func processGroupAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP, // Stop Ctrl-Break propagation to allow shutdown-hooks
	}
}

// interrupt contains signal_windows_test.go sendCtrlBreak() as there's no main source with the same.
func interrupt(p *os.Process) error {
	pid := p.Pid
	d, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return errorInterrupting(pid, err)
	}
	proc, err := d.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		return errorInterrupting(pid, err)
	}
	r, _, err := proc.Call(syscall.CTRL_BREAK_EVENT, uintptr(pid))
	if r == 0 { // because err != nil on success "The operation completed successfully"
		return errorInterrupting(pid, err)
	}
	return nil
}

func errorInterrupting(pid int, err error) error {
	return fmt.Errorf("couldn't Interrupt pid(%d): %w", pid, err)
}

// ensureProcessDone attempts to work around flakey logic in os.Process Wait on Windows. This code block should be
// revisited if https://golang.org/issue/25965 is solved.
func ensureProcessDone(p *os.Process) error {
	// Process.handle is not exported. Lookup the process again, using logic similar to exec_windows/findProcess()
	const da = syscall.STANDARD_RIGHTS_READ | syscall.PROCESS_TERMINATE |
		syscall.PROCESS_QUERY_INFORMATION | syscall.SYNCHRONIZE
	h, e := syscall.OpenProcess(da, true, uint32(p.Pid))
	if e != nil {
		if errno, ok := e.(syscall.Errno); ok && uintptr(errno) == errorInvalidParameter {
			return nil // don't error if the process isn't around anymore
		}
		return os.NewSyscallError("OpenProcess", e)
	}
	defer syscall.CloseHandle(h) //nolint:errcheck

	// Try to wait for the process to close naturally first, using logic from exec_windows/findProcess()
	// Difference here, is we are waiting 100ms not infinite. If there's a timeout, we kill the proc.
	s, e := syscall.WaitForSingleObject(h, 100)
	switch s {
	case syscall.WAIT_OBJECT_0:
		return nil // process is no longer around
	case syscall.WAIT_TIMEOUT:
		return syscall.TerminateProcess(h, uint32(0)) // kill, but don't effect the exit code
	case syscall.WAIT_FAILED:
		return os.NewSyscallError("WaitForSingleObject", e)
	default:
		return errors.New("os: unexpected result from WaitForSingleObject")
	}
}

func isExecutable(f os.FileInfo) bool { // In windows, we cannot read execute bit
	return strings.HasSuffix(f.Name(), ".exe")
}
