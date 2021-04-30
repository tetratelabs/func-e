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

package init

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/fs"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
)

var (
	// extensionDescriptorTemplate represents a descriptor of a new extension.
	extensionDescriptorTemplate = `#
# Envoy Wasm extension created with getenvoy toolkit.
#
kind: Extension

name: {{ .Name }}

category: {{ .Category }}
language: {{ .Language }}

# Runtime the extension is being developed against.
runtime:
  envoy:
    version: {{ .Runtime.Envoy.Version }}
`
)

func generateExtensionDescriptor(descriptor *extension.Descriptor) ([]byte, error) {
	tmpl, err := template.New("").Parse(extensionDescriptorTemplate)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	err = tmpl.Execute(&out, descriptor)
	if err != nil {
		return nil, fmt.Errorf("failed to render extension descriptor template: %w", err)
	}
	return out.Bytes(), nil
}

func generateWorkspace(opts *ScaffoldOpts) error {
	dir, err := fs.CreateExtensionDir(opts.ExtensionDir)
	if err != nil {
		return err
	}
	descriptor, err := generateExtensionDescriptor(opts.Extension)
	if err != nil {
		return err
	}
	if e := dir.WriteFile(model.DescriptorFile, descriptor); e != nil {
		return e
	}
	opts.ProgressSink.OnFile(dir.Rel(model.DescriptorFile))
	return nil
}
