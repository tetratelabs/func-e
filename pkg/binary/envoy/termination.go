// Copyright 2019 Tetrate
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

package envoy

import (
	"syscall"

	"github.com/tetratelabs/getenvoy/pkg/binary"
	"github.com/tetratelabs/log"
)

func (r *Runtime) handleTermination() {
	if r.cmd.ProcessState != nil {
		if r.cmd.ProcessState.Success() {
			log.Infof("Envoy process (PID=%d) exited successfully", r.cmd.Process.Pid)
			return
		}
		log.Infof("Envoy process (PID=%d) terminated prematurely", r.cmd.Process.Pid)
		return
	}

	// Execute all registered preTermination functions
	for _, f := range r.preTermination {
		if err := f(r); err != nil {
			log.Error(err.Error())
		}
	}

	// Forward on the SIGINT to Envoy
	log.Infof("Sending Envoy process (PID=%d) SIGINT", r.cmd.Process.Pid)
	r.cmd.Process.Signal(syscall.SIGINT) //nolint

}

// RegisterPreTermination registers the passed functions to be run after Envoy has started
// and just before GetEnvoy instructs Envoy to terminate
func (r *Runtime) RegisterPreTermination(f ...func(binary.Runner) error) {
	r.preTermination = append(r.preTermination, f...)
}
