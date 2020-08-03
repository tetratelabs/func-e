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

const (
	// extensionMetaDir identifies a directory that holds meta information
	// about an extension created by getenvoy toolkit.
	extensionMetaDir = ".getenvoy/extension"
)

// WorkspaceDir represents a directory with an extension created by getenvoy toolkit.
type WorkspaceDir interface {
	// GetRootDir returns path to the root dir of an extension.
	GetRootDir() string
	// GetMetaDir returns path to the meta dir of an extension.
	GetMetaDir() string

	// Rel returns path relative to the workspace root dir
	// for a given file in the meta dir.
	Rel(path string) string
	// Abs returns absolute path for a given file in the meta dir.
	Abs(path string) string

	// HasFile checks whether meta dir includes a file with a given name.
	HasFile(path string) (bool, error)
	// ReadFile reads from a given file in the meta dir.
	ReadFile(path string) ([]byte, error)
	// WriteFile writes into a given file in the meta dir.
	WriteFile(path string, data []byte) error

	// HasDir checks whether meta dir includes a directory with a given name.
	HasDir(path string) (bool, error)

	// ListDirs lists directories under a given path (non-recursively).
	ListDirs(path string) ([]string, error)
	// ListFiles recursively lists files under a given path.
	ListFiles(path string) ([]string, error)

	// RemoveAll recursively removes all files under a given path.
	RemoveAll(path string) error
}
