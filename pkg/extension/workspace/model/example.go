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

	"gopkg.in/yaml.v3"

	exampleconfig "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/example"
)

var (
	// notice that the file is named `envoy.tmpl.yaml` rather than `envoy.tmpl.yaml`
	// in order to meet constraints of Envoy.
	exampleEnvoyBootstrapFileAltNames  = []string{"envoy.tmpl.yaml", "envoy.tmpl.json"}
	exampleExtensionConfigFileAltNames = []string{"extension.yaml", "extension.json", "extension.txt", "extension"}
)

// NewExample returns a new Example that consists of a given set of files.
func NewExample(files ImmutableFileSet) (Example, error) {
	if !files.Has(exampleDescriptorFile) {
		return nil, fmt.Errorf("extension descriptor file %q is missing", exampleDescriptorFile)
	}
	descriptor, err := parseExampleDescriptor(files.Get(exampleDescriptorFile))
	if err != nil {
		return nil, err
	}
	exampl := &example{
		files:      files,
		descriptor: descriptor,
	}
	if _, file := exampl.GetEnvoyConfig(); file == nil {
		return nil, fmt.Errorf("envoy bootstrap config file is missing: every example must include one of %v", exampleEnvoyBootstrapFileAltNames)
	}
	if _, file := exampl.GetExtensionConfig(); file == nil {
		return nil, fmt.Errorf("extension config file is missing: every example must include one of %v", exampleExtensionConfigFileAltNames)
	}
	return exampl, nil
}

func parseExampleDescriptor(file *File) (*exampleconfig.Descriptor, error) {
	descriptor := exampleconfig.NewExampleDescriptor()
	err := yaml.Unmarshal(file.Content, descriptor)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal example descriptor %s: %w", file.Source, err)
	}
	descriptor.Default()
	err = descriptor.Validate()
	if err != nil {
		return nil, fmt.Errorf("example descriptor is not valid %s: %w", file.Source, err)
	}
	return descriptor, nil
}

type example struct {
	files      ImmutableFileSet
	descriptor *exampleconfig.Descriptor
}

func (e *example) GetFiles() ImmutableFileSet {
	return e.files
}

func (e *example) GetDescriptor() *exampleconfig.Descriptor {
	return e.descriptor
}

func (e *example) GetEnvoyConfig() (string, *File) {
	return e.getFirstPresentFile(exampleEnvoyBootstrapFileAltNames)
}

func (e *example) GetExtensionConfig() (string, *File) {
	return e.getFirstPresentFile(exampleExtensionConfigFileAltNames)
}

func (e *example) getFirstPresentFile(altNames []string) (string, *File) {
	for _, fileName := range altNames {
		if e.files.Has(fileName) {
			return fileName, e.files.Get(fileName)
		}
	}
	return "", nil
}
