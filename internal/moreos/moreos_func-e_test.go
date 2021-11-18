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
	"embed"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/test/fakebinary"
)

var (
	// fakeFuncESrc is a test source file used to simulate how func-e manages its child process
	//go:embed testdata/fake_func-e.go
	fakeFuncESrc []byte // Embedding the fakeFuncESrc is easier than file I/O and ensures it doesn't skew coverage

	// Include the source imported by fakeFuncESrc directly and indirectly
	// TODO: replace wildcard with {{goos}} after https://github.com/golang/go/issues/48348.
	//go:embed moreos.go
	//go:embed proc_*.go
	moreosSrcDir embed.FS
)

// Test_CallSignals tests sending signals to fake func-e.
func Test_CallSignals(t *testing.T) {
	type testCase struct {
		name           string
		signal         func(*os.Process) error
		skip           bool
		waitForExiting bool
	}

	tests := []testCase{
		{
			name:           "Interrupt",
			signal:         Interrupt,
			waitForExiting: true,
		},
		{
			name:           "SIGTERM",
			signal:         func(proc *os.Process) error { return proc.Signal(syscall.SIGTERM) },
			waitForExiting: true,
			// On Windows, os.Process.Signal is not implemented; it will return an error instead of sending
			// a signal.
			skip: runtime.GOOS == OSWindows,
		},
		{
			name: "Kill",
			// On Linux, we propagate SIGKILL to the child process as the configured SysProcAttr.Pdeathsig
			// in proc_linux.go.
			signal:         func(proc *os.Process) error { return proc.Kill() },
			waitForExiting: false, // since the process is killed, it is immediately exit.
			// This works only for Linux, sending kill -9 on Darwin will not kill the process, we need to
			// kill via pgid or kill the child first.
			skip: runtime.GOOS != OSLinux,
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why.

		t.Run(tc.name, func(t *testing.T) {
			if tc.skip {
				t.Skip()
			}

			tempDir := t.TempDir()

			// Build a fake envoy and pass the path via via ENVOY_PATH so that fake func-e uses it.
			fakeEnvoy := filepath.Join(tempDir, "envoy"+Exe)
			fakebinary.RequireFakeEnvoy(t, fakeEnvoy)
			t.Setenv("ENVOY_PATH", fakeEnvoy)

			fakeFuncE := filepath.Join(tempDir, "func-e"+Exe)
			requireFakeFuncE(t, fakeFuncE)

			stdout := new(bytes.Buffer)

			// With an arg so fakeFuncE runs fakeEnvoy as its child and doesn't exit.
			arg := "1.1.1" // version.LastKnownEnvoy would introduce an import cycle
			cmd := exec.Command(fakeFuncE, "run", arg, "-c")
			cmd.SysProcAttr = ProcessGroupAttr() // Make sure we have a new process group.
			cmd.Stdout = stdout

			stderr, err := cmd.StderrPipe()
			require.NoError(t, err)
			stderrScanner := bufio.NewScanner(stderr)

			require.NoError(t, cmd.Start())

			// Block until we reach an expected line or timeout.
			requireScannedWaitFor(t, stderrScanner, "starting main dispatch loop")
			require.Equal(t, Sprintf("starting: %s %s -c\n", fakeEnvoy, arg), stdout.String())

			fakeFuncEProcess, err := process.NewProcess(int32(cmd.Process.Pid))
			require.NoError(t, err)

			// Get all fake func-e children processes.
			children, err := fakeFuncEProcess.Children()
			require.NoError(t, err)
			require.Equal(t, len(children), 1) // Should have only one child process i.e. the fake envoy process.
			fakeEnvoyProcess := &os.Process{Pid: int(children[0].Pid)}

			tc.signal(cmd.Process)
			if tc.waitForExiting {
				// When we decide to wait for the fake envoy for exiting, we wait for "exiting" message
				// from envoy (see: internal/test/fakebinary/testdata/fake_envoy.go#82) after receiving
				// interrupt from func-e (the fake func-e forwards fake envoy's stderr.
				// See: internal/moreos/testdata/fake_func-e.go#38).
				requireScannedWaitFor(t, stderrScanner, "exiting")
			}

			// Wait for the process to die; this could error due to the kill signal.
			err = cmd.Wait()
			if tc.waitForExiting {
				// When it is terminated using interrupt or SIGTERM, we expect cmd.Wait to be properly stopped.
				require.NoError(t, err)
			}

			// Wait and check if fake func-e and envoy processes are stopped.
			requireFindProcessError(t, cmd.Process, process.ErrorProcessNotRunning)
			requireFindProcessError(t, fakeEnvoyProcess, process.ErrorProcessNotRunning)

			// Ensure both processes are killed.
			require.NoError(t, EnsureProcessDone(cmd.Process))
			require.NoError(t, EnsureProcessDone(fakeEnvoyProcess))
		})
	}
}

// requireFakeFuncE builds a func-e binary only depends on fakeFuncESrc and the sources in this package.
// This is used to test integrated use of tools like ProcessGroupAttr without mixing unrelated concerns or dependencies.
func requireFakeFuncE(t *testing.T, path string) {
	workDir := t.TempDir()

	// Copy the sources needed for fake func-e, but nothing else
	moreosDir := filepath.Join(workDir, "internal", "moreos")
	require.NoError(t, os.MkdirAll(moreosDir, 0o700))
	moreosSrcs, err := moreosSrcDir.ReadDir(".")
	require.NoError(t, err)
	for _, src := range moreosSrcs {
		data, err := moreosSrcDir.ReadFile(src.Name())
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(moreosDir, src.Name()), data, 0o600))
	}

	fakeFuncEBin := fakebinary.RequireBuildFakeBinary(t, workDir, "func-e", fakeFuncESrc)
	require.NoError(t, os.WriteFile(path, fakeFuncEBin, 0o700)) //nolint:gosec
}

func requireFindProcessError(t *testing.T, proc *os.Process, expectedErr error) {
	// Wait until the operating system removes or adds the scheduled process.
	require.Eventually(t, func() bool {
		_, err := process.NewProcess(int32(proc.Pid)) // because os.FindProcess is no-op in Linux!
		return err == expectedErr
	}, 100*time.Millisecond, 5*time.Millisecond, "timeout waiting for expected error %v", expectedErr)
}

func requireScannedWaitFor(t *testing.T, s *bufio.Scanner, waitFor string) {
	require.Eventually(t, func() bool {
		for s.Scan() {
			if strings.HasPrefix(s.Text(), waitFor) {
				return true
			}
		}
		return false
	}, 5*time.Second, 100*time.Millisecond, "timeout waiting for scanner to find %q", waitFor)
}
