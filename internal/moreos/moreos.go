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
	"os"
	"syscall"
)

const (
	// LineSeparator is the runtime.GOOS-specific new line or line feed. Ex. "\n"
	LineSeparator = ln
	// Exe is the runtime.GOOS-specific suffix for executables. Ex. "" unless windows which is ".exe"
	Exe = exe
)

// ProcessGroupAttr sets attributes that ensure exec.Cmd doesn't propagate signals from func-e by default.
// This is used to ensure shutdown hooks can apply
func ProcessGroupAttr() *syscall.SysProcAttr {
	return processGroupAttr() // un-exported to prevent godoc drift
}

// Interrupt attempts to interrupt the process. It doesn't necessarily kill it.
func Interrupt(p *os.Process) error {
	return interrupt(p) // un-exported to prevent godoc drift
}

// IsExecutable returns true if the input can be run as an exec.Cmd
func IsExecutable(f os.FileInfo) bool {
	return isExecutable(f)
}
