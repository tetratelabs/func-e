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

	"github.com/shirou/gopsutil/v3/process"
)

const exe = ""

func processGroupAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}

func interrupt(p *os.Process) error {
	if err := p.Signal(syscall.SIGINT); err != nil && err != os.ErrProcessDone {
		return err
	}
	return nil
}

func ensureProcessDone(p *os.Process) error {
	proc, err := process.NewProcess(int32(p.Pid))
	if err != nil {
		if err == process.ErrorProcessNotRunning {
			return nil
		}
		return err
	}

	children, err := proc.Children()
	if err != nil {
		// on macOS, when a process doesn't have children from the beginning, pgrep returns with error:
		// "exit status 1", hence we ignore the error here.
		return nil
	}

	for _, child := range children {
		if err := child.Kill(); err != nil && err != process.ErrorProcessNotRunning {
			return err
		}
	}

	return nil
}

func isExecutable(f os.FileInfo) bool {
	return f.Mode()&0111 != 0
}
