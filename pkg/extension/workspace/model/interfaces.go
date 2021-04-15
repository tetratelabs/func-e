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
	exampleconfig "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/example"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/fs"
	scaffoldutil "github.com/tetratelabs/getenvoy/pkg/util/scaffold"
)

// Workspace represents a workspace with an extension created by getenvoy toolkit.
type Workspace interface {
	// GetDir returns extension directory.
	GetDir() fs.WorkspaceDir

	// GetExtensionDescriptor returns extension descriptor.
	GetExtensionDescriptor() *extension.Descriptor

	// HasToolchain returns true if workspace includes configuration for a toolchain
	// with a given name.
	HasToolchain(toolchainName string) (bool, error)
	// GetToolchainConfig returns configuration of a toolchain with a given name.
	GetToolchainConfig(toolchainName string) (*File, error)
	// SaveToolchainConfig persists given toolchain configuration.
	SaveToolchainConfig(toolchainName string, data []byte) error

	// ListExamples returns a list of existing examples.
	ListExamples() ([]string, error)
	// HasExample returns true if workspace includes an example with a given name.
	HasExample(exampleName string) (bool, error)
	// GetExample returns an example with a given name.
	GetExample(exampleName string) (Example, error)
	// SaveExample persists a given example.
	SaveExample(exampleName string, example Example, opts ...SaveOption) error
	// RemoveExample removes a given example.
	RemoveExample(exampleName string, opts ...RemoveOption) error
}

// Example represents an example.
type Example interface {
	GetFiles() ImmutableFileSet
	GetDescriptor() *exampleconfig.Descriptor
	GetEnvoyConfig() (string, *File)
	GetExtensionConfig() (string, *File)
}

// ImmutableFileSet represents an immutable set of configuration files.
type ImmutableFileSet interface {
	GetNames() []string
	Has(name string) bool
	Get(name string) *File
}

// FileSet represents a mutable set of configuration files.
type FileSet interface {
	ImmutableFileSet
	Add(name string, file *File)
}

// File represents a configuration file.
type File struct {
	Source  string
	Content []byte
}

// SaveOption modifies behavior of Save operations.
type SaveOption interface {
	ApplyToSave(opts *SaveOptions)
}

// SaveOptions contains options of a Save operation.
type SaveOptions struct {
	progress scaffoldutil.ProgressSink
}

// ApplyOptions modifies options of a Save operation.
func (o *SaveOptions) ApplyOptions(opts ...SaveOption) *SaveOptions {
	for _, opt := range opts {
		opt.ApplyToSave(o)
	}
	return o
}

// Default sets default values to optional fields.
func (o *SaveOptions) Default() {
	if o.progress == nil {
		o.progress = scaffoldutil.NoOpProgressSink()
	}
}

// RemoveOption modifies behavior of Remove operations.
type RemoveOption interface {
	ApplyToRemove(opts *RemoveOptions)
}

// RemoveOptions contains options of a Remove operation.
type RemoveOptions struct {
	progress scaffoldutil.ProgressSink
}

// ApplyOptions modifies options of a Remove operation.
func (o *RemoveOptions) ApplyOptions(opts ...RemoveOption) *RemoveOptions {
	for _, opt := range opts {
		opt.ApplyToRemove(o)
	}
	return o
}

// Default sets default values to optional fields.
func (o *RemoveOptions) Default() {
	if o.progress == nil {
		o.progress = scaffoldutil.NoOpProgressSink()
	}
}

// ProgressSink passes progress sink into Save and Remove operations.
type ProgressSink struct {
	scaffoldutil.ProgressSink
}

// ApplyToSave modifies options of a Save operation.
func (s ProgressSink) ApplyToSave(opts *SaveOptions) {
	opts.progress = s
}

// ApplyToRemove modifies options of a Remove operation.
func (s ProgressSink) ApplyToRemove(opts *RemoveOptions) {
	opts.progress = s
}
