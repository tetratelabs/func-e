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

package wasm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegistryHosts(t *testing.T) {
	tests := []struct {
		name           string
		insecure       bool
		plainHTTP      bool
		inputHost      string
		expectedScheme string
		expectedHost   string
		expectedPath   string
	}{
		{
			name:           "docker.io",
			insecure:       false,
			plainHTTP:      false,
			inputHost:      "docker.io",
			expectedScheme: "https",
			expectedHost:   "docker.io",
			expectedPath:   "/v2",
		},
		{
			name:           "localhost:5000",
			insecure:       false,
			plainHTTP:      false,
			inputHost:      "localhost:5000",
			expectedScheme: "https",
			expectedHost:   "localhost:5000",
			expectedPath:   "/v2",
		},
		{
			name:           "localhost:5000,plainHTTP=true",
			insecure:       false,
			plainHTTP:      true,
			inputHost:      "localhost:5000",
			expectedScheme: "http",
			expectedHost:   "localhost:5000",
			expectedPath:   "/v2",
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			hosts, err := registryHosts(tc.insecure, tc.plainHTTP)(tc.inputHost)
			require.NoError(t, err)
			require.Equal(t, 1, len(hosts))
			host := hosts[0]

			require.Equal(t, tc.expectedScheme, host.Scheme)
			require.Equal(t, tc.expectedHost, host.Host)
			require.Equal(t, tc.expectedPath, host.Path)
		})
	}
}
