// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"bytes"
	"io"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/admin"
	rootcmd "github.com/tetratelabs/func-e/internal/cmd"
	"github.com/tetratelabs/func-e/internal/envoy"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/test"
	"github.com/tetratelabs/func-e/internal/test/httptest"
	"github.com/tetratelabs/func-e/internal/version"
)

const deprecationWarning = "WARNING: $FUNC_E_HOME (--home-dir) is deprecated and will be removed in a future version.\n" +
	"Please use --config-home, --data-home, --state-home or --runtime-dir instead.\n"

func TestFuncEValidateArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedErr string
	}{
		{
			name:        "--envoy-versions-url not a URL",
			args:        []string{"--envoy-versions-url", "/not/url", "which"},
			expectedErr: `"/not/url" is not a valid Envoy versions URL`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := runTestCommand(t, &globals.GlobalOpts{}, tc.args)
			require.EqualError(t, err, tc.expectedErr)
		})
	}
}

func TestHomeDir(t *testing.T) {
	type testCase struct {
		name           string
		args           []string
		setup          func()
		expected       string
		expectedStderr string
	}

	u, err := user.Current()
	require.NoError(t, err)

	alt1 := filepath.Join(u.HomeDir, "alt1")
	alt2 := filepath.Join(u.HomeDir, "alt2")

	tests := []testCase{ // we don't test default as that depends on the runtime env
		{
			name:     "default",
			args:     []string{"which"},
			expected: filepath.Join(u.HomeDir, ".local", "share", "func-e"),
		},
		{
			name: "FUNC_E_HOME env (legacy mode)",
			args: []string{"which"},
			setup: func() {
				t.Setenv("FUNC_E_HOME", alt1)
			},
			expected:       alt1,
			expectedStderr: deprecationWarning,
		},
		{
			name:           "--home-dir arg (legacy mode)",
			args:           []string{"--home-dir", alt1, "which"},
			expected:       alt1,
			expectedStderr: deprecationWarning,
		},
		{
			name: "prioritizes --home-dir arg over FUNC_E_HOME env",
			args: []string{"--home-dir", alt1, "which"},
			setup: func() {
				t.Setenv("FUNC_E_HOME", alt2)
			},
			expected:       alt1,
			expectedStderr: deprecationWarning,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}

			o := &globals.GlobalOpts{}
			stdout := new(bytes.Buffer)
			stderr := new(bytes.Buffer)

			err := rootcmd.ParseFlags(stdout, stderr, tc.args, o)
			require.NoError(t, err)
			require.Equal(t, tc.expected, o.DataHome)
			require.Equal(t, tc.expectedStderr, stderr.String())
		})
	}
}

func TestDataHome(t *testing.T) {
	u, err := user.Current()
	require.NoError(t, err)

	testDirConfig(t, dirConfigTest{
		envVar:      "FUNC_E_DATA_HOME",
		flag:        "--data-home",
		suffix:      "data",
		defaultPath: filepath.Join(u.HomeDir, ".local", "share", "func-e"),
		accessor:    func(o *globals.GlobalOpts) string { return o.DataHome },
	})
}

func TestStateHome(t *testing.T) {
	u, err := user.Current()
	require.NoError(t, err)

	testDirConfig(t, dirConfigTest{
		envVar:      "FUNC_E_STATE_HOME",
		flag:        "--state-home",
		suffix:      "state",
		defaultPath: filepath.Join(u.HomeDir, ".local", "state", "func-e"),
		accessor:    func(o *globals.GlobalOpts) string { return o.StateHome },
	})
}

type dirConfigTest struct {
	envVar      string
	flag        string
	suffix      string
	defaultPath string
	accessor    func(*globals.GlobalOpts) string
}

func testDirConfig(t *testing.T, cfg dirConfigTest) {
	t.Helper()
	type testCase struct {
		name     string
		args     []string
		env      map[string]string
		expected string
	}

	u, err := user.Current()
	require.NoError(t, err)

	alt1 := filepath.Join(u.HomeDir, "alt-"+cfg.suffix)
	alt2 := filepath.Join(u.HomeDir, "alt-"+cfg.suffix+"-2")

	tests := []testCase{
		{
			name:     "default",
			args:     []string{"which"},
			expected: cfg.defaultPath,
		},
		{
			name: cfg.envVar + " env",
			args: []string{"which"},
			env: map[string]string{
				cfg.envVar: alt1,
			},
			expected: alt1,
		},
		{
			name:     cfg.flag + " flag",
			args:     []string{cfg.flag, alt1, "which"},
			expected: alt1,
		},
		{
			name: "prioritizes " + cfg.flag + " arg over " + cfg.envVar + " env",
			args: []string{cfg.flag, alt1, "which"},
			env: map[string]string{
				cfg.envVar: alt2,
			},
			expected: alt1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			o := &globals.GlobalOpts{}
			err := runTestCommand(t, o, tc.args)
			require.NoError(t, err)
			require.Equal(t, tc.expected, cfg.accessor(o))
		})
	}
}

