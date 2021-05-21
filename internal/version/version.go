// Copyright 2021 Tetrate
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

package version

import _ "embed" // We embed the Envoy version so that we can cache it in CI

// GetEnvoy is the version of the CLI, used in help statements and HTTP requests via "User-Agent".
// Override this via "-X github.com/tetratelabs/getenvoy/internal/version.GetEnvoy=XXX"
var GetEnvoy = "dev"

// Envoy is the default version to download. This is embedded for re-use in build and CI scripts.
//go:embed envoy.txt
var Envoy string
