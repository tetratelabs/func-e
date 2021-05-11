// Copyright 2019 Tetrate
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
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"github.com/tetratelabs/getenvoy/api"
)

// Print retrieves the manifest from the passed location and writes it to the passed writer
func Print(manifest *api.Manifest, writer io.Writer) error {
	w := tabwriter.NewWriter(writer, 0, 8, 5, ' ', 0)
	fmt.Fprintln(w, "REFERENCE\tVERSION")

	for _, flavor := range deterministicFlavors(manifest.Flavors) {
		for _, version := range deterministicVersions(flavor.Versions) {
			for _, build := range deterministicBuilds(version.Builds) {
				ref := Reference{flavor.Name, version.Name, platformFromEnum(build.Platform.String())}
				fmt.Fprintf(w, "%s\t%s\n", ref.String(), version.Name)
			}
		}
	}
	return w.Flush()
}

func deterministicFlavors(flavors map[string]*api.Flavor) []*api.Flavor {
	flavorList := make([]*api.Flavor, 0, len(flavors))
	for _, flavor := range flavors {
		flavorList = append(flavorList, flavor)
	}
	sort.Slice(flavorList, func(i, j int) bool {
		return flavorList[i].Name < flavorList[j].Name
	})
	return flavorList
}

func deterministicVersions(versions map[string]*api.Version) []*api.Version {
	versionList := make([]*api.Version, 0, len(versions))
	for _, version := range versions {
		versionList = append(versionList, version)
	}
	sort.Slice(versionList, func(i, j int) bool {
		// Note: version is reverse alphabetical so that newer versions are first
		return versionList[i].Name > versionList[j].Name
	})
	return versionList
}

func deterministicBuilds(builds map[string]*api.Build) []*api.Build {
	buildList := make([]*api.Build, 0, len(builds))
	for _, build := range builds {
		buildList = append(buildList, build)
	}
	sort.Slice(buildList, func(i, j int) bool {
		// Note: build is reverse alphabetical so that Linux versions are before Darwin
		return buildList[i].String() > buildList[j].String()
	})
	return buildList
}
