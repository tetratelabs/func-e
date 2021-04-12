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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocate(t *testing.T) {
	tests := []struct {
		name               string
		reference          string
		envVar             string
		locationOverride   string
		responseStatusCode int
		want               string
		wantErr            bool
	}{
		{
			name:               "standard 1.15.3 linux-glibc matches",
			reference:          "standard:1.15.3/linux-glibc",
			want:               "standard:1.15.3/linux-glibc",
			responseStatusCode: http.StatusOK,
		},
		{
			name:               "@ uses env var",
			reference:          "@",
			envVar:             "standard:1.15.3/linux-glibc",
			want:               "standard:1.15.3/linux-glibc",
			responseStatusCode: http.StatusOK,
		},
		{
			name:               "standard 1.15.3 matches",
			reference:          "standard:1.15.3",
			want:               fmt.Sprintf("standard:1.15.3/%v", platform()),
			responseStatusCode: http.StatusOK,
		},
		{
			name:               "standard-fips1402:1.10.0/linux-glibc matches",
			reference:          "standard-fips1402:1.10.0/linux-glibc",
			want:               "standard-fips1402:1.10.0/linux-glibc",
			responseStatusCode: http.StatusOK,
		},
		{
			name:               "sTanDard:nIgHTLY/LiNuX-gLiBc matches",
			reference:          "sTanDard:nIgHTLY/LiNuX-gLiBc",
			want:               "standard:nightly/linux-glibc",
			responseStatusCode: http.StatusOK,
		},
		{
			name:               "Error if not found",
			reference:          "notaFlavor:1.15.3/notaPlatform",
			responseStatusCode: http.StatusOK,
			wantErr:            true,
		},
		{
			name:      "Error if passed nil key",
			reference: "notAReference",
			wantErr:   true,
		},
		{
			name:               "Error on failed fetch",
			reference:          "standard:1.15.3",
			responseStatusCode: http.StatusTeapot,
			wantErr:            true,
		},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			mock := mockServer(tc.responseStatusCode, "manifest.golden")
			defer mock.Close()
			location := mock.URL
			if tc.locationOverride != "" {
				location = tc.locationOverride
			}
			defer func(originalURL string) {
				err := SetURL(originalURL)
				assert.NoError(t, err)
			}(GetURL())
			err := SetURL(location)
			assert.NoError(t, err)
			if len(tc.envVar) > 0 {
				os.Setenv(referenceEnv, tc.envVar)
				defer os.Unsetenv(referenceEnv)
			}
			key, _ := NewKey(tc.reference)
			if got, err := Locate(key); tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, "", got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
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
			got, err := NewKey(tc.reference)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}
