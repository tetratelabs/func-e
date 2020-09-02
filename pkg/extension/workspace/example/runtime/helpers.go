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

package runtime

import (
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
)

// GetEnvoyReference returns either a path to a custom Envoy binary or
// a version of Envoy provided by getenvoy.io.
func (o *RunOpts) GetEnvoyReference() string {
	// custom Envoy binary takes priority over Envoy version
	if path := o.GetEnvoyPath(); path != "" {
		return path
	}
	return o.GetEnvoyVersion()
}

// GetEnvoyPath returns a path to a custom Envoy binary, if any.
func (o *RunOpts) GetEnvoyPath() string {
	return o.Envoy.Path
}

// GetEnvoyVersion returns effective version of Envoy.
func (o *RunOpts) GetEnvoyVersion() string {
	// Envoy version from command line
	if o.Envoy.Version != "" {
		return o.Envoy.Version
	}
	// Envoy version from Example descriptor
	if version := o.Example.GetDescriptor().GetRuntime().GetEnvoy().GetVersion(); version != "" {
		return version
	}
	// Envoy version from Extension descriptor
	return o.Workspace.GetExtensionDescriptor().Runtime.Envoy.Version
}

// GetExtensionConfig returns effective extension config.
func (o *RunOpts) GetExtensionConfig() *model.File {
	// extension config from command line
	if o.Extension.Config.Source != "" {
		return &o.Extension.Config
	}
	// extension config from Example
	_, file := o.Example.GetExtensionConfig()
	return file
}
