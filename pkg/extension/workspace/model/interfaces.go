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

package model

import (
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/fs"
)

// Workspace represents a workspace with an extension created by getenvoy toolkit.
type Workspace interface {
	// GetDir returns extension directory.
	GetDir() fs.WorkspaceDir

	// GetExtensionDescriptor returns extension descriptor.
	GetExtensionDescriptor() *extension.Descriptor

	// HasToolchain returns true if workspace includes configuration for a toolchain
	// with a given name.
	HasToolchain(toolchain string) (bool, error)
	// GetToolchainConfigBytes returns configuration of a toolchain with a given name.
	GetToolchainConfigBytes(toolchain string) (source string, data []byte, err error)
	// SaveToolchainConfigBytes persists given toolchain configuration.
	SaveToolchainConfigBytes(toolchain string, data []byte) error
}
