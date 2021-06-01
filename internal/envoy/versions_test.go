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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/version"
)

func TestAddVersions(t *testing.T) {
	goodVersions := map[string]version.EnvoyVersion{
		"1.14.7": {
			ReleaseDate: "2021-04-15",
			Tarballs: map[string]string{
				"darwin/amd64": "https://getenvoy.io/versions/1.14.7/envoy-1.14.7-darwin-x86_64.tar.gz",
				"linux/amd64":  "https://getenvoy.io/versions/1.14.7/envoy-1.14.7-linux-x86_64.tar.gz",
			},
		},
		"1.17.3": {
			ReleaseDate: "2021-05-11",
			Tarballs: map[string]string{
				"linux/amd64": "https://getenvoy.io/versions/1.17.3/envoy-1.17.3-linux-x86_64.tar.gz",
			},
		},
		"1.18.3": {
			ReleaseDate: "2021-05-11",
			Tarballs: map[string]string{
				"darwin/amd64": "https://getenvoy.io/versions/1.18.3/envoy-1.18.3-darwin-x86_64.tar.gz",
				"linux/amd64":  "https://getenvoy.io/versions/1.18.3/envoy-1.18.3-linux-x86_64.tar.gz",
			},
		},
	}

	tests := []struct {
		name     string
		out      map[string]string
		update   map[string]version.EnvoyVersion
		platform string
		expected map[string]string
	}{
		{
			name:     "darwin",
			out:      map[string]string{},
			update:   goodVersions,
			platform: "darwin/amd64",
			expected: map[string]string{"1.14.7": "2021-04-15", "1.18.3": "2021-05-11"},
		},
		{
			name:     "linux",
			platform: "linux/amd64",
			out:      map[string]string{},
			update:   goodVersions,
			expected: map[string]string{"1.14.7": "2021-04-15", "1.17.3": "2021-05-11", "1.18.3": "2021-05-11"},
		},
		{
			name:     "already exists",
			out:      map[string]string{"1.14.7": "2020-01-01"},
			update:   goodVersions,
			platform: "darwin/amd64",
			expected: map[string]string{"1.14.7": "2020-01-01", "1.18.3": "2021-05-11"},
		},
		{
			name:     "unsupported OS",
			out:      map[string]string{},
			update:   goodVersions,
			platform: "windows/amd64",
			expected: map[string]string{},
		},
		{
			name:     "unsupported Arch",
			out:      map[string]string{},
			update:   goodVersions,
			platform: "linux/arm64",
			expected: map[string]string{},
		},
		{
			name:     "empty version list",
			out:      map[string]string{},
			platform: "darwin/amd64",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, AddVersions(tc.out, tc.update, tc.platform))
			require.Equal(t, tc.expected, tc.out)
		})
	}
}

func TestAddVersions_Validates(t *testing.T) {
	tests := []struct {
		name   string
		update map[string]version.EnvoyVersion
	}{
		{
			name: "invalid releaseDate",
			update: map[string]version.EnvoyVersion{
				"1.14.7": {
					ReleaseDate: "ice cream",
					Tarballs: map[string]string{
						"darwin/amd64": "https://getenvoy.io/versions/1.14.7/envoy-1.14.7-darwin-x86_64.tar.gz",
					},
				},
			},
		},
		{
			name: "missing releaseDate",
			update: map[string]version.EnvoyVersion{
				"1.14.7": {
					Tarballs: map[string]string{
						"darwin/amd64": "https://getenvoy.io/versions/1.14.7/envoy-1.14.7-darwin-x86_64.tar.gz",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			err := AddVersions(map[string]string{}, tc.update, "darwin/amd64")
			require.Error(t, err)
			require.Contains(t, err.Error(), `invalid releaseDate of version "1.14.7" for platform "darwin/amd64":`)
		})
	}
}
