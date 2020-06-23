// Copyright 2020 Tetrate
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

package util

import (
	"bytes"
	"io"
	"os"
	"os/exec"

	argutil "github.com/tetratelabs/getenvoy/pkg/util/args"
)

var (
	// GetEnvoyBinaryPath holds a path to a 'getenvoy' binary under test.
	GetEnvoyBinaryPath = "getenvoy"
)

// cmdBuilder represents a command builder.
type cmdBuilder struct {
	cmd *exec.Cmd
}

// GetEnvoy returns a new command builder.
func GetEnvoy(cmdline string) *cmdBuilder { //nolint:golint
	args, err := argutil.SplitCommandLine(cmdline)
	if err != nil {
		panic(err)
	}
	return &cmdBuilder{exec.Command(GetEnvoyBinaryPath, args...)} //nolint:gosec
}

func (b *cmdBuilder) Arg(arg string) *cmdBuilder {
	return b.Args(arg)
}

func (b *cmdBuilder) Args(args ...string) *cmdBuilder {
	b.cmd.Args = append(b.cmd.Args, args...)
	return b
}

func (b *cmdBuilder) Exec() (string, string, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	b.cmd.Stdout = io.MultiWriter(os.Stdout, stdout) // we want to see full `getenvoy` output in the test log
	b.cmd.Stderr = io.MultiWriter(os.Stderr, stderr)
	err := b.cmd.Run()
	return stdout.String(), stderr.String(), err
}
