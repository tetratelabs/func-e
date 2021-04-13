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
	"github.com/tetratelabs/getenvoy/pkg/types"
)

// NewSimpleManifest returns a new manifest auto-generated for a given
// list of references.
func NewSimpleManifest(references ...string) (*api.Manifest, error) {
	manifest := new(api.Manifest)
	manifest.ManifestVersion = "v0.1.0"
	manifest.Flavors = make(map[string]*api.Flavor)
	for _, reference := range references {
		ref, err := types.ParseReference(reference)
		if err != nil {
			return nil, err
		}
		flavor, exists := manifest.Flavors[ref.Flavor]
		if !exists {
			flavor = &api.Flavor{Name: ref.Flavor, FilterProfile: ref.Flavor}
			manifest.Flavors[ref.Flavor] = flavor
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
				DownloadLocationUrl: (&types.Reference{Flavor: ref.Flavor, Version: ref.Version, Platform: platform.String()}).String(),
			}
		}
	}
	return manifest, nil
}
