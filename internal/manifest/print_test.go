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
)

func TestPrint(t *testing.T) {
	got := bytes.NewBuffer(nil)
	err := Print(goodManifest, got)

	require.NoError(t, err)
	require.Equal(t, `REFERENCE                                VERSION
nightly/linux-glibc                      nightly
1.17.1/linux-glibc                       1.17.1
1.17.1/darwin                            1.17.1
standard-fips1402:1.10.0/linux-glibc     1.10.0
`, got.String())
}

var goodManifest = &Manifest{
	ManifestVersion: "v0.1.0",
	Flavors: map[string]*Flavor{
		"standard": {
			Name: "standard",
			Versions: map[string]*Version{
				"1.17.1": {
					Name: "1.17.1",
					Builds: map[string]*Build{
						"LINUX_GLIBC": {
							Platform:            "LINUX_GLIBC",
							DownloadLocationURL: "standard:1.17.1/linux-glibc",
						},
						"DARWIN": {
							Platform:            "DARWIN",
							DownloadLocationURL: "standard:1.17.1/darwin",
						},
					},
				},
				"nightly": {
					Name: "nightly",
					Builds: map[string]*Build{
						"LINUX_GLIBC": {
							Platform:            "LINUX_GLIBC",
							DownloadLocationURL: "standard:nightly/linux-glibc",
						},
					},
				},
			},
		},
		"standard-fips1402": {
			Name: "standard-fips1402",
			Versions: map[string]*Version{
				"1.10.0": {
					Name: "1.10.0",
					Builds: map[string]*Build{
						"LINUX_GLIBC": {
							Platform:            "LINUX_GLIBC",
							DownloadLocationURL: "standard-fips1402:1.10.0/linux-glibc",
						},
					},
				},
			},
		},
	},
}
