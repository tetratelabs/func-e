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

package debug

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/tetratelabs/getenvoy/pkg/binary"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/log"
)

// EnableNodeCollection is a preset option that registers collection of node level information for debugging
var EnableNodeCollection = func(r *envoy.Runtime) {
	if err := os.MkdirAll(filepath.Join(r.DebugStore(), "node"), os.ModePerm); err != nil {
		log.Errorf("unable to create directory to write node data to: %v", err)
		return
	}
	registerCommand(r, "ps", ps)
}

func registerCommand(r *envoy.Runtime, cmd string, execFunc func(binary.Runner) error) {
	_, err := exec.LookPath(cmd)
	if err != nil {
		log.Errorf("%v is not available, unable to collect data pre-termination", cmd)
		return
	}
	r.RegisterPreTermination(execFunc)
}

func ps(r binary.Runner) error {
	f, err := os.Create(filepath.Join(r.DebugStore(), "node/ps.txt"))
	if err != nil {
		return fmt.Errorf("unable to create file to write ps output to: %v", err)
	}
	defer func() { _ = f.Close() }()
	// #nosec -> all command parameters are hardcoded so we're safe!
	cmd := exec.Command("ps", "-eo", "user,stat,rss,vsz,minflt,majflt,pcpu,pmem,args")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error running ps: %v", err)
	}
	_, err = f.Write(out)
	return err
}
