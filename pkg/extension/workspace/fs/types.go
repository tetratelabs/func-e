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

package fs

import (
	"io/ioutil"
	"os"
	"path/filepath"

	osutil "github.com/tetratelabs/getenvoy/pkg/util/os"
)

// workspaceDir represents a directory with an extension created by getenvoy toolkit.
type workspaceDir string

func (d workspaceDir) GetRootDir() string {
	return string(d)
}

func (d workspaceDir) GetMetaDir() string {
	return filepath.Join(d.GetRootDir(), extensionMetaDir)
}

func (d workspaceDir) Rel(path string) string {
	return filepath.Join(extensionMetaDir, path)
}

func (d workspaceDir) Abs(path string) string {
	return filepath.Join(d.GetMetaDir(), path)
}

func (d workspaceDir) HasFile(path string, test func(os.FileInfo) error) (bool, error) {
	path = d.Abs(path)
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if err := test(info); err != nil {
		return false, err
	}
	return true, nil
}

func (d workspaceDir) ReadFile(path string) ([]byte, error) {
	path = d.Abs(path)
	return ioutil.ReadFile(filepath.Clean(path))
}

func (d workspaceDir) WriteFile(path string, data []byte) error {
	path = d.Abs(path)
	if err := osutil.EnsureDirExists(filepath.Dir(path)); err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}
