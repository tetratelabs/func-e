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
	"github.com/tetratelabs/getenvoy/api"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
)

// NewSimpleManifest returns a new manifest auto-generated for a given
// list of references.
func NewSimpleManifest(reference string) (*api.Manifest, error) {
	m := new(api.Manifest)
	m.ManifestVersion = "v0.1.0"
	m.Flavors = make(map[string]*api.Flavor)
	ref, err := manifest.ParseReference(reference)
	if err != nil {
		return nil, err
	}
	flavor, exists := m.Flavors[ref.Flavor]
	if !exists {
		flavor = &api.Flavor{Name: ref.Flavor, FilterProfile: ref.Flavor}
		m.Flavors[ref.Flavor] = flavor
	}
	if flavor.Versions == nil {
		flavor.Versions = make(map[string]*api.Version)
	}
	version, exists := flavor.Versions[ref.Version]
	if !exists {
		version = &api.Version{Name: ref.Version}
		flavor.Versions[ref.Version] = version
	}
	if version.Builds == nil {
		version.Builds = make(map[string]*api.Build)
	}
	platforms := SupportedPlatforms
	if ref.Platform != "" {
		platform, err := ParsePlatform(ref.Platform)
		if err != nil {
			return nil, err
		}
		platforms = Platforms{platform}
	}
	for _, platform := range platforms {
		version.Builds[platform.Code()] = &api.Build{
			Platform:            platform.BuildPlatform(),
			DownloadLocationUrl: ref.String(),
		}
	}
	return m, nil
}
