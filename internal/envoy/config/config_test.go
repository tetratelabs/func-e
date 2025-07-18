// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal"
)

func getTestDataPath() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Unable to determine current file path")
	}
	return filepath.Join(filepath.Dir(filename), "testdata")
}

func TestParseListeners(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Unable to determine current file path")
	}

	sourceDir := filepath.Dir(filename)
	adminLocalhostPath := filepath.Join(sourceDir, "testdata", "admin_localhost.yaml")
	adminEphemeralPath := filepath.Join(sourceDir, "testdata", "admin_ephemeral.yaml")
	noAdminPath := filepath.Join(sourceDir, "testdata", "no_admin.yaml")
	minimalPath := filepath.Join(sourceDir, "testdata", "minimal.yaml")
	staticFilePath := filepath.Join(sourceDir, "testdata", "static_file.yaml")
	udpProxyPath := filepath.Join(sourceDir, "testdata", "udp_proxy.yaml")

	tests := []struct {
		name      string
		args      []string
		expect    *Config
		expectErr string
	}{
		{
			name: "admin_localhost",
			args: []string{"-c", adminLocalhostPath},
			expect: &Config{
				Admin: "127.0.0.1:9901",
				StaticListeners: []Listener{{
					Name:    "test_listener",
					Address: "0.0.0.0:10000",
				}},
			},
		},
		{
			name: "admin_ephemeral",
			args: []string{"-c", adminEphemeralPath},
			expect: &Config{
				Admin: "127.0.0.1:0",
				StaticListeners: []Listener{{
					Name:    "test_listener",
					Address: "0.0.0.0:10000",
				}},
			},
		},
		{
			name: "no_admin",
			args: []string{"-c", noAdminPath},
			expect: &Config{
				Admin: "",
				StaticListeners: []Listener{{
					Name:    "test_listener",
					Address: "0.0.0.0:10000",
				}},
			},
		},
		{
			name: "minimal",
			args: []string{"-c", minimalPath},
			expect: &Config{
				Admin: "",
				StaticListeners: []Listener{{
					Name:    "main",
					Address: "127.0.0.1:0",
					Filters: []filterInfo{{
						Name:   "envoy.filters.network.http_connection_manager",
						Type:   "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager",
						Config: internal.MinimalTypedConfigYaml,
					}},
				}},
			},
		},
		{
			name: "static_file",
			args: []string{"-c", staticFilePath},
			expect: &Config{
				Admin: "",
				StaticListeners: []Listener{{
					Name:    "main",
					Address: "127.0.0.1:0",
					Filters: []filterInfo{{
						Name:   "envoy.http_connection_manager",
						Type:   "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager",
						Config: internal.StaticFileTypedConfigYaml,
					}},
				}},
			},
		},
		{
			name: "file_last_wins",
			args: []string{"-c", adminLocalhostPath, "-c", adminEphemeralPath},
			expect: &Config{
				Admin: "127.0.0.1:0",
				StaticListeners: []Listener{{
					Name:    "test_listener",
					Address: "0.0.0.0:10000",
				}},
			},
		},

		{
			name: "multiple_configs_last_has_admin",
			args: []string{
				"--config-yaml", `admin: {address: {socket_address: {address: "127.0.0.1", port_value: 9901}}}`,
				"--config-yaml", `admin: {address: {socket_address: {address: "127.0.0.2", port_value: 9902}}}`,
			},
			expect: &Config{
				Admin:           "127.0.0.2:9902",
				StaticListeners: []Listener{},
			},
		},
		{
			name: "multiple_configs_last_has_no_admin",
			args: []string{
				"--config-yaml", `admin: {address: {socket_address: {address: "127.0.0.1", port_value: 9901}}}`,
				"--config-yaml", `static_resources: {listeners: [{name: test_listener, address: {socket_address: {address: "127.0.0.1", port_value: 8080}}, filter_chains: [{filters: [{name: envoy.minimal}]}]}]}`,
			},
			expect: &Config{
				Admin: "127.0.0.1:9901",
				StaticListeners: []Listener{{
					Name:    "test_listener",
					Address: "127.0.0.1:8080",
					Filters: []filterInfo{{
						Name: "envoy.minimal",
					}},
				}},
			},
		},
		{
			name: "multiple_configs_first_has_no_admin",
			args: []string{
				"--config-yaml", `static_resources: {listeners: [{name: test_listener, address: {socket_address: {address: "127.0.0.1", port_value: 8080}}, filter_chains: [{filters: [{name: envoy.minimal}]}]}]}`,
				"--config-yaml", `admin: {address: {socket_address: {address: "127.0.0.2", port_value: 9902}}}`,
			},
			expect: &Config{
				Admin: "127.0.0.2:9902",
				StaticListeners: []Listener{{
					Name:    "test_listener",
					Address: "127.0.0.1:8080",
					Filters: []filterInfo{{
						Name: "envoy.minimal",
					}},
				}},
			},
		},
		{
			name: "no_admin_in_any",
			args: []string{
				"--config-yaml", `static_resources: {listeners: [{name: test_listener, address: {socket_address: {address: "127.0.0.1", port_value: 8080}}, filter_chains: [{filters: [{name: envoy.minimal}]}]}]}`,
				"--config-yaml", `static_resources: {clusters: [{name: test_cluster}]}`,
			},
			expect: &Config{
				Admin: "",
				StaticListeners: []Listener{{
					Name:    "test_listener",
					Address: "127.0.0.1:8080",
					Filters: []filterInfo{{
						Name: "envoy.minimal",
					}},
				}},
			},
		},
		{
			name:      "invalid_yaml",
			args:      []string{"--config-yaml", "invalid: {yaml"},
			expectErr: "failed to unmarshal YAML: yaml: line 1: did not find expected ',' or '}'",
		},
		{
			name:      "missing_value",
			args:      []string{"--config-yaml"},
			expectErr: "missing value for --config-yaml",
		},
		{
			name: "mixed_config_path_and_yaml_last_wins",
			args: []string{
				"-c", adminLocalhostPath,
				"--config-yaml", `admin: {address: {socket_address: {address: "127.0.0.3", port_value: 9903}}}`,
			},
			expect: &Config{
				Admin: "127.0.0.3:9903",
				StaticListeners: []Listener{{
					Name:    "test_listener",
					Address: "0.0.0.0:10000",
				}},
			},
		},
		{
			name: "mixed_config_path_and_yaml_file_wins",
			args: []string{
				"--config-yaml", `admin: {address: {socket_address: {address: "127.0.0.3", port_value: 9903}}}`,
				"-c", adminEphemeralPath,
			},
			expect: &Config{
				Admin: "127.0.0.1:0",
				StaticListeners: []Listener{{
					Name:    "test_listener",
					Address: "0.0.0.0:10000",
				}},
			},
		},
		{
			name: "admin_with_other_fields_ignored",
			args: []string{
				"--config-yaml", `admin: {address: {socket_address: {address: "127.0.0.1", port_value: 9901}}}`,
				"--config-yaml", `admin: {address: {socket_address: {address: "127.0.0.2", port_value: 9902}}, some_other_field: "ignored"}`,
			},
			expect: &Config{
				Admin:           "127.0.0.2:9902",
				StaticListeners: []Listener{},
			},
		},
		{
			name: "admin_partial_override_behavior",
			args: []string{
				"--config-yaml", `admin: {address: {socket_address: {address: "127.0.0.1", port_value: 9901}}}`,
				"--config-yaml", `admin: {address: {socket_address: {address: "127.0.0.2"}}}`,
			},
			expect: &Config{
				Admin:           "127.0.0.2:0",
				StaticListeners: []Listener{},
			},
		},
		{
			name: "udp_proxy",
			args: []string{"-c", udpProxyPath},
			expect: &Config{
				Admin: "",
				StaticListeners: []Listener{{
					Name:     "udp_listener",
					Address:  "127.0.0.1:10000",
					Protocol: "UDP",
					Filters: []filterInfo{{
						Name: "envoy.filters.udp_listener.udp_proxy",
						Type: "type.googleapis.com/envoy.extensions.filters.udp.udp_proxy.v3.UdpProxyConfig",
						Config: `'@type': type.googleapis.com/envoy.extensions.filters.udp.udp_proxy.v3.UdpProxyConfig
stat_prefix: foo
matcher:
    on_no_match:
        action:
            name: route
            typed_config:
                '@type': type.googleapis.com/envoy.extensions.filters.udp.udp_proxy.v3.Route
                cluster: cluster_0
`,
					}},
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseListeners(tt.args)
			if tt.expectErr != "" {
				require.EqualError(t, err, tt.expectErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expect, result)
			}
		})
	}
}

func TestFindAdminAddress(t *testing.T) {
	testdataDir := getTestDataPath()

	noAdminPath := filepath.Join(testdataDir, "no_admin.yaml")
	adminLocalhostPath := filepath.Join(testdataDir, "admin_localhost.yaml")

	tests := []struct {
		name      string
		args      []string
		expect    string
		expectErr string
	}{
		{
			name:   "file_with_admin",
			args:   []string{"-c", adminLocalhostPath},
			expect: "127.0.0.1:9901",
		},
		{
			name:   "file_without_admin",
			args:   []string{"-c", noAdminPath},
			expect: "",
		},
		{
			name: "config_with_admin",
			args: []string{
				"--config-yaml", `admin: {address: {socket_address: {address: "127.0.0.1", port_value: 9901}}}`,
			},
			expect: "127.0.0.1:9901",
		},
		{
			name: "config_without_admin",
			args: []string{
				"--config-yaml", `static_resources: {listeners: [{name: test_listener}]}`,
			},
			expect: "",
		},
		{
			name: "last_admin_wins",
			args: []string{
				"--config-yaml", `admin: {address: {socket_address: {address: "127.0.0.1", port_value: 9901}}}`,
				"--config-yaml", `admin: {address: {socket_address: {address: "127.0.0.2", port_value: 9902}}}`,
			},
			expect: "127.0.0.2:9902",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hostPort, err := FindAdminAddress(tt.args)
			if tt.expectErr != "" {
				require.EqualError(t, err, tt.expectErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expect, hostPort)
			}
		})
	}
}
