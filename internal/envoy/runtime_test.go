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

package envoy

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/moreos"
)

func TestEnsureAdminAddressPath(t *testing.T) {
	runDir := t.TempDir()

	runAdminAddressPath := filepath.Join(runDir, "admin-address.txt")
	tests := []struct {
		name                 string
		args                 []string
		wantAdminAddressPath string
		wantArgs             []string
	}{
		{
			name:                 "no args",
			args:                 []string{"envoy"},
			wantAdminAddressPath: runAdminAddressPath,
			wantArgs:             []string{"envoy", "--admin-address-path", runAdminAddressPath},
		},
		{
			name:                 "args",
			args:                 []string{"envoy", "-c", "/tmp/google_com_proxy.v2.yaml"},
			wantAdminAddressPath: runAdminAddressPath,
			wantArgs:             []string{"envoy", "-c", "/tmp/google_com_proxy.v2.yaml", "--admin-address-path", runAdminAddressPath},
		},
		{
			name:                 "already",
			args:                 []string{"envoy", "--admin-address-path", "/tmp/admin.txt", "-c", "/tmp/google_com_proxy.v2.yaml"},
			wantAdminAddressPath: "/tmp/admin.txt",
			wantArgs:             []string{"envoy", "--admin-address-path", "/tmp/admin.txt", "-c", "/tmp/google_com_proxy.v2.yaml"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r := NewRuntime(&globals.RunOpts{RunDir: runDir})
			r.cmd = exec.Command(tt.args[0], tt.args[1:]...)

			err := r.ensureAdminAddressPath()
			require.NoError(t, err)
			require.Equal(t, tt.wantAdminAddressPath, r.adminAddressPath)
			require.Equal(t, tt.wantArgs, r.cmd.Args)
		})
	}
}

func TestEnsureAdminAddressPath_ValidateExisting(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedErr string
	}{
		{
			name:        "value empty",
			args:        []string{"envoy", "--admin-address-path", "", "-c", "/tmp/google_com_proxy.v2.yaml"},
			expectedErr: `missing value to argument "--admin-address-path"`,
		},
		{
			name:        "value missing",
			args:        []string{"envoy", "-c", "/tmp/google_com_proxy.v2.yaml", "--admin-address-path"},
			expectedErr: `missing value to argument "--admin-address-path"`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r := NewRuntime(&globals.RunOpts{})
			r.cmd = exec.Command(tt.args[0], tt.args[1:]...)

			err := r.ensureAdminAddressPath()
			require.Equal(t, tt.args, r.cmd.Args)
			require.Empty(t, r.adminAddressPath)
			require.EqualError(t, err, tt.expectedErr)
		})
	}
}

func TestPidFilePath(t *testing.T) {
	r := NewRuntime(&globals.RunOpts{RunDir: "run"})
	require.Equal(t, filepath.Join("run", "envoy.pid"), r.pidPath)
}

func TestString(t *testing.T) {
	cmdExited := NewRuntime(&globals.RunOpts{})
	cmdExited.cmd = exec.Command("echo")
	require.NoError(t, cmdExited.cmd.Run())

	cmdFailed := NewRuntime(&globals.RunOpts{})
	cmdFailed.cmd = exec.Command("cat"+moreos.Exe, "icecream")
	require.Error(t, cmdFailed.cmd.Run())

	// Fork a process that hangs
	cmdRunning := NewRuntime(&globals.RunOpts{})
	cmdRunning.cmd = exec.Command("cat" + moreos.Exe)
	cmdRunning.cmd.SysProcAttr = moreos.ProcessGroupAttr()
	require.NoError(t, cmdRunning.cmd.Start())
	defer cmdRunning.cmd.Process.Kill() //nolint

	tmpDir := t.TempDir()
	files := NewRuntime(&globals.RunOpts{})

	stdoutLog, err := os.OpenFile(filepath.Join(tmpDir, "stdout.log"), os.O_CREATE|os.O_WRONLY, 0600)
	require.NoError(t, err)
	defer stdoutLog.Close()
	files.OutFile = stdoutLog
	stdoutLog.Write([]byte("foo")) //nolint
	stdoutLog.Sync()               //nolint

	stderrLog, err := os.OpenFile(filepath.Join(tmpDir, "stderr.log"), os.O_CREATE|os.O_WRONLY, 0600)
	require.NoError(t, err)
	defer stderrLog.Close()
	files.ErrFile = stderrLog
	stderrLog.Write([]byte("bar")) //nolint
	stderrLog.Sync()               //nolint

	tests := []struct {
		name     string
		runtime  *Runtime
		expected string
	}{
		{
			name:     "command exited",
			runtime:  cmdExited,
			expected: "{stdout: , stderr: , exitStatus: 0}",
		},
		{
			name:     "command failed",
			runtime:  cmdFailed,
			expected: "{stdout: , stderr: , exitStatus: 1}",
		},
		{
			name:     "command running",
			runtime:  cmdRunning,
			expected: "{stdout: , stderr: , exitStatus: -1}",
		},
		{
			name:     "console files exist",
			runtime:  files,
			expected: "{stdout: foo, stderr: bar, exitStatus: -1}",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.runtime.String())
		})
	}
}
