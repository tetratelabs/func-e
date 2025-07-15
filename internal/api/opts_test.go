// Copyright 2025 Tetrate
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

package api_test

import (
	"os/user"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/api"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestInitializeGlobalOpts(t *testing.T) {
	u, err := user.Current()
	require.NoError(t, err)

	alt1 := filepath.Join(u.HomeDir, "alt1")
	defaultHome := filepath.Join(u.HomeDir, ".func-e")
	defaultPlatform := globals.DefaultPlatform
	defaultVersionsURL := globals.DefaultEnvoyVersionsURL

	tests := []struct {
		name             string
		envoyVersionsURL string
		homeDir          string
		platform         string
		expected         globals.GlobalOpts
		expectedErr      string
	}{
		{
			name:             "--envoy-versions-url not a URL",
			envoyVersionsURL: "/not/url",
			expectedErr:      `"/not/url" is not a valid Envoy versions URL`,
		},
		{
			name:     "default is ~/.func-e",
			expected: globals.GlobalOpts{HomeDir: defaultHome, Platform: defaultPlatform, EnvoyVersionsURL: defaultVersionsURL},
		},
		{
			name:     "--home-dir arg",
			homeDir:  alt1,
			expected: globals.GlobalOpts{HomeDir: alt1, Platform: defaultPlatform, EnvoyVersionsURL: defaultVersionsURL},
		},
		{
			name:     "--platform flag",
			platform: "darwin/amd64",
			expected: globals.GlobalOpts{HomeDir: defaultHome, Platform: version.Platform("darwin/amd64"), EnvoyVersionsURL: defaultVersionsURL},
		},
		{
			name:             "--envoy-versions-url flag",
			envoyVersionsURL: "http://versions/arg",
			expected:         globals.GlobalOpts{HomeDir: defaultHome, Platform: defaultPlatform, EnvoyVersionsURL: "http://versions/arg"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := &globals.GlobalOpts{}
			err := api.InitializeGlobalOpts(o, tc.envoyVersionsURL, tc.homeDir, tc.platform)

			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tc.expected.HomeDir, o.HomeDir)
			require.Equal(t, tc.expected.Platform, o.Platform)
			require.Equal(t, tc.expected.EnvoyVersionsURL, o.EnvoyVersionsURL)
		})
	}
}
