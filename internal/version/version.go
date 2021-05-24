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

// EnvoyVersions include metadata about the latest version and all available tarball URLs
type EnvoyVersions struct {
	// LatestVersion is the latest stable version. See https://github.com/envoyproxy/envoy/blob/main/RELEASES.md
	LatestVersion string
	// Versions are all stable versions of Envoy keyed on version number (without a 'v' prefix)
	Versions map[string]EnvoyVersion
}

// EnvoyVersion is the release date and tarballs for this version, keyed on platform
type EnvoyVersion struct {
	// ReleaseDate is the date of the version tag https://github.com/envoyproxy/envoy/versions. Ex. "2021-04-16"
	ReleaseDate string `json:"releaseDate"`

	// Tarballs is a map of platform to tarball URL
	// platform is '$os/$arch' where $os is the supported operating system (runtime.GOOS) and $arch
	// is the supported architecture (runtime.GOARCH). Ex "darwin/amd64"
	// tarball URL is the URL of the tar.gz or tar.xz that bundles Envoy. Minimum contents are "$version/bin/envoy".
	Tarballs map[string]string `json:"tarballs,omitempty"`
}
