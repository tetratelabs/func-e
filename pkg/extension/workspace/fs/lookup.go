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
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	osutil "github.com/tetratelabs/getenvoy/pkg/util/os"
)

// CreateWorkspaceDir creates a new workspace directory at a given path.
func CreateWorkspaceDir(at string) (WorkspaceDir, error) {
	if err := osutil.EnsureDirExists(filepath.Join(at, extensionMetaDir)); err != nil {
		return nil, errors.Errorf("failed to create extension directory at: %s", at)
	}
	return GetWorkspaceDir(at)
}

// GetWorkspaceDir returns the workspace directory rooted at a given path
// or an error otherwise.
func GetWorkspaceDir(at string) (WorkspaceDir, error) {
	path, err := filepath.Abs(at)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to determine if there is an extension directory at: %s", at)
	}
	if IsWorkspaceDir(path) {
		return workspaceDir(path), nil
	}
	return nil, errors.Wrapf(err, "not an extension directory: %s", path)
}

// IsWorkspaceDir returns true if a given path corresponds to a directory
// with an extension created by getenvoy toolkit.
func IsWorkspaceDir(path string) bool {
	metaDir := filepath.Join(path, extensionMetaDir)
	info, err := os.Stat(metaDir)
	if err != nil {
		return false
	}
	return info.Mode().IsDir()
}

// FindWorkspaceDir attempts to find a directory with an extension
// created by getenvoy toolkit in the current working directory and its parents.
func FindWorkspaceDir() (WorkspaceDir, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "failed to determine current working dir")
	}
	return FindWorkspaceDirAt(cwd)
}

// FindWorkspaceDirAt attempts to find a directory with an extension
// created by getenvoy toolkit in a given directory and its parents.
func FindWorkspaceDirAt(path string) (WorkspaceDir, error) {
	root, err := findUp(path, IsWorkspaceDir)
	if err == os.ErrNotExist {
		return nil, errors.Errorf("there is no extension directory at or above: %s", path)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to determine if there is an extension directory at or above: %s", path)
	}
	return workspaceDir(root), nil
}

func findUp(path string, predicate func(string) bool) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	for dir, prev := path, ""; dir != prev; prev, dir = dir, filepath.Dir(dir) {
		if predicate(dir) {
			return dir, nil
		}
	}
	return "", os.ErrNotExist
}
