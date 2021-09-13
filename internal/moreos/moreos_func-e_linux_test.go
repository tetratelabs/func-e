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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/test/fakebinary"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestProcessGroupAttr_Kill(t *testing.T) {
	tempDir := t.TempDir()

	// Build a fake envoy and pass the ENV hint so that fake func-e uses it
	fakeEnvoy := filepath.Join(tempDir, "envoy"+Exe)
	fakebinary.RequireFakeEnvoy(t, fakeEnvoy)
	t.Setenv("ENVOY_PATH", fakeEnvoy)

	fakeFuncE := filepath.Join(tempDir, "func-e"+Exe)
	requireFakeFuncE(t, fakeFuncE)

	stderr := new(bytes.Buffer)
	stdout := new(bytes.Buffer)

	// With an arg so fakeFuncE runs fakeEnvoy as its child and doesn't exit.
	arg := string(version.LastKnownEnvoy)
	cmd := exec.Command(fakeFuncE, "run", arg, "-c")
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	require.NoError(t, cmd.Start())

	// Block until we reach an expected line or timeout.
	reader := bufio.NewReader(stderr)
	waitFor := "initializing epoch 0"
	if !assert.Eventually(t, func() bool {
		b, e := reader.Peek(512)
		return e != nil && strings.Contains(string(b), waitFor)
	}, 5*time.Second, 100*time.Millisecond) {
		require.FailNowf(t, "timeout waiting for stderr to contain %q", waitFor)
	}

	require.Equal(t, Sprintf("starting: %s %s -c\n", fakeEnvoy, arg), stdout.String())

	fakeFuncEProcess, err := process.NewProcess(int32(cmd.Process.Pid))
	require.NoError(t, err)

	// Get all fake func-e children processes.
	children, err := fakeFuncEProcess.Children()
	require.NoError(t, err)
	require.Equal(t, len(children), 1) // Should have only one child process i.e. the fake envoy process.
	fakeEnvoyProcess := &os.Process{Pid: int(children[0].Pid)}
	requireFindProcessNoError(t, fakeEnvoyProcess)

	// Kill the fake func-e process.
	// This works only for linux, sending kill -9 on darwin will not kill the process, we need to kill
	// via pgid or kill the child first.
	require.NoError(t, cmd.Process.Kill())

	// The child process is expected to receive ENVOY_SIGTERM.
	require.Contains(t, "caught ENVOY_SIGTERM\nexiting\n", stderr.String())

	// Wait for the process to die; this could error due to the kill signal.
	cmd.Wait() //nolint

	// Wait and check if fake func-e and envoy processes are killed.
	requireFindProcessError(t, cmd.Process, process.ErrorProcessNotRunning)
	requireFindProcessError(t, fakeEnvoyProcess, process.ErrorProcessNotRunning)

	// Ensure both processes are killed.
	require.NoError(t, EnsureProcessDone(cmd.Process))
	require.NoError(t, EnsureProcessDone(fakeEnvoyProcess))
}

func requireFindProcessError(t *testing.T, proc *os.Process, expectedErr error) {
	// Wait until the operating system removes or adds the scheduled process.
	if !assert.Eventually(t, func() bool {
		_, err := process.NewProcess(int32(proc.Pid)) // because os.FindProcess is no-op in Linux!
		return err == expectedErr
	}, 100*time.Millisecond, 5*time.Millisecond) {
		if expectedErr == nil {
			require.FailNow(t, "timeout waiting for finding process with no error")
		}
		require.FailNow(t, "timeout waiting for expected error %v", expectedErr)
	}
}

func requireFindProcessNoError(t *testing.T, proc *os.Process) {
	requireFindProcessError(t, proc, nil) // expect to find the process with no error.
}
