// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package globals

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEnvoyVersionsDir(t *testing.T) {
	tests := []struct {
		name     string
		dataHome string
		homeDir  string
		expected string
	}{
		{
			name:     "separate directories",
			dataHome: "/home/user/.local/share/func-e",
			homeDir:  "",
			expected: "/home/user/.local/share/func-e/envoy-versions",
		},
		{
			name:     "legacy mode",
			dataHome: "/home/user/func-e",
			homeDir:  "/home/user/func-e",
			expected: "/home/user/func-e/versions",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := &GlobalOpts{
				DataHome: tc.dataHome,
				HomeDir:  tc.homeDir,
			}
			actual := o.EnvoyVersionsDir()
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestEnvoyVersionFile(t *testing.T) {
	tests := []struct {
		name       string
		configHome string
		dataHome   string
		homeDir    string
		expected   string
	}{
		{
			name:       "separate directories",
			configHome: "/home/user/.config/func-e",
			dataHome:   "/home/user/.local/share/func-e",
			homeDir:    "",
			expected:   "/home/user/.config/func-e/envoy-version",
		},
		{
			name:       "legacy mode",
			configHome: "/home/user/func-e",
			dataHome:   "/home/user/func-e",
			homeDir:    "/home/user/func-e",
			expected:   "/home/user/func-e/version",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := &GlobalOpts{
				ConfigHome: tc.configHome,
				DataHome:   tc.dataHome,
				HomeDir:    tc.homeDir,
			}
			actual := o.EnvoyVersionFile()
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestEnvoyRunDir(t *testing.T) {
	tests := []struct {
		name      string
		stateHome string
		homeDir   string
		runID     string
		expected  string
	}{
		{
			name:      "separate directories",
			stateHome: "/home/user/.local/state/func-e",
			homeDir:   "",
			runID:     "2025-01-15T12:34:56",
			expected:  "/home/user/.local/state/func-e/envoy-runs/2025-01-15T12:34:56",
		},
		{
			name:      "legacy mode",
			stateHome: "/home/user/func-e",
			homeDir:   "/home/user/func-e",
			runID:     "1619574747231823000",
			expected:  "/home/user/func-e/runs/1619574747231823000",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := &GlobalOpts{
				StateHome: tc.stateHome,
				HomeDir:   tc.homeDir,
			}
			actual := o.EnvoyRunDir(tc.runID)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestEnvoyRuntimeDir(t *testing.T) {
	tests := []struct {
		name       string
		runtimeDir string
		homeDir    string
		runID      string
		expected   string
	}{
		{
			name:       "separate directories",
			runtimeDir: "/tmp/func-e-1000",
			homeDir:    "",
			runID:      "20250115_123456_000",
			expected:   filepath.Join("/tmp/func-e-1000", "20250115_123456_000"),
		},
		{
			name:       "legacy mode",
			runtimeDir: "/home/user/func-e",
			homeDir:    "/home/user/func-e",
			runID:      "1619574747231823000",
			expected:   filepath.Join("/home/user/func-e/runs", "1619574747231823000"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := &GlobalOpts{
				RuntimeDir: tc.runtimeDir,
				HomeDir:    tc.homeDir,
			}
			actual := o.EnvoyRuntimeDir(tc.runID)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestGenerateRunID(t *testing.T) {
	tests := []struct {
		name            string
		homeDir         string
		now             time.Time
		expectedPattern string
	}{
		{
			name:            "timestamp format",
			homeDir:         "",
			now:             time.Date(2025, 1, 15, 12, 34, 56, 789000000, time.UTC),
			expectedPattern: "20250115_123456_000",
		},
		{
			name:            "legacy mode - epoch nanoseconds",
			homeDir:         "/home/user/func-e",
			now:             time.Date(2021, 4, 27, 18, 39, 7, 231823000, time.UTC),
			expectedPattern: "1619548747231823000",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := &GlobalOpts{
				HomeDir: tc.homeDir,
			}
			actual := o.GenerateRunID(tc.now)
			require.Equal(t, tc.expectedPattern, actual)
		})
	}
}

func TestEnvoyVersionFileSource(t *testing.T) {
	tests := []struct {
		name     string
		homeDir  string
		expected string
	}{
		{
			name:     "separate directories",
			homeDir:  "",
			expected: "$FUNC_E_CONFIG_HOME/envoy-version",
		},
		{
			name:     "legacy mode",
			homeDir:  "/home/user/func-e",
			expected: "$FUNC_E_HOME/version",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			o := &GlobalOpts{
				HomeDir: tc.homeDir,
			}
			actual := o.EnvoyVersionFileSource()
			require.Equal(t, tc.expected, actual)
		})
	}
}
