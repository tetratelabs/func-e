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
	"html/template"

	"github.com/pkg/errors"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/fs"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
)

var (
	// extensionDescriptorTemplate represents a descriptor of a new extension.
	extensionDescriptorTemplate = `#
# Envoy Wasm extension created with getenvoy toolkit.
#
kind: Extension

language: {{ .Language }}
category: {{ .Category }}

# Runtime the extension is being developed against.
runtime:
  envoy:
    version: {{ .EnvoyVersion }}
`
)

func generateExtensionDescriptor(opts *ScaffoldOpts) ([]byte, error) {
	tmpl, err := template.New("").Parse(extensionDescriptorTemplate)
	if err != nil {
		// must be caught by unit tests
		panic(err)
	}
	var out bytes.Buffer
	err = tmpl.Execute(&out, opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render extension descriptor template")
	}
	return out.Bytes(), nil
}

func generateWorkspace(opts *ScaffoldOpts) error {
	dir, err := fs.CreateWorkspaceDir(opts.OutputDir)
	if err != nil {
		return err
	}
	descriptor, err := generateExtensionDescriptor(opts)
	if err != nil {
		return err
	}
	if err := dir.WriteFile(model.DescriptorFile, descriptor); err != nil {
		return err
	}
	opts.ProgressHandler.OnFile(dir.Rel(model.DescriptorFile))
	return nil
}
