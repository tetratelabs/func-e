// Copyright 2020 Tetrate
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

package init

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

func TestExtensionDirValidatorReject(t *testing.T) {
	type testCase struct {
		name        string
		path        string
		expectedErr string
	}

	cwd, err := os.Getwd()
	require.NoError(t, err, "error getting current working directory")

	file := os.Args[0]
	pathUnderFile := filepath.Join(file, "subdir")
	tests := []testCase{
		{
			name:        "output path that is a file",
			path:        file,
			expectedErr: fmt.Sprintf(`extension directory is a file: %s`, file),
		},
		{
			name:        "output path under a file",
			path:        pathUnderFile,
			expectedErr: fmt.Sprintf(`stat %s: not a directory`, pathUnderFile),
		},
		{
			name:        "output path not empty",
			path:        cwd,
			expectedErr: fmt.Sprintf(`extension directory must be empty or new: %s`, cwd),
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			err = newParams().ExtensionDir.Validator(test.path)
			require.EqualError(t, err, test.expectedErr)
		})
	}
}

func TestExtensionDirValidatorAccept(t *testing.T) {
	tempDir, removeTempDir := RequireNewTempDir(t)
	defer removeTempDir()

	type testCase struct {
		name string
		path string
	}

	tests := []testCase{
		{
			name: "output path that is an empty directory",
			path: tempDir,
		},
		{
			name: "output path that doesn't exist",
			path: filepath.Join(tempDir, "subdir"),
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			err := newParams().ExtensionDir.Validator(test.path)
			require.NoError(t, err)
		})
	}
}
