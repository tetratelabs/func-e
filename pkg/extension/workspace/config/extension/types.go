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

package extension

import (
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config"
)

const (
	// Kind identifies extension descriptor.
	Kind = "Extension"
)

// Category represents an extension category.
type Category string

func (c Category) String() string {
	return string(c)
}

const (
	// EnvoyHTTPFilter represents an Envoy HTTP filter.
	EnvoyHTTPFilter Category = "envoy.filters.http"
	// EnvoyNetworkFilter represents an Envoy Network filter.
	EnvoyNetworkFilter Category = "envoy.filters.network"
	// EnvoyAccessLogger represents an Envoy Access Logger.
	EnvoyAccessLogger Category = "envoy.access_loggers"
)

// Language represents a programming language.
type Language string

func (l Language) String() string {
	return string(l)
}

const (
	// LanguageRust represens a Rust programming language.
	LanguageRust Language = "rust"
)

// Descriptor represents an extension created by getenvoy toolkit.
type Descriptor struct {
	config.Meta `json:",inline"`

	// Extension category.
	Category Category `json:"category"`
	// Extension language.
	Language Language `json:"language"`

	// Runtime the extension is being developed against.
	Runtime Runtime `json:"runtime"`
}

// Runtime describes the runtime the extension is being developed against.
type Runtime struct {
	// `Envoy` runtime.
	Envoy EnvoyRuntime `json:"envoy"`
}

// EnvoyRuntime describes `Envoy` runtime the extension is being developed against.
type EnvoyRuntime struct {
	// Version of `Envoy` provided by getenvoy.io.
	Version string `json:"version"`
}
