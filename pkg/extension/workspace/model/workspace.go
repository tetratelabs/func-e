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
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/fs"
)

const (
	// configExtension represents an extension of all configuration files,
	// including extension descriptor, built-in toolchain config, custom toolchain config, etc.
	configExtension = ".yaml"
	// DescriptorFile represents a relative path of the extension descriptor file.
	DescriptorFile        = "extension" + configExtension
	toolchainsDir         = "toolchains"
	examplesDir           = "examples"
	exampleDescriptorFile = "example" + configExtension
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
		return nil, fmt.Errorf("failed to read extension descriptor %s: %w", dir.Abs(path), err)
	}
	descriptor := extension.NewExtensionDescriptor()
	err = yaml.Unmarshal(data, descriptor)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal extension descriptor %s: %w", dir.Abs(path), err)
	}
	descriptor.Default()
	err = descriptor.Validate()
	if err != nil {
		return nil, fmt.Errorf("extension descriptor is not valid %s: %w", dir.Abs(path), err)
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

func (w workspace) HasToolchain(toolchainName string) (bool, error) {
	return w.dir.HasFile(toolchainLayout(toolchainName).ConfigFile())
}

func (w workspace) GetToolchainConfig(toolchainName string) (*File, error) {
	return w.readFile(toolchainLayout(toolchainName).ConfigFile())
}

func (w workspace) SaveToolchainConfig(toolchainName string, data []byte) error {
	return w.writeFile(toolchainLayout(toolchainName).ConfigFile(), data)
}

func (w workspace) ListExamples() ([]string, error) {
	return w.dir.ListDirs(examplesDir)
}

func (w workspace) HasExample(exampleName string) (bool, error) {
	return w.dir.HasDir(exampleLayout(exampleName).RootDir())
}

func (w workspace) GetExample(exampleName string) (Example, error) {
	fileNames, err := w.dir.ListFiles(exampleLayout(exampleName).RootDir())
	if err != nil {
		return nil, err
	}
	fileSet := NewFileSet()
	for _, fileName := range fileNames {
		file, err := w.readFile(exampleLayout(exampleName).File(fileName))
		if err != nil {
			return nil, err
		}
		fileSet.Add(fileName, file)
	}
	return NewExample(fileSet)
}

func (w workspace) SaveExample(exampleName string, example Example, opts ...SaveOption) error {
	options := SaveOptions{}
	options.ApplyOptions(opts...).Default()

	options.progress.OnStart()
	if err := w.removeAll(exampleLayout(exampleName).RootDir()); err != nil {
		return err
	}
	for _, fileName := range example.GetFiles().GetNames() {
		path := exampleLayout(exampleName).File(fileName)
		if err := w.writeFile(path, example.GetFiles().Get(fileName).Content); err != nil {
			return err
		}
		options.progress.OnFile(w.dir.Rel(path))
	}
	options.progress.OnComplete()
	return nil
}

func (w workspace) RemoveExample(exampleName string, opts ...RemoveOption) error {
	options := RemoveOptions{}
	options.ApplyOptions(opts...).Default()

	options.progress.OnStart()
	fileNames, err := w.dir.ListFiles(exampleLayout(exampleName).RootDir())
	if err != nil {
		return err
	}
	for _, fileName := range fileNames {
		path := exampleLayout(exampleName).File(fileName)
		if err := w.removeAll(path); err != nil {
			return err
		}
		options.progress.OnFile(w.dir.Rel(path))
	}
	if err := w.removeAll(exampleLayout(exampleName).RootDir()); err != nil {
		return err
	}
	options.progress.OnComplete()
	return nil
}

func (w workspace) readFile(path string) (*File, error) {
	data, err := w.dir.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", w.dir.Abs(path), err)
	}
	return &File{Source: w.dir.Abs(path), Content: data}, nil
}

func (w workspace) writeFile(path string, data []byte) error {
	err := w.dir.WriteFile(path, data)
	if err != nil {
		return fmt.Errorf("failed to write %s: %w", w.dir.Abs(path), err)
	}
	return nil
}

func (w workspace) removeAll(path string) error {
	err := w.dir.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("failed to remove %s: %w", w.dir.Abs(path), err)
	}
	return nil
}

type toolchainLayout string

func (toolchain toolchainLayout) ConfigFile() string {
	return filepath.Join(toolchainsDir, string(toolchain)+configExtension)
}

type exampleLayout string

func (example exampleLayout) RootDir() string {
	return filepath.Join(examplesDir, string(example))
}

func (example exampleLayout) File(name string) string {
	return filepath.Join(example.RootDir(), name)
}
