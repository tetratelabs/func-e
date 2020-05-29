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
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/fs"
)

const (
	// configExtension represents an extension of all configuration files,
	// including extension descriptor, built-in toolchain config, custom toolchain config, etc.
	configExtension = ".yaml"
	// DescriptorFile represents a relative path of the extension descriptor file.
	DescriptorFile = "extension" + configExtension
	toolchainsDir  = "toolchains"
)

// WorkspaceAt returns a workspace at a given directory.
func WorkspaceAt(dir fs.WorkspaceDir) (Workspace, error) {
	// fail early if extension descriptor is not valid
	descriptor, err := getExtensionDescriptor(dir)
	if err != nil {
		return nil, err
	}
	return &workspace{dir, descriptor}, nil
}

// getExtensionDescriptor returns extension descriptor from a given workspace directory.
func getExtensionDescriptor(dir fs.WorkspaceDir) (*extension.Descriptor, error) {
	path := DescriptorFile
	data, err := dir.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read extension descriptor: %s", dir.Abs(path))
	}
	descriptor := extension.NewExtensionDescriptor()
	err = config.Unmarshal(data, descriptor)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal extension descriptor: %s", dir.Abs(path))
	}
	descriptor.Default()
	err = descriptor.Validate()
	if err != nil {
		return nil, errors.Wrapf(err, "extension descriptor is not valid: %s", dir.Abs(path))
	}
	return descriptor, nil
}

// workspace represents a workspace with an extension created by getenvoy toolkit.
type workspace struct {
	dir       fs.WorkspaceDir
	extension *extension.Descriptor
}

func (w *workspace) GetDir() fs.WorkspaceDir {
	return w.dir
}

func (w workspace) GetExtensionDescriptor() *extension.Descriptor {
	return w.extension
}

func (w workspace) HasToolchain(toolchain string) (bool, error) {
	path := w.toolchainConfigFile(toolchain)
	return w.dir.HasFile(path, func(info os.FileInfo) error {
		if !info.Mode().IsRegular() {
			return errors.Errorf("toolchain configuration must be stored in a regular file, however given file is not regular: %s", w.dir.Abs(path))
		}
		return nil
	})
}

func (w workspace) GetToolchainConfigBytes(toolchain string) (string, []byte, error) {
	path := w.toolchainConfigFile(toolchain)
	data, err := w.dir.ReadFile(path)
	if err != nil {
		return "", nil, errors.Wrapf(err, "failed to read toolchain configuration: %s", w.dir.Abs(path))
	}
	return w.dir.Abs(path), data, nil
}

func (w workspace) SaveToolchainConfigBytes(toolchain string, data []byte) error {
	path := w.toolchainConfigFile(toolchain)
	err := w.dir.WriteFile(path, data)
	if err != nil {
		return errors.Wrapf(err, "failed to write toolchain configuration: %s", w.dir.Abs(path))
	}
	return nil
}

func (w workspace) toolchainConfigFile(toolchain string) string {
	return filepath.Join(toolchainsDir, toolchain+configExtension)
}
