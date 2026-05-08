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
	adminYaml := "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 9901}}}"
	noAdminYaml := "static_resources: {}"
	tests := []struct {
		name                     string
		args                     []string
		expectedAdminAddressPath string
		expectedArgs             []string
		expectedLogs             string
	}{
		{
			name:         "leaves missing config unchanged for Envoy to report",
			args:         []string{"envoy"},
			expectedArgs: []string{"envoy"},
			expectedLogs: "",
		},
		{
			name:         "leaves empty config value unchanged for Envoy to report",
			args:         []string{"-c", ""},
			expectedArgs: []string{"-c", ""},
			expectedLogs: "",
		},
		{
			name:                     "adds admin path when bootstrap file cannot be inspected",
			args:                     []string{"-c", "/tmp/google_com_proxy.v2.yaml"},
			expectedAdminAddressPath: runAdminAddressPath,
			expectedArgs:             []string{"-c", "/tmp/google_com_proxy.v2.yaml", "--admin-address-path", runAdminAddressPath},
			expectedLogs:             "failed to find admin address: failed to read config file /tmp/google_com_proxy.v2.yaml: open /tmp/google_com_proxy.v2.yaml: no such file or directory\n",
		},
		{
			name:                     "adds admin path for equals-form config path",
			args:                     []string{"--config-path=/tmp/google_com_proxy.v2.yaml"},
			expectedAdminAddressPath: runAdminAddressPath,
			expectedArgs:             []string{"--config-path=/tmp/google_com_proxy.v2.yaml", "--admin-address-path", runAdminAddressPath},
			expectedLogs:             "failed to find admin address: failed to read config file /tmp/google_com_proxy.v2.yaml: open /tmp/google_com_proxy.v2.yaml: no such file or directory\n",
		},
		{
			name:                     "adds admin path when config already has admin server",
			args:                     []string{"--config-yaml=" + adminYaml},
			expectedAdminAddressPath: runAdminAddressPath,
			expectedArgs:             []string{"--config-yaml=" + adminYaml, "--admin-address-path", runAdminAddressPath},
			expectedLogs:             "",
		},
		{
			name:         "does not backfill from config hidden behind ignore-rest",
			args:         []string{"--", "--config-yaml", adminYaml},
			expectedArgs: []string{"--", "--config-yaml", adminYaml},
		},
		{
			name:                     "does not trust admin path hidden behind ignore-rest",
			args:                     []string{"--config-yaml=" + adminYaml, "--", "--admin-address-path", "/tmp/ignored-admin.txt"},
			expectedAdminAddressPath: runAdminAddressPath,
			expectedArgs:             []string{"--config-yaml=" + adminYaml, "--admin-address-path", runAdminAddressPath, "--", "--admin-address-path", "/tmp/ignored-admin.txt"},
		},
		{
			name:                     "inserts generated admin args before ignore-rest",
			args:                     []string{"--config-yaml", noAdminYaml, "--", "--log-level", "debug"},
			expectedAdminAddressPath: runAdminAddressPath,
			expectedArgs:             []string{"--config-yaml", noAdminYaml, "--config-yaml", adminEphemeralConfig, "--admin-address-path", runAdminAddressPath, "--", "--log-level", "debug"},
			expectedLogs:             "configuring ephemeral admin server\n",
		},
		{
			name:                     "keeps caller-provided admin path value",
			args:                     []string{"--admin-address-path", "/tmp/admin.txt", "-c", "/tmp/google_com_proxy.v2.yaml"},
			expectedAdminAddressPath: "/tmp/admin.txt",
			expectedArgs:             []string{"--admin-address-path", "/tmp/admin.txt", "-c", "/tmp/google_com_proxy.v2.yaml"},
			expectedLogs:             "failed to find admin address: failed to read config file /tmp/google_com_proxy.v2.yaml: open /tmp/google_com_proxy.v2.yaml: no such file or directory\n",
		},
		{
			name:                     "keeps caller-provided admin path equals form",
			args:                     []string{"--admin-address-path=/tmp/admin.txt", "--config-path=/tmp/google_com_proxy.v2.yaml"},
			expectedAdminAddressPath: "/tmp/admin.txt",
			expectedArgs:             []string{"--admin-address-path=/tmp/admin.txt", "--config-path=/tmp/google_com_proxy.v2.yaml"},
			expectedLogs:             "failed to find admin address: failed to read config file /tmp/google_com_proxy.v2.yaml: open /tmp/google_com_proxy.v2.yaml: no such file or directory\n",
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

			require.Equal(t, tt.expectedAdminAddressPath, adminAddressPath)
			require.Equal(t, tt.expectedArgs, args)
			require.Equal(t, tt.expectedLogs, logBuf.String())
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
		{
			name:        "equals value empty",
			args:        []string{"--admin-address-path=", "-c", "/tmp/google_com_proxy.v2.yaml"},
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
			cmdRunning.cmd.Process.Kill()
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
