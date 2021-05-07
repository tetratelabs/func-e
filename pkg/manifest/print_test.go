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
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/api"
)

func TestPrint(t *testing.T) {
	got := bytes.NewBuffer(nil)
	err := Print(goodManifest, got)

	require.NoError(t, err)
	require.Equal(t, `REFERENCE                                FLAVOR                VERSION
standard:nightly/linux-glibc             standard              nightly
standard:1.17.1/linux-glibc              standard              1.17.1
standard:1.17.1/darwin                   standard              1.17.1
standard-fips1402:1.10.0/linux-glibc     standard-fips1402     1.10.0
`, got.String())
}

var goodManifest = &api.Manifest{
	ManifestVersion: "v0.1.0",
	Flavors: map[string]*api.Flavor{
		"standard": {
			Name:          "standard",
			FilterProfile: "standard",
			Versions: map[string]*api.Version{
				"1.17.1": {
					Name: "1.17.1",
					Builds: map[string]*api.Build{
						api.Build_LINUX_GLIBC.String(): {
							Platform:            api.Build_LINUX_GLIBC,
							DownloadLocationUrl: "standard:1.17.1/linux-glibc",
						},
						api.Build_DARWIN.String(): {
							Platform:            api.Build_DARWIN,
							DownloadLocationUrl: "standard:1.17.1/darwin",
						},
					},
				},
				"nightly": {
					Name: "nightly",
					Builds: map[string]*api.Build{
						api.Build_LINUX_GLIBC.String(): {
							Platform:            api.Build_LINUX_GLIBC,
							DownloadLocationUrl: "standard:nightly/linux-glibc",
						},
					},
				},
			},
		},
		"standard-fips1402": {
			Name:          "standard-fips1402",
			FilterProfile: "standard",
			Compliances:   []api.Compliance{api.Compliance_FIPS1402},
			Versions: map[string]*api.Version{
				"1.10.0": {
					Name: "1.10.0",
					Builds: map[string]*api.Build{
						api.Build_LINUX_GLIBC.String(): {
							Platform:            api.Build_LINUX_GLIBC,
							DownloadLocationUrl: "standard-fips1402:1.10.0/linux-glibc",
						},
					},
				},
			},
		},
	},
}
