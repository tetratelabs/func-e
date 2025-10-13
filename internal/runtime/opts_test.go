// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package runtime_test

import (
	"os/user"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/runtime"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestInitializeGlobalOpts(t *testing.T) {
	u, err := user.Current()
	require.NoError(t, err)

	alt1 := filepath.Join(u.HomeDir, "alt1")
	defaultConfigHome := filepath.Join(u.HomeDir, ".config/func-e")
	defaultDataHome := filepath.Join(u.HomeDir, ".local/share/func-e")
	defaultStateHome := filepath.Join(u.HomeDir, ".local/state/func-e")
	defaultRuntimeDir := "/tmp/func-e-" + u.Uid
	defaultPlatform := globals.DefaultPlatform
	defaultVersionsURL := globals.DefaultEnvoyVersionsURL

	tests := []struct {
		name             string
		envoyVersionsURL string
		homeDir          string
		configHome       string
		dataHome         string
		stateHome        string
		runtimeDir       string
		platform         string
		runID            string
		expected         globals.GlobalOpts
		expectedErr      string
	}{
		{
			name:             "--envoy-versions-url not a URL",
			envoyVersionsURL: "/not/url",
			expectedErr:      `"/not/url" is not a valid Envoy versions URL`,
		},
		{
			name:     "default directories",
			expected: globals.GlobalOpts{ConfigHome: defaultConfigHome, DataHome: defaultDataHome, StateHome: defaultStateHome, RuntimeDir: defaultRuntimeDir, Platform: defaultPlatform, EnvoyVersionsURL: defaultVersionsURL},
		},
		{
			name:     "--home-dir arg (legacy mode)",
			homeDir:  alt1,
			expected: globals.GlobalOpts{ConfigHome: alt1, DataHome: alt1, StateHome: alt1, RuntimeDir: alt1, Platform: defaultPlatform, EnvoyVersionsURL: defaultVersionsURL},
		},
		{
			name:     "--platform flag",
			platform: "darwin/amd64",
			expected: globals.GlobalOpts{ConfigHome: defaultConfigHome, DataHome: defaultDataHome, StateHome: defaultStateHome, RuntimeDir: defaultRuntimeDir, Platform: version.Platform("darwin/amd64"), EnvoyVersionsURL: defaultVersionsURL},
		},
		{
			name:             "--envoy-versions-url flag",
			envoyVersionsURL: "http://versions/arg",
			expected:         globals.GlobalOpts{ConfigHome: defaultConfigHome, DataHome: defaultDataHome, StateHome: defaultStateHome, RuntimeDir: defaultRuntimeDir, Platform: defaultPlatform, EnvoyVersionsURL: "http://versions/arg"},
		},
		{
			name:       "--config-home only",
			configHome: alt1,
			expected:   globals.GlobalOpts{ConfigHome: alt1, DataHome: defaultDataHome, StateHome: defaultStateHome, RuntimeDir: defaultRuntimeDir, Platform: defaultPlatform, EnvoyVersionsURL: defaultVersionsURL},
		},
		{
			name:     "--data-home only",
			dataHome: alt1,
			expected: globals.GlobalOpts{ConfigHome: defaultConfigHome, DataHome: alt1, StateHome: defaultStateHome, RuntimeDir: defaultRuntimeDir, Platform: defaultPlatform, EnvoyVersionsURL: defaultVersionsURL},
		},
		{
			name:      "--state-home only",
			stateHome: alt1,
			expected:  globals.GlobalOpts{ConfigHome: defaultConfigHome, DataHome: defaultDataHome, StateHome: alt1, RuntimeDir: defaultRuntimeDir, Platform: defaultPlatform, EnvoyVersionsURL: defaultVersionsURL},
		},
		{
			name:       "--runtime-dir only",
			runtimeDir: alt1,
			expected:   globals.GlobalOpts{ConfigHome: defaultConfigHome, DataHome: defaultDataHome, StateHome: defaultStateHome, RuntimeDir: alt1, Platform: defaultPlatform, EnvoyVersionsURL: defaultVersionsURL},
		},
		{
			name:       "all directories specified",
			configHome: filepath.Join(u.HomeDir, "config"),
			dataHome:   filepath.Join(u.HomeDir, "data"),
			stateHome:  filepath.Join(u.HomeDir, "state"),
			runtimeDir: filepath.Join(u.HomeDir, "runtime"),
			expected:   globals.GlobalOpts{ConfigHome: filepath.Join(u.HomeDir, "config"), DataHome: filepath.Join(u.HomeDir, "data"), StateHome: filepath.Join(u.HomeDir, "state"), RuntimeDir: filepath.Join(u.HomeDir, "runtime"), Platform: defaultPlatform, EnvoyVersionsURL: defaultVersionsURL},
		},
		{
			name:     "custom runID",
			runID:    "my-custom-run",
			expected: globals.GlobalOpts{ConfigHome: defaultConfigHome, DataHome: defaultDataHome, StateHome: defaultStateHome, RuntimeDir: defaultRuntimeDir, Platform: defaultPlatform, EnvoyVersionsURL: defaultVersionsURL},
		},
		{
			name:     "custom runID numeric (Docker use case)",
			runID:    "0",
			expected: globals.GlobalOpts{ConfigHome: defaultConfigHome, DataHome: defaultDataHome, StateHome: defaultStateHome, RuntimeDir: defaultRuntimeDir, Platform: defaultPlatform, EnvoyVersionsURL: defaultVersionsURL},
		},
		{
			name:        "custom runID with forward slash rejected",
			runID:       "invalid/path",
			expectedErr: `runID cannot contain path separators (/ or \): "invalid/path"`,
		},
		{
			name:        "custom runID with backslash rejected",
			runID:       `invalid\path`,
			expectedErr: `runID cannot contain path separators (/ or \): "invalid\\path"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := &globals.GlobalOpts{}
			err := runtime.InitializeGlobalOpts(o, tc.envoyVersionsURL, tc.homeDir, tc.configHome, tc.dataHome, tc.stateHome, tc.runtimeDir, tc.platform, tc.runID)

			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tc.expected.ConfigHome, o.ConfigHome)
			require.Equal(t, tc.expected.DataHome, o.DataHome)
			require.Equal(t, tc.expected.StateHome, o.StateHome)
			require.Equal(t, tc.expected.RuntimeDir, o.RuntimeDir)
			require.Equal(t, tc.expected.Platform, o.Platform)
			require.Equal(t, tc.expected.EnvoyVersionsURL, o.EnvoyVersionsURL)

			if tc.runID != "" {
				require.Equal(t, tc.runID, o.RunID, "Custom RunID should be used")
			} else {
				require.NotEmpty(t, o.RunID, "RunID should be auto-generated when empty")
			}
		})
	}
}
