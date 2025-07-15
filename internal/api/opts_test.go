// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

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
