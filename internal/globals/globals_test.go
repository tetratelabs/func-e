// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package globals

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGlobalOpts_Mkdirs(t *testing.T) {
	testCases := []struct {
		name          string
		initialDirs   map[string]fs.FileMode // directories to create before calling Mkdirs (empty means create none)
		runID         string                 // empty means no per-run directories
		expectedError string
		expectedDirs  map[string]fs.FileMode // directories that should exist with correct perms after Mkdirs
	}{
		{
			name:          "creates only base directories when no runID",
			initialDirs:   map[string]fs.FileMode{}, // nothing exists
			runID:         "",                       // no per-run directories
			expectedError: "",
			expectedDirs: map[string]fs.FileMode{
				"config":              0o750,
				"data":                0o750,
				"data/envoy-versions": 0o750,
				// StateHome, RuntimeDir NOT created when no runID
			},
		},
		{
			name: "creates missing directories when some exist",
			initialDirs: map[string]fs.FileMode{
				"config": 0o755,
				"data":   0o755,
			},
			runID:         "",
			expectedError: "",
			expectedDirs: map[string]fs.FileMode{
				"data/envoy-versions": 0o750,
			},
		},
		{
			name: "idempotent when all directories exist",
			initialDirs: map[string]fs.FileMode{
				"config":              0o755,
				"data":                0o755,
				"data/envoy-versions": 0o755,
			},
			runID:         "",
			expectedError: "",
			expectedDirs:  map[string]fs.FileMode{}, // All pre-exist, just verify no errors
		},
		{
			name:          "creates per-run directories when runID is set",
			initialDirs:   map[string]fs.FileMode{},
			runID:         "20250413_123045_001",
			expectedError: "",
			expectedDirs: map[string]fs.FileMode{
				"config":                               0o750,
				"data":                                 0o750,
				"data/envoy-versions":                  0o750,
				"state/envoy-runs/20250413_123045_001": 0o750, // Leaf per-run dir
				"runtime/20250413_123045_001":          0o700, // Leaf per-run dir
			},
		},
		{
			name: "creates per-run directories when some base directories exist",
			initialDirs: map[string]fs.FileMode{
				"config": 0o755,
			},
			runID:         "20250413_123045_002",
			expectedError: "",
			expectedDirs: map[string]fs.FileMode{
				"data":                                 0o750,
				"data/envoy-versions":                  0o750,
				"state/envoy-runs/20250413_123045_002": 0o750,
				"runtime/20250413_123045_002":          0o700,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Each test gets its own isolated temp directory
			baseDir := t.TempDir()

			configHome := filepath.Join(baseDir, "config")
			dataHome := filepath.Join(baseDir, "data")
			stateHome := filepath.Join(baseDir, "state")
			runtimeDir := filepath.Join(baseDir, "runtime")

			// Pre-create directories specified in test case
			for dir, perm := range tc.initialDirs {
				fullPath := filepath.Join(baseDir, dir)
				require.NoError(t, os.MkdirAll(fullPath, perm))
			}

			// Set up GlobalOpts
			o := &GlobalOpts{
				ConfigHome: configHome,
				DataHome:   dataHome,
				StateHome:  stateHome,
				RuntimeDir: runtimeDir,
			}

			// Set up per-run directories if runID is specified
			if tc.runID != "" {
				o.RunID = tc.runID
				o.RunDir = o.EnvoyRunDir(tc.runID)
				o.RuntimeDir = o.EnvoyRuntimeDir(tc.runID)
			}

			// Call Mkdirs
			err := o.Mkdirs()

			// Check error expectation
			if tc.expectedError != "" {
				require.Error(t, err)
				require.EqualError(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)

			// Verify expected directories exist with correct permissions
			for dir, expectedPerm := range tc.expectedDirs {
				fullPath := filepath.Join(baseDir, dir)
				require.DirExists(t, fullPath, "directory %q should exist", dir)

				info, err := os.Stat(fullPath)
				require.NoError(t, err, "should be able to stat directory %q", dir)
				actualPerm := info.Mode().Perm()
				require.Equal(t, expectedPerm, actualPerm, "directory %q should have permissions %o, got %o", dir, expectedPerm, actualPerm)
			}

			// Verify idempotency - calling Mkdirs again should not error
			err = o.Mkdirs()
			require.NoError(t, err, "Mkdirs should be idempotent")
		})
	}
}

func TestGlobalOpts_Mkdirs_RuntimeDirPermissions(t *testing.T) {
	// Specific test to verify XDG spec compliance: per-run RuntimeDir must be 0700
	baseDir := t.TempDir()

	o := &GlobalOpts{
		ConfigHome: filepath.Join(baseDir, "config"),
		DataHome:   filepath.Join(baseDir, "data"),
		StateHome:  filepath.Join(baseDir, "state"),
		RuntimeDir: filepath.Join(baseDir, "runtime"),
	}

	// Without runID set, RuntimeDir is NOT created
	require.NoError(t, o.Mkdirs())
	_, err := os.Stat(o.RuntimeDir)
	require.True(t, os.IsNotExist(err), "RuntimeDir should not exist when no runID set")

	// Verify base directories have 0750 permissions
	for _, dir := range []string{o.ConfigHome, o.DataHome} {
		info, err := os.Stat(dir)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0o750), info.Mode().Perm(), "directory %q should have 0750 permissions", dir)
	}
}

func TestGlobalOpts_Mkdirs_PerRunRuntimeDirPermissions(t *testing.T) {
	// Verify per-run RuntimeDir has 0700 permissions per XDG spec
	baseDir := t.TempDir()

	o := &GlobalOpts{
		ConfigHome: filepath.Join(baseDir, "config"),
		DataHome:   filepath.Join(baseDir, "data"),
		StateHome:  filepath.Join(baseDir, "state"),
		RuntimeDir: filepath.Join(baseDir, "runtime"),
	}

	runID := "20250413_123045_999"
	o.RunID = runID
	o.RunID = runID
	o.RunDir = o.EnvoyRunDir(runID)
	o.RuntimeDir = o.EnvoyRuntimeDir(runID)

	require.NoError(t, o.Mkdirs())

	// Verify per-run RuntimeDir has 0700 permissions (XDG spec requirement)
	info, err := os.Stat(o.RuntimeDir)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o700), info.Mode().Perm(), "per-run RuntimeDir must have 0700 permissions per XDG spec")

	// Verify per-run RunDir has 0750 permissions
	info, err = os.Stat(o.RunDir)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o750), info.Mode().Perm(), "per-run RunDir should have 0750 permissions")
}