func TestRuntimeDir(t *testing.T) {
	type testCase struct {
		name     string
		args     []string
		env      map[string]string
		expected string
	}

	alt1 := "/tmp/alt-runtime"
	alt2 := "/tmp/alt-runtime-2"

	u, err := user.Current()
	require.NoError(t, err)

	tests := []testCase{
		{
			name:     "default is /tmp/func-e-${UID}",
			args:     []string{"which"},
			expected: "/tmp/func-e-" + u.Uid,
		},
		{
			name: "FUNC_E_RUNTIME_DIR env",
			args: []string{"which"},
			env: map[string]string{
				"FUNC_E_RUNTIME_DIR": alt1,
			},
			expected: alt1,
		},
		{
			name:     "--runtime-dir flag",
			args:     []string{"--runtime-dir", alt1, "which"},
			expected: alt1,
		},
		{
			name: "prioritizes --runtime-dir arg over FUNC_E_RUNTIME_DIR env",
			args: []string{"--runtime-dir", alt1, "which"},
			env: map[string]string{
				"FUNC_E_RUNTIME_DIR": alt2,
			},
			expected: alt1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			o := &globals.GlobalOpts{}
			err := runTestCommand(t, o, tc.args)
			require.NoError(t, err)
			require.Equal(t, tc.expected, o.RuntimeDir)
		})
	}
}

func TestPlatformArg(t *testing.T) {
	type testCase struct {
		name     string
		args     []string
		env      map[string]string
		expected version.Platform
	}

	tests := []testCase{
		{
			name: "FUNC_E_PLATFORM env",
			args: []string{"which"},
			env: map[string]string{
				"FUNC_E_PLATFORM": "linux/amd64",
			},
			expected: version.Platform("linux/amd64"),
		},
		{
			name:     "--platform flag",
			args:     []string{"--platform", "darwin/amd64", "which"},
			expected: version.Platform("darwin/amd64"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			o := &globals.GlobalOpts{}
			err := runTestCommand(t, o, tc.args)
			require.NoError(t, err)
			require.Equal(t, tc.expected, o.Platform)
		})
	}
}

func TestEnvoyVersionsURL(t *testing.T) {
	type testCase struct {
		name     string
		args     []string
		env      map[string]string
		expected string
	}

	tests := []testCase{ // we don't test default as that depends on the runtime env
		{
			name:     "default is https://archive.tetratelabs.io/envoy/envoy-versions.json",
			args:     []string{"which"},
			expected: "https://archive.tetratelabs.io/envoy/envoy-versions.json",
		},
		{
			name: "ENVOY_VERSIONS_URL env",
			args: []string{"which"},
			env: map[string]string{
				"ENVOY_VERSIONS_URL": "http://ENVOY_VERSIONS_URL/env",
			},
			expected: "http://ENVOY_VERSIONS_URL/env",
		},
		{
			name:     "--envoy-versions-url flag",
			args:     []string{"--envoy-versions-url", "http://versions/arg", "which"},
			expected: "http://versions/arg",
		},
		{
			name: "prioritizes --envoy-versions-url arg over ENVOY_VERSIONS_URL env",
			args: []string{"--envoy-versions-url", "http://versions/arg", "which"},
			env: map[string]string{
				"ENVOY_VERSIONS_URL": "http://ENVOY_VERSIONS_URL/env",
			},
			expected: "http://versions/arg",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			o := &globals.GlobalOpts{}
			err := runTestCommand(t, o, tc.args)

			require.NoError(t, err)
			require.Equal(t, tc.expected, o.EnvoyVersionsURL)
		})
	}
}

// runTestCommand exercises global flag parsing without running any real command.
// The args must include a valid command name for kong to parse (e.g., "which" or "versions").
func runTestCommand(t *testing.T, o *globals.GlobalOpts, args []string) error {
	t.Helper()
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	return rootcmd.ParseFlags(stdout, stderr, args, o)
}

// setupTest returns globals.GlobalOpts with test defaults.
func setupTest(t *testing.T) *globals.GlobalOpts {
	t.Helper()
	result := globals.GlobalOpts{}
	result.EnvoyVersion = version.LastKnownEnvoy
	result.Platform = globals.DefaultPlatform
	result.Out = io.Discard
	result.EnvoyOut = io.Discard
	result.EnvoyErr = io.Discard

	result.ConfigHome = t.TempDir()
	result.DataHome = t.TempDir()
	result.StateHome = t.TempDir()
	result.RuntimeDir = t.TempDir()

	baseURL := "http://" + admin.ServerAddr
	handler := test.NewEnvoyVersionsHandler(t, baseURL, version.LastKnownEnvoy)
	result.HTTPClientFunc = httptest.HandlerFactory(handler)
	result.EnvoyVersionsURL = baseURL + "/envoy-versions.json"
	result.GetEnvoyVersions = envoy.NewGetVersions(result.HTTPClientFunc, result.EnvoyVersionsURL, result.UserAgent)
	return &result
}
