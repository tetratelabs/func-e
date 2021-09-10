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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/stretchr/testify/require"
)

// TestErrorWithWindowsPathSeparator makes sure errors don't accidentally escape the windows path separator.
// this is extracted so that maintainers can make sure it works without using windows.
func TestErrorWithWindowsPathSeparator(t *testing.T) {
	err := errors.New("/foo/bar is bad")
	require.EqualError(t, errorWithWindowsPathSeparator(err), `\foo\bar is bad`)

	wrapped := errors.New("bad day")
	err = fmt.Errorf("/foo/bar is unhappy: %w", wrapped)
	wErr := errorWithWindowsPathSeparator(err)
	require.EqualError(t, wErr, `\foo\bar is unhappy: bad day`)
	require.Same(t, wrapped, errors.Unwrap(wErr))
}

func TestIsExecutable(t *testing.T) {
	tempDir := t.TempDir()

	bin := filepath.Join(tempDir, "envoy"+Exe)
	require.NoError(t, os.WriteFile(bin, []byte{}, 0700))

	f, err := os.Stat(bin)
	require.NoError(t, err)

	require.True(t, isExecutable(f))
}

func TestIsExecutable_Not(t *testing.T) {
	tempDir := t.TempDir()

	bin := filepath.Join(tempDir, "foo.txt")
	require.NoError(t, os.WriteFile(bin, []byte{}, 0600))

	f, err := os.Stat(bin)
	require.NoError(t, err)

	require.False(t, isExecutable(f))
}

func TestReplacePathSeparator(t *testing.T) {
	path := "/foo/bar"

	expected := path
	if runtime.GOOS == OSWindows {
		expected = "\\foo\\bar"
	}

	require.Equal(t, expected, ReplacePathSeparator(path))
}

func TestSprintf(t *testing.T) {
	template := "%s\n\n%s\n"

	expected := "foo\n\nbar\n"
	if runtime.GOOS == OSWindows {
		expected = "foo\r\n\r\nbar\r\n"
	}

	require.Equal(t, expected, Sprintf(template, "foo", "bar"))

	// ensure idempotent
	require.Equal(t, expected, expected)
}

func TestFprintf(t *testing.T) {
	template := "%s\n\n%s\n"
	stdout := new(bytes.Buffer)
	count, err := Fprintf(stdout, template, "foo", "bar")
	require.NoError(t, err)

	expected := "foo\n\nbar\n"
	if runtime.GOOS == OSWindows {
		expected = "foo\r\n\r\nbar\r\n"
	}

	require.Equal(t, expected, stdout.String())
	require.Equal(t, len(expected), count)
}

// TestSprintf_IdiomaticPerOS is here to ensure that the EOL translation makes sense. For example, in UNIX, we expect
// \n and windows \r\n. This uses a real command to prove the point.
func TestSprintf_IdiomaticPerOS(t *testing.T) {
	stdout := new(bytes.Buffer)
	cmd := exec.Command("echo", "cats")
	if runtime.GOOS == OSWindows {
		cmd = exec.Command("cmd", "/c", "echo", "cats")
	}
	cmd.Stdout = stdout
	require.NoError(t, cmd.Run())
	require.Equal(t, Sprintf("cats\n"), stdout.String())
}

func TestProcessGroupAttr_Interrupt(t *testing.T) {
	// Fork a process that hangs
	cmd := exec.Command("sleep"+Exe, "1000")
	cmd.SysProcAttr = ProcessGroupAttr()
	require.NoError(t, cmd.Start())

	// Verify the process exists
	require.NoError(t, findProcess(cmd.Process))

	// Interrupt it
	require.NoError(t, Interrupt(cmd.Process))

	// Wait for the process to die; this could error due to the interrupt signal
	cmd.Wait() //nolint
	require.Error(t, findProcess(cmd.Process))

	// Ensure interrupting it again doesn't error
	require.NoError(t, Interrupt(cmd.Process))
}

func Test_EnsureProcessDone(t *testing.T) {
	// Fork a process that hangs
	cmd := exec.Command("sleep"+Exe, "1000")
	cmd.SysProcAttr = ProcessGroupAttr()
	require.NoError(t, cmd.Start())

	// Kill it
	require.NoError(t, EnsureProcessDone(cmd.Process))

	// Wait for the process to die; this could error due to the kill signal
	cmd.Wait() //nolint
	require.Error(t, findProcess(cmd.Process))

	// Ensure killing it again doesn't error
	require.NoError(t, EnsureProcessDone(cmd.Process))
}

func Test_EnsureChildProcessDone(t *testing.T) {
	// There are 3 processes spawned by this test target, with the following relationship:
	//
	// runner (go run testdata/main/main.go)
	// └── main (the executable, compiled from testdata/main/main.go. This is a fake func-e)
	//     └── child (sleep, this is a fake envoy)
	cmd := exec.Command("go"+Exe, "run", filepath.Join("testdata", "main", "main.go"))
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)
	cmd.Start()

	buf := bufio.NewReader(stdout)
	for {
		line, _, readErr := buf.ReadLine()
		require.NoError(t, readErr)
		// As soon as we detect output from the main program, we know the child process in main program
		// is started. See: testdata/main/main.go#34.
		if string(line) == "ok" {
			break
		}
	}

	// This is the "go run" process.
	runner, err := process.NewProcess(int32(cmd.Process.Pid))
	require.NoError(t, err)

	parents, err := runner.Children()
	require.NoError(t, err)
	require.Equal(t, len(parents), 1)

	// This is the main process executed by "go run".
	main, err := process.NewProcess(parents[0].Pid)
	require.NoError(t, err)

	children, err := main.Children()
	require.NoError(t, err)
	require.Equal(t, len(children), 1)

	// Kill the main process.
	require.NoError(t, EnsureProcessDone(&os.Process{Pid: int(main.Pid)}))

	// Wait and check the status of child and main processes.
	// The following check is here to make sure the child and main processes are not killed because
	// of the runner is died.
	time.Sleep(time.Millisecond)
	child := children[0]
	require.Error(t, findProcess(&os.Process{Pid: int(child.Pid)}))
	require.Error(t, findProcess(&os.Process{Pid: int(main.Pid)}))

	// The runner needs to be killed too.
	require.NoError(t, EnsureProcessDone(&os.Process{Pid: int(runner.Pid)}))
	cmd.Wait() //nolint

	require.Error(t, findProcess(&os.Process{Pid: int(runner.Pid)}))

	// Ensure killing it again doesn't error.
	require.NoError(t, EnsureProcessDone(&os.Process{Pid: int(main.Pid)}))
}

func findProcess(proc *os.Process) error {
	_, err := process.NewProcess(int32(proc.Pid)) // because os.FindProcess is no-op in Linux!
	return err
}
