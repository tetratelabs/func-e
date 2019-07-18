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
	"strings"
	"text/tabwriter"

	"github.com/tetratelabs/getenvoy-package/api"
)

// PrettyPrint tabwrites the passed manifest to the passed writer
func PrettyPrint(writer io.Writer, manifest *api.Manifest) error {
	w := new(tabwriter.Writer).Init(writer, 0, 8, 5, ' ', 0)
	fmt.Fprintln(w, "FLAVOR\tVERSION\tAVAILABLE ON")

	for _, flavor := range deterministicFlavors(manifest.Flavors) {
		for _, version := range deterministicVersions(flavor.Versions) {
			osList := []string{}
			for os := range version.GetOperatingSystems() {
				osList = append(osList, os)
			}
			sort.Slice(osList, func(i, j int) bool {
				return strings.ToUpper(osList[i]) < strings.ToUpper(osList[j])
			})
			prettyOS := strings.Trim(fmt.Sprintf("%v", osList), "[]")
			fmt.Fprintf(w, "%v\t%v\t%v\n", flavor.Name, version.Name, prettyOS)
		}
	}
	return w.Flush()
}

func deterministicFlavors(flavors map[string]*api.Flavor) []*api.Flavor {
	flavorList := []*api.Flavor{}
	for _, flavor := range flavors {
		flavorList = append(flavorList, flavor)
	}
	sort.Slice(flavorList, func(i, j int) bool {
		return flavorList[i].Name < flavorList[j].Name
	})
	return flavorList
}

func deterministicVersions(versions map[string]*api.Version) []*api.Version {
	versionList := []*api.Version{}
	for _, version := range versions {
		versionList = append(versionList, version)
	}

	sort.Slice(versionList, func(i, j int) bool {
		// Note: version is reverse alphabetical so that newer versions are first
		return versionList[i].Name > versionList[j].Name
	})
	return versionList
}
