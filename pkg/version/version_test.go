// Copyright 2020 Tetrate
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

package version

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "ad-hoc build",
			version:  "",
			expected: "dev",
		},
		{
			name:     "dev build",
			version:  "dev",
			expected: "dev",
		},
		{
			name:     "release build",
			version:  "0.0.1",
			expected: "0.0.1",
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			previous := version
			defer func() {
				version = previous
			}()

			version = test.version
			require.Equal(t, test.expected, versionOrDefault())
		})
	}
}

func TestIsDevBuild(t *testing.T) {
	tests := []struct {
		name     string
		build    BuildInfo
		expected bool
	}{
		{
			name:     "dev build",
			build:    BuildInfo{Version: "dev"},
			expected: true,
		},
		{
			name:     "release build",
			build:    BuildInfo{Version: "0.0.1"},
			expected: false,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			previous := Build
			defer func() {
				Build = previous
			}()

			Build = test.build
			require.Equal(t, test.expected, IsDevBuild())
		})
	}
}
