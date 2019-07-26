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
		key              *Key
		locationOverride string
		want             string
		wantErr          bool
	}{
		{
			name: "standard 1.11.0 debian matches",
			key:  &Key{"standard", "1.11.0", "debian"},
			want: "standard:1.11.0/debian",
		},
		{
			name: "standard-fips1402 1.10.0 debian matches",
			key:  &Key{"standard-fips1402", "1.10.0", "debian"},
			want: "standard-fips1402:1.10.0/debian",
		},
		{
			name: "sTanDard nIgHTLY rHeL matches",
			key:  &Key{"sTanDard", "nIgHTLY", "rHeL"},
			want: "standard:nightly/rhel",
		},
		{
			name:    "Error if not found",
			key:     &Key{"notaFlavor", "1.11.0", "notanOS"},
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

func TestNewKey(t *testing.T) {
	tests := []struct {
		reference string
		want      *Key
		wantErr   bool
	}{
		{"flavor:version/family", &Key{Flavor: "flavor", Version: "version", OperatingSystemFamily: "family"}, false},
		{"flavor:version/", nil, true},
		{"flavor:version", nil, true},
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
