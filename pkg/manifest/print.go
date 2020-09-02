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
	"net/http"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/golang/protobuf/jsonpb"
	"github.com/tetratelabs/getenvoy-package/api"
	"github.com/tetratelabs/getenvoy/pkg/transport"
)

// Print retrieves the manifest from the passed location and writes it to the passed writer
func Print(writer io.Writer) error {
	manifest, err := fetch(GetURL())
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(writer, 0, 8, 5, ' ', 0)
	fmt.Fprintln(w, "REFERENCE\tFLAVOR\tVERSION")

	for _, flavor := range deterministicFlavors(manifest.Flavors) {
		for _, version := range deterministicVersions(flavor.Versions) {
			for _, build := range deterministicBuilds(version.Builds) {
				ref := fmt.Sprintf("%v:%v/%v", flavor.Name, version.Name, platformFromEnum(build.Platform.String()))
				fmt.Fprintf(w, "%v\t%v\t%v\n", ref, flavor.Name, version.Name)
			}
		}
	}
	return w.Flush()
}

func platformFromEnum(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "_", "-")
	return s
}

func fetch(url string) (*api.Manifest, error) {
	// #nosec => This is by design, users can call out to wherever they like!
	resp, err := transport.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received %v response code from %v", resp.StatusCode, url)
	}
	defer resp.Body.Close() //nolint
	result := api.Manifest{}
	if err := jsonpb.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshalling manifest: %v", err)
	}
	return &result, nil
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

func deterministicBuilds(builds map[string]*api.Build) []*api.Build {
	buildList := []*api.Build{}
	for _, flavor := range builds {
		buildList = append(buildList, flavor)
	}
	sort.Slice(buildList, func(i, j int) bool {
		// Note: build is reverse alphabetical so that Linux versions are before Darwin
		return buildList[i].String() > buildList[j].String()
	})
	return buildList
}
