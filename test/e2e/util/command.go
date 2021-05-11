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
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	// GetEnvoyPath holds a path to a 'getenvoy' binary under test.
	GetEnvoyPath = "getenvoy"
)

// cmdBuilder represents a command builder.
type cmdBuilder struct {
	cmd *exec.Cmd
}

// GetEnvoy returns a new command builder.
func GetEnvoy(cmdline string) *cmdBuilder { //nolint:golint
	args, err := SplitCommandLine(cmdline)
	if err != nil {
		panic(err)
	}
	return &cmdBuilder{exec.Command(GetEnvoyPath, args...)} //nolint:gosec
}

func (b *cmdBuilder) WorkingDir(arg string) *cmdBuilder {
	b.cmd.Dir = arg
	return b
}

func (b *cmdBuilder) Arg(arg string) *cmdBuilder {
	return b.Args(arg)
}

func (b *cmdBuilder) Args(args ...string) *cmdBuilder {
	b.cmd.Args = append(b.cmd.Args, args...)
	return b
}

func (b *cmdBuilder) String() string {
	return fmt.Sprintf("%s: %s", b.cmd.Dir, strings.Join(b.cmd.Args, " "))
}

func (b *cmdBuilder) Exec() (string, string, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	b.cmd.Stdout = io.MultiWriter(os.Stdout, stdout) // we want to see full `getenvoy` output in the test log
	b.cmd.Stderr = io.MultiWriter(os.Stderr, stderr)
	err := b.cmd.Run()
	return stdout.String(), stderr.String(), err
}

func (b *cmdBuilder) Start(t *testing.T, terminateTimeout time.Duration) (io.Reader, io.Reader, func()) {
	stdout := newSyncBuffer()
	stderr := newSyncBuffer()
	b.cmd.Stdout = io.MultiWriter(os.Stdout, stdout) // we want to see full `getenvoy` output in the test log
	b.cmd.Stderr = io.MultiWriter(os.Stderr, stderr)
	err := b.cmd.Start()
	require.NoError(t, err, `error starting [%v]`, b)

	errc := make(chan error, 1)
	go func() {
		errc <- b.cmd.Wait()
	}()

	return stdout, stderr, func() {
		err := b.cmd.Process.Signal(syscall.SIGTERM)
		require.NoError(t, err, `error terminating [%v]`, b.cmd)

		select {
		case e := <-errc:
			require.NoError(t, e, `error running [%v]`, b.cmd)
		case <-time.After(terminateTimeout):
			t.Fatal(fmt.Sprintf("getenvoy command didn't exit gracefully within %s", terminateTimeout))
		}
	}
}
