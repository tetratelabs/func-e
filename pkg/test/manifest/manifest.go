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

package manifest

import (
	"github.com/tetratelabs/getenvoy/pkg/manifest"
)

// NewSimpleManifest returns a new manifest auto-generated for a given
// list of references.
func NewSimpleManifest(reference string) (*manifest.Manifest, error) {
	m := new(manifest.Manifest)
	m.ManifestVersion = "v0.1.0"
	m.Flavors = make(map[string]*manifest.Flavor)
	ref, err := manifest.ParseReference(reference)
	if err != nil {
		return nil, err
	}
	flavor, exists := m.Flavors[ref.Flavor]
	if !exists {
		flavor = &manifest.Flavor{Name: ref.Flavor}
		m.Flavors[ref.Flavor] = flavor
	}
	if flavor.Versions == nil {
		flavor.Versions = make(map[string]*manifest.Version)
	}
	version, exists := flavor.Versions[ref.Version]
	if !exists {
		version = &manifest.Version{Name: ref.Version}
		flavor.Versions[ref.Version] = version
	}
	if version.Builds == nil {
		version.Builds = make(map[string]*manifest.Build)
	}
	platforms := []string{"DARWIN", "LINUX_GLIBC"}
	if ref.Platform != "" {
		platforms = []string{ref.Platform}
	}
	for _, platform := range platforms {
		version.Builds[platform] = &manifest.Build{
			Platform:            platform,
			DownloadLocationURL: ref.String(),
		}
	}
	return m, nil
}
