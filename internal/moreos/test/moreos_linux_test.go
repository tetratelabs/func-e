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

package test

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/test"
	"github.com/tetratelabs/func-e/internal/version"
)

// TestProcessGroupAttr_Kill sends SIGKILL to running fake func-e that spawns a fake envoy process.
func TestProcessGroupAttr_Kill(t *testing.T) {
	tempDir := t.TempDir()
	fakeFuncE := filepath.Join(tempDir, "func-e")
	fakeEnvoyDir := filepath.Join(tempDir, "versions", string(version.LastKnownEnvoy))
	fakeEnvoy := filepath.Join(fakeEnvoyDir, "envoy")
	require.NoError(t, os.MkdirAll(fakeEnvoyDir, 0755))

	test.RequireFakeFuncE(t, fakeFuncE+moreos.Exe)
	test.RequireFakeEnvoy(t, fakeEnvoy+moreos.Exe)

	cmd := exec.CommandContext(context.Background(), fakeFuncE)
	cmd.Env = []string{"FUNC_E_HOME=" + tempDir}
	stderr := new(bytes.Buffer)
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

	fakeFuncEProcess, err := process.NewProcess(int32(cmd.Process.Pid))
	require.NoError(t, err)

	// Get all fake func-e children processes.
	children, err := fakeFuncEProcess.Children()
	require.NoError(t, err)
	require.Equal(t, len(children), 1) // Should have only one child process i.e. the fake envoy process.

	// Kill the fake func-e process.
	// This works only for linux, sending kill -9 on darwin will not kill the process, we need to kill
	// via pgid or kill the child first.
	require.NoError(t, cmd.Process.Kill())
	// Wait for the process to die; this could error due to the kill signal.
	cmd.Wait() //nolint

	require.NoError(t, moreos.EnsureProcessDone(cmd.Process))
	require.NoError(t, moreos.EnsureProcessDone(&os.Process{Pid: int(children[0].Pid)}))
}
