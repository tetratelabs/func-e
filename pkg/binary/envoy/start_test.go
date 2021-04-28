// Copyright 2019 Tetrate
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
	"errors"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/globals"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

func TestHandlePreStartReturnsOnFirstError(t *testing.T) {
	r := NewRuntime(&globals.RunOpts{})
	err1, err2 := errors.New("1"), errors.New("2")
	r.RegisterPreStart(func() error {
		return err1
	})
	r.RegisterPreStart(func() error {
		return err2
	})

	actualErr := r.handlePreStart()
	require.Equal(t, err1, actualErr)
}

func TestHandlePreStartReturnsError(t *testing.T) {
	r := NewRuntime(&globals.RunOpts{})
	first := false
	err := errors.New("1")
	r.RegisterPreStart(func() error {
		first = true
		return nil
	})
	r.RegisterPreStart(func() error {
		return err
	})

	actualErr := r.handlePreStart()
	require.Equal(t, true, first)
	require.Equal(t, err, actualErr)
}

func TestHandlePreStartEnsuresAdminAddressPath(t *testing.T) {
	tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
	defer removeTempDir()

	r := NewRuntime(&globals.RunOpts{WorkingDir: tempDir})
	r.cmd = exec.Command("envoy")

	// Verify the admin address path set (same assertion as ensureAdminAddressPath)
	actualErr := r.handlePreStart()
	require.NoError(t, actualErr)
	require.Equal(t, []string{"envoy", "--admin-address-path", "admin-address.txt"}, r.cmd.Args)
}

func TestHandlePreStartEnsuresAdminAddressPathLast(t *testing.T) {
	r := NewRuntime(&globals.RunOpts{})
	r.cmd = exec.Command("envoy")

	r.RegisterPreStart(func() error {
		r.AppendArgs([]string{"--admin-address-path", "/tmp/admin.txt"})
		return nil
	})

	actualErr := r.handlePreStart()
	require.NoError(t, actualErr)
	require.Equal(t, "/tmp/admin.txt", r.adminAddressPath)
	require.Equal(t, []string{"envoy", "--admin-address-path", "/tmp/admin.txt"}, r.cmd.Args)
}

func TestEnsureAdminAddressPath(t *testing.T) {
	workingDir, removeWorkingDir := morerequire.RequireNewTempDir(t)
	defer removeWorkingDir()

	tests := []struct {
		name                 string
		args                 []string
		wantAdminAddressPath string
		wantArgs             []string
	}{
		{
			name:                 "no args",
			args:                 []string{"envoy"},
			wantAdminAddressPath: filepath.Join(workingDir, "admin-address.txt"),
			wantArgs:             []string{"envoy", "--admin-address-path", "admin-address.txt"},
		},
		{
			name:                 "args",
			args:                 []string{"envoy", "-c", "/tmp/google_com_proxy.v2.yaml"},
			wantAdminAddressPath: filepath.Join(workingDir, "admin-address.txt"),
			wantArgs:             []string{"envoy", "-c", "/tmp/google_com_proxy.v2.yaml", "--admin-address-path", "admin-address.txt"},
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
			r := NewRuntime(&globals.RunOpts{WorkingDir: workingDir})
			r.cmd = exec.Command(tt.args[0], tt.args[1:]...)

			err := r.ensureAdminAddressPath()
			require.NoError(t, err)
			require.Equal(t, tt.wantAdminAddressPath, r.adminAddressPath)
			require.Equal(t, tt.wantArgs, r.cmd.Args)
		})
	}
}

func TestEnsureAdminAddressPathValidateExisting(t *testing.T) {
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
