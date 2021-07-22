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
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/test/morerequire"
)

func TestIsExecutable(t *testing.T) {
	tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
	defer removeTempDir()

	bin := filepath.Join(tempDir, "envoy"+Exe)
	require.NoError(t, os.WriteFile(bin, []byte{}, 0700))

	f, err := os.Stat(bin)
	require.NoError(t, err)

	require.True(t, isExecutable(f))
}

func TestIsExecutable_Not(t *testing.T) {
	tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
	defer removeTempDir()

	bin := filepath.Join(tempDir, "foo.txt")
	require.NoError(t, os.WriteFile(bin, []byte{}, 0600))

	f, err := os.Stat(bin)
	require.NoError(t, err)

	require.False(t, isExecutable(f))
}

func TestSprintf(t *testing.T) {
	template := "%s\n\n%s\n"

	expected := "foo\n\nbar\n"
	if runtime.GOOS == OSWindows {
		expected = "foo\r\n\r\nbar\r\n"
	}

	require.Equal(t, expected, Sprintf(template, "foo", "bar"))
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
	cmd.Stdout = stdout
	require.NoError(t, cmd.Run())
	require.Equal(t, Sprintf("cats\n"), stdout.String())
}

func TestProcessGroupAttr_Interrupt(t *testing.T) {
	// Fork a process that hangs
	cmd := exec.Command("cat" + Exe)
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
	cmd := exec.Command("cat" + Exe)
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

func findProcess(proc *os.Process) error {
	_, err := process.NewProcess(int32(proc.Pid)) // because os.FindProcess is no-op in Linux!
	return err
}
