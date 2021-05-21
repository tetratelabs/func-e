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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocateBuild(t *testing.T) {
	tests := []struct{ name, reference, want, wantErr string }{
		{
			name:      "standard 1.17.1 linux-glibc matches",
			reference: "standard:1.17.1/linux-glibc",
			want:      "standard:1.17.1/linux-glibc",
		},
		{
			name:      "standard 1.17.1 matches",
			reference: "standard:1.17.1",
			want:      fmt.Sprintf("standard:1.17.1/%v", currentPlatform()),
		},
		{
			name:      "standard-fips1402:1.10.0/linux-glibc matches",
			reference: "standard-fips1402:1.10.0/linux-glibc",
			want:      "standard-fips1402:1.10.0/linux-glibc",
		},
		{
			name:      "sTanDard:nIgHTLY/LiNuX-gLiBc matches",
			reference: "sTanDard:nIgHTLY/LiNuX-gLiBc",
			want:      "standard:nightly/linux-glibc",
		},
		{
			name:      "Error if not found",
			reference: "notaFlavor:1.17.1/notaPlatform",
			wantErr:   `unable to find matching GetEnvoy build for reference "notaflavor:1.17.1/notaplatform"`,
		},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			key, err := ParseReference(tc.reference)
			require.NoError(t, err)

			if have, err := LocateBuild(key, goodManifest); tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
				require.Equal(t, "", have)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, have)
			}
		})
	}
}

func TestFetchManifest(t *testing.T) {
	goodManifestBytes, err := json.Marshal(goodManifest)
	require.NoError(t, err)

	// This tests we can parse a flavor besides "standard", and don't crash on new fields
	nonStandardManifestBytes := []byte(`{
  "manifestVersion": "v0.1.2",
  "flavors": {
    "experiment": {
      "name": "experiment",
      "versions": {
        "1.15": {
          "name": "1.15",
          "builds": {
            "DARWIN": {
              "downloadLocationUrl": "1.17.1/darwin",
              "platform": "DARWIN",
              "ARCH": "amd64"
            }
          }
        }
      }
    }
  }
}`)

	var nonStandardManifest = &Manifest{
		ManifestVersion: "v0.1.2",
		Flavors: map[string]*Flavor{
			"experiment": {
				Name: "experiment",
				Versions: map[string]*Version{
					"1.15": {
						Name: "1.15",
						Builds: map[string]*Build{
							"DARWIN": {
								Platform:            "DARWIN",
								DownloadLocationURL: "1.17.1/darwin",
							},
						},
					},
				},
			},
		},
	}

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
			name:                  "allows non-standard flavor and unknown fields",
			responseStatusCode:    http.StatusOK,
			responseManifestBytes: nonStandardManifestBytes,
			want:                  nonStandardManifest, // instead of crash
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
			have, err := FetchManifest(mock.URL)
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
