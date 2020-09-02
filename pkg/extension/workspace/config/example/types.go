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

package example

import (
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config"
)

const (
	// Kind identifies example.
	Kind = "Example"
)

// Descriptor represents an example created by getenvoy toolkit.
type Descriptor struct {
	config.Meta `json:",inline"`
	// Runtime the example should be run in.
	Runtime *Runtime `json:"runtime,omitempty"`
}

// Runtime describes the runtime the example should be run in.
type Runtime struct {
	// `Envoy` runtime.
	Envoy *EnvoyRuntime `json:"envoy,omitempty"`
}

// EnvoyRuntime describes `Envoy` runtime the example should use.
type EnvoyRuntime struct {
	// Version of `Envoy` provided by getenvoy.io.
	Version string `json:"version,omitempty"`
}
