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
	"fmt"
	"os"
	"path/filepath"
)

// extensionDir represents a directory with an extension created by getenvoy toolkit.
type extensionDir string

func (d extensionDir) GetRootDir() string {
	return string(d)
}

func (d extensionDir) GetMetaDir() string {
	return filepath.Join(d.GetRootDir(), extensionMetaDir)
}

func (d extensionDir) Rel(path string) string {
	return filepath.Join(extensionMetaDir, path)
}

func (d extensionDir) Abs(path string) string {
	return filepath.Join(d.GetMetaDir(), path)
}

func (d extensionDir) HasFile(path string) (bool, error) {
	return d.hasFile(path, isRegularFile)
}

func (d extensionDir) ReadFile(path string) ([]byte, error) {
	path = d.Abs(path)
	return os.ReadFile(filepath.Clean(path))
}

func (d extensionDir) WriteFile(path string, data []byte) error {
	path = d.Abs(path)
	dir := filepath.Dir(path)
	// This is required until we hunt down the directory structure we ultimately need and create that once instead of
	// once per file as is happening here. pkg/extension needs simplifying, and doing so may obviate the hunt.
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", dir, err)
	}
	return os.WriteFile(path, data, 0600)
}

func (d extensionDir) HasDir(path string) (bool, error) {
	return d.hasFile(path, isDir)
}

func (d extensionDir) ListDirs(path string) ([]string, error) {
	path = d.Abs(path)
	infos, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	dirNames := make([]string, 0)
	for _, info := range infos {
		if info.IsDir() {
			dirNames = append(dirNames, info.Name())
		}
	}
	return dirNames, nil
}

func (d extensionDir) ListFiles(path string) ([]string, error) {
	root := d.Abs(path)
	fileNames := make([]string, 0)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		fileNames = append(fileNames, relPath)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return fileNames, nil
}

func (d extensionDir) RemoveAll(path string) error {
	path = d.Abs(path)
	return os.RemoveAll(path)
}

func (d extensionDir) hasFile(path string, test func(string, os.FileInfo) error) (bool, error) {
	path = d.Abs(path)
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if err := test(path, info); err != nil {
		return false, err
	}
	return true, nil
}

func isRegularFile(path string, info os.FileInfo) error {
	if !info.Mode().IsRegular() {
		return fmt.Errorf("unexpected file type: expected a regular file at a given path: %s", path)
	}
	return nil
}

func isDir(path string, info os.FileInfo) error {
	if !info.Mode().IsDir() {
		return fmt.Errorf("unexpected file type: expected a directory at a given path: %s", path)
	}
	return nil
}
