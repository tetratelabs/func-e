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

package envoy

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/manifest"
)

func TestPrintVersions(t *testing.T) {
	tests := []struct {
		name, goos string
		manifest   *manifest.Manifest
		expected   string
	}{
		{
			name:     "darwin",
			goos:     "darwin",
			manifest: goodManifest,
			expected: `1.18.3
1.14.6
`,
		},
		{
			name:     "linux",
			goos:     "linux",
			manifest: goodManifest,
			expected: `1.17.1
1.14.6
`,
		},
		{
			name:     "unsupported OS",
			goos:     "windows",
			manifest: goodManifest,
		},
		{
			name:     "empty manifest",
			goos:     "darwin",
			manifest: &manifest.Manifest{},
			expected: ``,
		},
		{
			name: "non-standard manifest",
			goos: "darwin",
			manifest: &manifest.Manifest{
				ManifestVersion: "v0.1.0",
				Flavors: map[string]*manifest.Flavor{
					"non-standard": {
						Name:     "non-standard",
						Versions: goodManifest.Flavors["standard"].Versions,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			out := new(bytes.Buffer)
			PrintVersions(tc.manifest, tc.goos, out)
			require.Equal(t, tc.expected, out.String())
		})
	}
}

var goodManifest = &manifest.Manifest{
	ManifestVersion: "v0.1.0",
	Flavors: map[string]*manifest.Flavor{
		"standard": {
			Name: "standard",
			Versions: map[string]*manifest.Version{
				"1.14.6": {
					Name: "1.14.6",
					Builds: map[string]*manifest.Build{
						"LINUX_GLIBC": {
							Platform:            "LINUX_GLIBC",
							DownloadLocationURL: "https://getenvoy.io/versions/1.14.6/envoy-1.14.6-linux-x86_64.tar.gz",
						},
						"DARWIN": {
							Platform:            "DARWIN",
							DownloadLocationURL: "https://getenvoy.io/versions/1.14.6/envoy-1.14.6-darwin-x86_64.tar.gz",
						},
					},
				},
				"1.17.1": {
					Name: "1.17.1",
					Builds: map[string]*manifest.Build{
						"LINUX_GLIBC": {
							Platform:            "LINUX_GLIBC",
							DownloadLocationURL: "https://getenvoy.io/versions/1.17.1/envoy-1.17.1-linux-x86_64.tar.gz",
						},
					},
				},
				"1.18.3": {
					Name: "1.18.3",
					Builds: map[string]*manifest.Build{
						"DARWIN": {
							Platform:            "DARWIN",
							DownloadLocationURL: "https://getenvoy.io/versions/1.18.3/envoy-1.18.3-darwin-x86_64.tar.gz",
						},
					},
				},
			},
		},
	},
}
