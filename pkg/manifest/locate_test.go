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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/tetratelabs/getenvoy/api"
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
			want:      fmt.Sprintf("standard:1.17.1/%v", platform()),
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
			key, err := NewKey(tc.reference)
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
	goodManifestBytes, err := protojson.Marshal(goodManifest)
	require.NoError(t, err)

	tests := []struct {
		name                  string
		responseStatusCode    int
		responseManifestBytes []byte
		want                  *api.Manifest
		wantErr               bool
	}{
		{
			name:                  "responds with parsed manifest",
			responseStatusCode:    http.StatusOK,
			responseManifestBytes: goodManifestBytes,
			want:                  goodManifest,
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
			// Use prototext comparison to avoid comparing internal state
			require.Equal(t, tc.want.String(), have.String())
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewKey(t *testing.T) {
	tests := []struct {
		reference string
		want      *Key
		wantErr   bool
	}{
		{"flavor:version/platform/platform", nil, true},
		{"flavor:version/DARWIN", &Key{Flavor: "flavor", Version: "version", Platform: "darwin"}, false},
		{"flavor:version/PLATFORM_GLIBC", &Key{Flavor: "flavor", Version: "version", Platform: "platform-glibc"}, false},
		{"fLaVoR:VeRsIoN/pLaTfOrM", &Key{Flavor: "flavor", Version: "version", Platform: "platform"}, false},
		{"flavor:version/", &Key{Flavor: "flavor", Version: "version", Platform: platform()}, false},
		{"flavor:version", &Key{Flavor: "flavor", Version: "version", Platform: platform()}, false},
		{"fLaVoR:VeRsIoN", &Key{Flavor: "flavor", Version: "version", Platform: platform()}, false},
		{"flavor:", nil, true},
		{"flavor", nil, true},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.reference, func(t *testing.T) {
			have, err := NewKey(tc.reference)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, have)
			} else {
				require.Nil(t, err)
				require.Equal(t, tc.want, have)
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
