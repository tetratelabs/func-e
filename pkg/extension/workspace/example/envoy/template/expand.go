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

package template

import (
	"bytes"
	"text/template"

	"github.com/pkg/errors"

	"github.com/tetratelabs/getenvoy/pkg/extension/manager"
)

// ExpandContext represents a context of the Expand operation.
type ExpandContext struct {
	DefaultExtension       manager.Extension
	DefaultExtensionConfig string
}

// Expand resolves placeholders in a given Envoy config template.
func Expand(content []byte, ctx *ExpandContext) ([]byte, error) {
	tmpl, err := template.New("").Parse(string(content))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse Envoy config template")
	}
	data := newExpandData(ctx)
	out := new(bytes.Buffer)
	err = tmpl.Execute(out, data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render Envoy config template")
	}
	return out.Bytes(), nil
}
