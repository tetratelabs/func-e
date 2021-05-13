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

package manifest

// Manifest is derived from /manifest.proto
type Manifest struct {
	ManifestVersion string `json:"manifestVersion,omitempty"`
	// Key is the flavor name
	Flavors map[string]*Flavor `json:"flavors,omitempty"`
}

// Flavor is a type of release, typically always "standard".
type Flavor struct {
	// Name is almost always "standard"
	Name string `json:"name,omitempty"`
	// Key is the version's name
	Versions map[string]*Version `json:"versions,omitempty"`
}

// Version is an aggregation of Builds.
type Version struct {
	// Name is the Envoy version
	// Examples: 1.10.0, 1.11.0, nightly
	Name string `json:"name,omitempty"`
	// Key is the build's platform
	Builds map[string]*Build `json:"builds,omitempty"`
}

// Build associates a platform with a download URL.
type Build struct {
	// Platform ex "DARWIN", "WINDOWS", "LINUX_GLIBC"
	Platform            string `json:"platform,omitempty"`
	DownloadLocationURL string `json:"downloadLocationUrl,omitempty"`
}
