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
	// LanguageRust represents a Rust programming language.
	LanguageRust Language = "rust"
	// LanguageTinyGo represents a TinyGo programming language.
	LanguageTinyGo Language = "tinygo"
)

// Descriptor represents an extension created by getenvoy toolkit.
type Descriptor struct {
	config.Meta `json:",inline"`

	// Extension reference name.
	//
	// E.g., `mycompany.filters.http.custom_metrics`.
	//
	// Extension reference name is part of a contract between Envoy and a
	// WebAssembly module.
	// Envoy acts in assumption that a single WebAssembly module might include
	// multiple extensions.
	// To clarify which of the extensions needs to be instanciated,
	// Envoy configuration will use the respective reference name.
	//
	// ATTENTION: Beware that extension reference name appears in several places:
	//             1) in the source code (optionally)
	//             2) in the extension descriptor file
	//
	//            The value in the source code determines the actual behavior
	//            at runtime. In case a WebAssembly module includes only one
	//            extension, it might choose to ignore the reference name at all.
	//
	//            The value in the descriptor is used by getenvoy toolkit itself,
	//            e.g. to automate creation of example Envoy configurations.
	//
	//            At the moment, it is the responsibility of the extension developer
	//            to keep those values in sync.
	//
	// It is advised to give every extension a globally unique name.
	Name string `json:"name"`
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
