// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/globals"
)

func TestEnsureAdminAddress(t *testing.T) {
	runDir := t.TempDir()

	runAdminAddressPath := filepath.Join(runDir, "admin-address.txt")
	tests := []struct {
		name                   string
		args                   []string
		expectAdminAddressPath string
		expectArgs             []string
		expectLogs             string
	}{
		{
			name:       "no args", // allows envoy to fail properly if no args are provided
			args:       []string{"envoy"},
			expectArgs: []string{"envoy"},
			expectLogs: "",
		},
		{
			name:       "empty config arg", // allows envoy to fail properly if no args are provided
			args:       []string{"-c", ""},
			expectArgs: []string{"-c", ""},
			expectLogs: "",
		},
		{
			name:                   "args",
			args:                   []string{"-c", "/tmp/google_com_proxy.v2.yaml"},
			expectAdminAddressPath: runAdminAddressPath,
			expectArgs:             []string{"-c", "/tmp/google_com_proxy.v2.yaml", "--admin-address-path", runAdminAddressPath},
			expectLogs:             "failed to find admin address: failed to read config file /tmp/google_com_proxy.v2.yaml: open /tmp/google_com_proxy.v2.yaml: no such file or directory\n",
		},
		{
			name:                   "already",
			args:                   []string{"--admin-address-path", "/tmp/admin.txt", "-c", "/tmp/google_com_proxy.v2.yaml"},
			expectAdminAddressPath: "/tmp/admin.txt",
			expectArgs:             []string{"--admin-address-path", "/tmp/admin.txt", "-c", "/tmp/google_com_proxy.v2.yaml"},
			expectLogs:             "failed to find admin address: failed to read config file /tmp/google_com_proxy.v2.yaml: open /tmp/google_com_proxy.v2.yaml: no such file or directory\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuf bytes.Buffer
			logf := func(format string, args ...any) {
				fmt.Fprintf(&logBuf, format+"\n", args...)
			}

			adminAddressPath, args, err := ensureAdminAddress(logf, runDir, tt.args)
			require.NoError(t, err)

			require.Equal(t, tt.expectAdminAddressPath, adminAddressPath)
			require.Equal(t, tt.expectArgs, args)
			require.Equal(t, tt.expectLogs, logBuf.String())
		})
	}
}

func TestEnsureAdminAddress_ValidateExisting(t *testing.T) {
	runDir := t.TempDir()
	tests := []struct {
		name        string
		args        []string
		expectedErr string
	}{
		{
			name:        "value empty",
			args:        []string{"--admin-address-path", "", "-c", "/tmp/google_com_proxy.v2.yaml"},
			expectedErr: `missing value to argument "--admin-address-path"`,
		},
		{
			name:        "value missing",
			args:        []string{"-c", "/tmp/google_com_proxy.v2.yaml", "--admin-address-path"},
			expectedErr: `missing value to argument "--admin-address-path"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adminAddressPath, args, err := ensureAdminAddress(t.Logf, runDir, tt.args)
			require.Equal(t, tt.args, args)
			require.Empty(t, adminAddressPath)
			require.EqualError(t, err, tt.expectedErr)
		})
	}
}

func TestString(t *testing.T) {
	cmdExited := NewRuntime(&globals.RunOpts{}, t.Logf)
	cmdExited.cmd = exec.Command("echo")
	require.NoError(t, cmdExited.cmd.Run())

	cmdFailed := NewRuntime(&globals.RunOpts{}, t.Logf)
	cmdFailed.cmd = exec.Command("cat", "icecream")
	require.Error(t, cmdFailed.cmd.Run())

	// Fork a process that hangs
	cmdRunning := NewRuntime(&globals.RunOpts{}, t.Logf)
	cmdRunning.cmd = exec.Command("cat")
	require.NoError(t, cmdRunning.cmd.Start())
	defer func() {
		if cmdRunning.cmd.Process != nil {
			_ = cmdRunning.cmd.Process.Kill()
		}
	}()

	tests := []struct {
		name     string
		runtime  *Runtime
		expected string
	}{
		{
			name:     "command exited",
			runtime:  cmdExited,
			expected: "{exitStatus: 0}",
		},
		{
			name:     "command failed",
			runtime:  cmdFailed,
			expected: "{exitStatus: 1}",
		},
		{
			name:     "command running",
			runtime:  cmdRunning,
			expected: "{exitStatus: -1}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.runtime.String())
		})
	}
}
