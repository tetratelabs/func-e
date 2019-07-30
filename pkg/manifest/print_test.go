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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tetratelabs/getenvoy-package/api"
)

func TestPrint(t *testing.T) {
	tests := []struct {
		name             string
		wantOutputFile   string
		wantErr          bool
		locationOverride string
	}{
		{
			name:           "Prints golden output",
			wantOutputFile: "list.golden",
		},

		{
			name:             "Errors on non-URL inputs",
			locationOverride: "dir/file.json",
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			mock := mockServer(http.StatusOK, "manifest.golden")
			defer mock.Close()
			got := bytes.NewBuffer(nil)
			location := mock.URL
			if tc.locationOverride != "" {
				location = tc.locationOverride
			}
			if err := Print(got, location); tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			want, _ := ioutil.ReadFile(filepath.Join("testdata", tc.wantOutputFile))
			assert.Equal(t, string(want), got.String())
		})
	}
}

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
			responseStatusCode:   http.StatusOK,
			responseManifestFile: "manifest.golden",
			want:                 goodManifest(),
		},
		{
			name:               "errors on non-200 response",
			responseStatusCode: http.StatusInternalServerError,
			want:               nil,
			wantErr:            true,
		},
		{
			name:                 "errors on unparsable manifest",
			responseStatusCode:   http.StatusOK,
			responseManifestFile: "malformed.golden",
			wantErr:              true,
		},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			mock := mockServer(tc.responseStatusCode, tc.responseManifestFile)
			defer mock.Close()
			got, err := fetch(mock.URL)
			assert.Equal(t, tc.want, got)
			if tc.wantErr {
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
		if responseStatusCode == http.StatusOK {
			bytes, _ := ioutil.ReadFile(filepath.Join("testdata", responseManifestFile))
			w.Write(bytes)
		}
	}))
}

func goodManifest() *api.Manifest {
	return &api.Manifest{
		ManifestVersion: "v0.1.0",
		Flavors: map[string]*api.Flavor{
			"standard": {
				Name:          "standard",
				FilterProfile: "standard",
				Versions: map[string]*api.Version{
					"1.11.0": {
						Name: "1.11.0",
						Builds: map[string]*api.Build{
							api.Build_LINUX_GLIBC.String(): {
								Platform:            api.Build_LINUX_GLIBC,
								DownloadLocationUrl: "standard:1.11.0/linux-glibc",
							},
							api.Build_DARWIN.String(): {
								Platform:            api.Build_DARWIN,
								DownloadLocationUrl: "standard:1.11.0/darwin",
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
}
