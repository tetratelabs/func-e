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

package binary

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/mitchellh/go-homedir"
)

// NewRuntime creates a new binary.Runtime with the local file storage set to the home directory
func NewRuntime(options ...func(*Runtime)) (*Runtime, error) {
	usrDir, err := homedir.Dir()
	local := filepath.Join(usrDir, ".getenvoy")
	runtime := &Runtime{
		local:          local,
		wg:             &sync.WaitGroup{},
		signals:        make(chan os.Signal),
		preStart:       make([]preStartFunc, 0),
		preTermination: make([]preTerminationFunc, 0),
	}
	for _, option := range options {
		option(runtime)
	}
	return runtime, err
}

// Runtime manages an Envoy lifecycle including fetching (if necessary) and running
type Runtime struct {
	local    string
	DebugDir string

	cmd *exec.Cmd
	wg  *sync.WaitGroup

	// This channel doesn't need to be on the struct, it's here to enable the testing of termination behavior.
	// Don't push any information onto this channel as it will likely have unintended consequences.
	signals chan os.Signal

	preStart       []preStartFunc
	preTermination []preTerminationFunc
}

type preStartFunc func(r *Runtime) error
type preTerminationFunc func(r *Runtime) error
