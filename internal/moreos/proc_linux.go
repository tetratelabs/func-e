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

const exe = ""

func processGroupAttr() *syscall.SysProcAttr {
	// Pdeathsig aims to ensure the process group is cleaned up even if this process dies. When func-e
	// dies, the process (envoy) will get SIGKILL.
	return &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGKILL}
}

func interrupt(p *os.Process) error {
	if err := p.Signal(syscall.SIGINT); err != nil && err != os.ErrProcessDone {
		return err
	}
	return nil
}

func ensureProcessDone(p *os.Process) error {
	if err := p.Kill(); err != nil && err != os.ErrProcessDone {
		return err
	}
	return nil
}

func isExecutable(f os.FileInfo) bool {
	return f.Mode()&0111 != 0
}
