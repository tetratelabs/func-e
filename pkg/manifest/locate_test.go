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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocate(t *testing.T) {
	tests := []struct {
		name             string
		key              Key
		locationOverride string
		want             string
		wantErr          bool
	}{
		{
			name: "Ubuntu bionic standard 1.11.0 matches",
			key:  Key{"standard", "1.11.0", "Ubuntu", "bionic"},
			want: "standard:1.11.0/debian",
		},
		{
			name: "Ubuntu xenial standard-fips1402 1.10.0 matches",
			key:  Key{"standard-fips1402", "1.10.0", "Ubuntu", "xenial"},
			want: "standard-fips1402:1.10.0/debian",
		},
		{
			name: "CentOS 7.1 standard nightly matches to CentOS 7",
			key:  Key{"standard", "nightly", "centos", "7.1"},
			want: "standard:nightly/centos",
		},
		{
			name: "cEnTOS 7 sTanDard nIgHTLY matches",
			key:  Key{"sTanDard", "nIgHTLY", "cEnTOS", "7"},
			want: "standard:nightly/centos",
		},
		{
			name: "MacOS standard 1.11.0 with no OS version returns the only macos build",
			key:  Key{Flavor: "standard", Version: "1.11.0", OperatingSystem: "macos"},
			want: "standard:1.11.0/macos",
		},
		{
			name:    "Error if not found",
			key:     Key{"notaFlavor", "1.11.0", "Ubuntu", "bionic"},
			wantErr: true,
		},
		{
			name:             "Error on non-url manifest locations",
			locationOverride: "not-a-url",
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			mock := mockServer(http.StatusOK, "manifest.golden")
			defer mock.Close()
			location := mock.URL
			if tc.locationOverride != "" {
				location = tc.locationOverride
			}
			if got, err := Locate(tc.key, location); tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, "", got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}
