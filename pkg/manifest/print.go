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
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"text/tabwriter"

	"net/url"

	"github.com/golang/protobuf/jsonpb"
	"github.com/tetratelabs/getenvoy-package/api"
)

// Print retrieves the manifest from the passed location and writes it to the passed writer
func Print(writer io.Writer, manifestLocation string) error {
	if _, err := url.Parse(manifestLocation); err != nil {
		return errors.New("only URL manifest locations are supported")
	}
	manifest, err := fetch(manifestLocation)
	if err != nil {
		return err
	}
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

func fetch(manifestURL string) (*api.Manifest, error) {
	// #nosec => This is by design, users can call out to wherever they like!
	resp, err := http.Get(manifestURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received %v response code from %v", resp.StatusCode, manifestURL)
	}
	defer func() { _ = resp.Body.Close() }()
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
