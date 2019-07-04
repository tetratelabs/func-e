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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tetratelabs/getenvoy/api"
)

func TestFetch(t *testing.T) {
	tests := []struct {
		name                 string
		responseStatusCode   int
		responseManifestFile string
		want                 *api.Manifest
		wantErr              bool
	}{
		{
			name:                 "responds with parsed manifest",
			responseStatusCode:   200,
			responseManifestFile: "manifest.golden",
			want:                 goodManifest(),
		},
		{
			name:               "errors on non-200 response",
			responseStatusCode: 500,
			wantErr:            true,
		},
		{
			name:                 "errors on unparsable manifest",
			responseStatusCode:   200,
			responseManifestFile: "malformed.golden",
			wantErr:              true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockServer(tt.responseStatusCode, tt.responseManifestFile)
			defer mock.Close()
			got, err := Fetch(mock.URL)
			assert.Equal(t, tt.want, got)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func mockServer(responseStatusCode int, responseManifestFile string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(responseStatusCode)
		if responseStatusCode == 200 {
			bytes, _ := ioutil.ReadFile(filepath.Join("testdata", responseManifestFile))
			w.Write(bytes)
		}
	}))
}

func goodManifest() *api.Manifest {
	return &api.Manifest{
		Version: "v0.1.0",
		Flavors: map[string]*api.Flavor{
			"standard": &api.Flavor{
				Name:          "standard",
				FilterProfile: "standard",
				Versions: map[string]*api.Version{
					"1.10.0": &api.Version{
						Name: "1.10.0",
						OperatingSystems: map[string]*api.OperatingSystem{
							"Ubuntu": &api.OperatingSystem{
								Name: api.OperatingSystemName_UBUNTU,
								Builds: []*api.Build{
									&api.Build{
										OperatingSystemVersion: "16.04",
										DownloadLocationUrl:    "http://example.com",
									},
								},
							},
							"macOS": &api.OperatingSystem{
								Name: api.OperatingSystemName_MACOS,
								Builds: []*api.Build{
									&api.Build{
										OperatingSystemVersion: "10.14",
										DownloadLocationUrl:    "http://example.com",
									},
								},
							},
							"CentOS": &api.OperatingSystem{
								Name: api.OperatingSystemName_CENTOS,
								Builds: []*api.Build{
									&api.Build{
										OperatingSystemVersion: "7",
										DownloadLocationUrl:    "http://example.com",
									},
								},
							},
						},
					},
					"nightly": &api.Version{
						Name: "nightly",
						OperatingSystems: map[string]*api.OperatingSystem{
							"Ubuntu": &api.OperatingSystem{
								Name: api.OperatingSystemName_UBUNTU,
								Builds: []*api.Build{
									&api.Build{
										OperatingSystemVersion: "16.04",
										DownloadLocationUrl:    "http://example.com",
									},
								},
							},
							"macOS": &api.OperatingSystem{
								Name: api.OperatingSystemName_MACOS,
								Builds: []*api.Build{
									&api.Build{
										OperatingSystemVersion: "10.14",
										DownloadLocationUrl:    "http://example.com",
									},
								},
							},
							"CentOS": &api.OperatingSystem{
								Name: api.OperatingSystemName_CENTOS,
								Builds: []*api.Build{
									&api.Build{
										OperatingSystemVersion: "7",
										DownloadLocationUrl:    "http://example.com",
									},
								},
							},
						},
					},
				},
			},
			"standard-fips1402": &api.Flavor{
				Name:          "standard-fips1402",
				FilterProfile: "standard",
				Compliances:   []api.Compliance{api.Compliance_FIPS_1402},
				Versions: map[string]*api.Version{
					"1.10.0": &api.Version{
						Name: "1.10.0",
						OperatingSystems: map[string]*api.OperatingSystem{
							"Ubuntu": &api.OperatingSystem{
								Name: api.OperatingSystemName_UBUNTU,
								Builds: []*api.Build{
									&api.Build{
										OperatingSystemVersion: "16.04",
										DownloadLocationUrl:    "http://example.com",
									},
								},
							},
							"CentOS": &api.OperatingSystem{
								Name: api.OperatingSystemName_CENTOS,
								Builds: []*api.Build{
									&api.Build{
										OperatingSystemVersion: "7",
										DownloadLocationUrl:    "http://example.com",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
