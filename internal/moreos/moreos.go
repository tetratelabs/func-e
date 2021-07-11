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
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"syscall"
)

const (
	// Exe is the runtime.GOOS-specific suffix for executables. Ex. "" unless windows which is ".exe"
	Exe = exe
	// OSDarwin is a Platform.OS a.k.a. "macOS"
	OSDarwin = "darwin"
	// OSLinux is a Platform.OS
	OSLinux = "linux"
	// OSWindows is a Platform.OS
	OSWindows = "windows"
)

// Sprintf is like Fprintf is like fmt.Sprintf, except any '\n' in the format are converted according to runtime.GOOS.
// This allows us to be consistent with Envoy, which handles \r\n on Windows.
// See also https://github.com/golang/go/issues/28822
func Sprintf(format string, a ...interface{}) string {
	if runtime.GOOS != OSWindows {
		return fmt.Sprintf(format, a...)
	}
	return fmt.Sprintf(strings.ReplaceAll(format, "\n", "\r\n"), a...)
}

// Fprintf is like fmt.Fprintf, but handles EOL according runtime.GOOS. See Sprintf for behavioural notes.
func Fprintf(w io.Writer, format string, a ...interface{}) (n int, err error) {
	if runtime.GOOS != OSWindows {
		return fmt.Fprintf(w, format, a...)
	}
	return fmt.Fprint(w, Sprintf(format))
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

// IsExecutable returns true if the input can be run as an exec.Cmd
func IsExecutable(f os.FileInfo) bool {
	return isExecutable(f)
}
