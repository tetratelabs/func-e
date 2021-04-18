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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/binary"
)

const debugDir = "~/.getenvoy/debug/1"

func TestHandlePreStartReturnsOnFirstError(t *testing.T) {
	r, _ := NewRuntime()
	err1, err2 := errors.New("1"), errors.New("2")
	r.RegisterPreStart(func(runner binary.Runner) error {
		return err1
	})
	r.RegisterPreStart(func(runner binary.Runner) error {
		return err2
	})

	actualErr := r.(*Runtime).handlePreStart()
	require.Equal(t, err1, actualErr)
}

func TestHandlePreStartReturnsError(t *testing.T) {
	r, _ := NewRuntime()
	first := false
	err := errors.New("1")
	r.RegisterPreStart(func(runner binary.Runner) error {
		first = true
		return nil
	})
	r.RegisterPreStart(func(runner binary.Runner) error {
		return err
	})

	actualErr := r.(*Runtime).handlePreStart()
	require.Equal(t, true, first)
	require.Equal(t, err, actualErr)
}

func TestHandlePreStartEnsuresAdminAddressPath(t *testing.T) {
	r, _ := NewRuntime()
	e := r.(*Runtime)
	e.cmd = exec.Command("envoy")

	// Verify the admin address path set (same assertion as ensureAdminAddressPath)
	actualErr := r.(*Runtime).handlePreStart()
	require.NoError(t, actualErr)
	require.Equal(t, []string{"envoy", "--admin-address-path", e.adminAddressPath}, e.cmd.Args)
}

func TestHandlePreStartEnsuresAdminAddressPathLast(t *testing.T) {
	r, _ := NewRuntime()
	e := r.(*Runtime)
	e.cmd = exec.Command("envoy")

	r.RegisterPreStart(func(runner binary.Runner) error {
		runner.AppendArgs([]string{"--admin-address-path", "/tmp/admin.txt"})
		return nil
	})

	actualErr := r.(*Runtime).handlePreStart()
	require.NoError(t, actualErr)
	require.Equal(t, "/tmp/admin.txt", e.adminAddressPath)
	require.Equal(t, []string{"envoy", "--admin-address-path", "/tmp/admin.txt"}, e.cmd.Args)
}

func TestEnsureAdminAddressPath(t *testing.T) {
	tests := []struct {
		name                 string
		args                 []string
		wantAdminAddressPath string
		wantArgs             []string
	}{
		{
			name:                 "no args",
			args:                 []string{"envoy"},
			wantAdminAddressPath: "~/.getenvoy/debug/1/admin-address.txt",
			wantArgs:             []string{"envoy", "--admin-address-path", "~/.getenvoy/debug/1/admin-address.txt"},
		},
		{
			name:                 "args",
			args:                 []string{"envoy", "-c", "/tmp/google_com_proxy.v2.yaml"},
			wantAdminAddressPath: "~/.getenvoy/debug/1/admin-address.txt",
			wantArgs:             []string{"envoy", "-c", "/tmp/google_com_proxy.v2.yaml", "--admin-address-path", "~/.getenvoy/debug/1/admin-address.txt"},
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
			r, _ := NewRuntime()
			e := r.(*Runtime)
			e.cmd = exec.Command(tt.args[0], tt.args[1:]...)
			e.debugDir = debugDir

			err := e.ensureAdminAddressPath()
			require.NoError(t, err)
			require.Equal(t, tt.wantAdminAddressPath, e.adminAddressPath)
			require.Equal(t, tt.wantArgs, e.cmd.Args)
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
			r, _ := NewRuntime()
			e := r.(*Runtime)
			e.cmd = exec.Command(tt.args[0], tt.args[1:]...)
			e.debugDir = debugDir

			err := e.ensureAdminAddressPath()
			require.Equal(t, tt.args, e.cmd.Args)
			require.Empty(t, e.adminAddressPath)
			require.EqualError(t, err, tt.expectedErr)
		})
	}
}
