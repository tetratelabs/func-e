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

// CreateExtensionDir creates a new extension directory at a given path.
func CreateExtensionDir(at string) (ExtensionDir, error) {
	dir := filepath.Join(at, extensionMetaDir)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create extension directory %q: %w", at, err)
	}
	return GetExtensionDir(at)
}

// GetExtensionDir returns the extension directory rooted at a given path
// or an error otherwise.
func GetExtensionDir(at string) (ExtensionDir, error) {
	path, err := filepath.Abs(at)
	if err != nil {
		return nil, fmt.Errorf("failed to determine if there is an extension directory %q: %w", at, err)
	}
	if IsExtensionDir(path) {
		return extensionDir(path), nil
	}
	return nil, fmt.Errorf("not an extension directory %q", path)
}

// IsExtensionDir returns true if a given path corresponds to a directory
// with an extension created by getenvoy toolkit.
func IsExtensionDir(path string) bool {
	metaDir := filepath.Join(path, extensionMetaDir)
	info, err := os.Stat(metaDir)
	if err != nil {
		return false
	}
	return info.Mode().IsDir()
}
