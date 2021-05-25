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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildPlatform(t *testing.T) {
	tests := []struct{ name, goos, expected string }{
		{
			name:     "darwin",
			goos:     "darwin",
			expected: "DARWIN",
		},
		{
			name:     "linux",
			goos:     "linux",
			expected: "LINUX_GLIBC",
		},
		{
			name:     "unsupported",
			goos:     "windows",
			expected: "",
		},
		{
			name:     "empty",
			goos:     "",
			expected: "",
		},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, BuildPlatform(tc.goos))
		})
	}
}

func TestGetManifest(t *testing.T) {
	goodManifestBytes, err := json.Marshal(goodManifest)
	require.NoError(t, err)

	// This tests we can parse a flavor besides "standard", and don't crash on new fields
	unknownFieldsManifestBytes := []byte(`{
  "manifestVersion": "v0.1.0",
  "flavors": {
    "standard": {
      "name": "standard",
      "favoriteColor": "blue.. no, yellow!",

      "versions": {
        "1.14.6": {
          "name": "1.14.6",
          "builds": {
            "DARWIN": {
              "platform": "DARWIN",
              "downloadLocationUrl": "https://getenvoy.io/versions/1.14.6/envoy-1.14.6-darwin-x86_64.tar.gz"
            },
            "LINUX_GLIBC": {
              "platform": "LINUX_GLIBC",
              "downloadLocationUrl": "https://getenvoy.io/versions/1.14.6/envoy-1.14.6-linux-x86_64.tar.gz"
            }
          }
        }
      }
    }
  }
}`)

	tests := []struct {
		name                  string
		responseStatusCode    int
		responseManifestBytes []byte
		want                  *Manifest
		wantErr               bool
	}{
		{
			name:                  "responds with parsed manifest",
			responseStatusCode:    http.StatusOK,
			responseManifestBytes: goodManifestBytes,
			want:                  goodManifest,
		},
		{
			name:                  "allows unknown fields",
			responseStatusCode:    http.StatusOK,
			responseManifestBytes: unknownFieldsManifestBytes,
			want:                  goodManifest, // instead of crash
		},
		{
			name:               "errors on non-200 response",
			responseStatusCode: http.StatusInternalServerError,
			want:               nil,
			wantErr:            true,
		},
		{
			name:                  "errors on unparsable manifest",
			responseStatusCode:    http.StatusOK,
			responseManifestBytes: []byte("ice cream"),
			wantErr:               true,
		},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			mock := mockServer(tc.responseStatusCode, tc.responseManifestBytes)
			defer mock.Close()
			have, err := GetManifest(mock.URL)
			require.Equal(t, tc.want, have)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func mockServer(responseStatusCode int, responseManifestBytes []byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(responseStatusCode)
		if responseStatusCode == http.StatusOK {
			w.Write(responseManifestBytes)
		}
	}))
}

var goodManifest = &Manifest{
	ManifestVersion: "v0.1.0",
	Flavors: map[string]*Flavor{
		"standard": {
			Name: "standard",
			Versions: map[string]*Version{
				"1.14.6": {
					Name: "1.14.6",
					Builds: map[string]*Build{
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
			},
		},
	},
}
