// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal"
	"github.com/tetratelabs/func-e/internal/admin"
)

func testDataPath(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to determine current file path")
	}
	return filepath.Join(filepath.Dir(filename), "testdata")
}

func TestParseListeners(t *testing.T) {
	testdataDir := testDataPath(t)
	adminLocalhostPath := filepath.Join(testdataDir, "admin_localhost.yaml")
	adminEphemeralPath := filepath.Join(testdataDir, "admin_ephemeral.yaml")
	noAdminPath := filepath.Join(testdataDir, "no_admin.yaml")
	accessLogPath := filepath.Join(testdataDir, "access_log.yaml")
	staticFilePath := filepath.Join(testdataDir, "static_file.yaml")
	udpProxyPath := filepath.Join(testdataDir, "udp_proxy.yaml")

	tests := []struct {
		name        string
		configPath  string
		configYaml  string
		expect      *Config
		expectedErr string
	}{
		{
			name:       "admin_localhost",
			configPath: adminLocalhostPath,
			expect: &Config{
				Admin: admin.ServerAddr,
				StaticListeners: []Listener{{
					Name:    "test_listener",
					Address: "0.0.0.0:10000",
				}},
			},
		},
		{
			name:       "admin_ephemeral",
			configPath: adminEphemeralPath,
			expect: &Config{
				Admin: "127.0.0.1:0",
				StaticListeners: []Listener{{
					Name:    "test_listener",
					Address: "0.0.0.0:10000",
				}},
			},
		},
		{
			name:       "no_admin",
			configPath: noAdminPath,
			expect: &Config{
				Admin: "",
				StaticListeners: []Listener{{
					Name:    "test_listener",
					Address: "0.0.0.0:10000",
				}},
			},
		},
		{
			name:       "access_log",
			configPath: accessLogPath,
			expect: &Config{
				Admin: "127.0.0.1:0",
				StaticListeners: []Listener{{
					Name:    "main",
					Address: "127.0.0.1:0",
					Filters: []filterInfo{{
						Name: "envoy.filters.network.http_connection_manager",
						Type: "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager",
						Config: `"@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
stat_prefix: ingress_http
access_log:
    - name: envoy.access_loggers.stdout
      typed_config:
        "@type": type.googleapis.com/envoy.extensions.access_loggers.stream.v3.StdoutAccessLog
route_config:
    name: local_route
    virtual_hosts:
        - name: direct_response_service
          domains: ["*"]
          routes:
            - match:
                prefix: "/"
              direct_response:
                status: 200
                body:
                    inline_string: "Hello, World!"
http_filters:
    - name: envoy.filters.http.router
      typed_config:
        "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
`,
					}},
				}},
			},
		},
		{
			name:       "static_file",
			configPath: staticFilePath,
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
			name:        "invalid_yaml",
			configYaml:  "invalid: {yaml",
			expectedErr: "failed to unmarshal YAML: yaml: line 1: did not find expected ',' or '}'",
		},
		{
			name:       "mixed_config_path_and_yaml_last_wins",
			configPath: adminLocalhostPath,
			configYaml: `admin: {address: {socket_address: {address: "127.0.0.3", port_value: 9903}}}`,
			expect: &Config{
				Admin: "127.0.0.3:9903",
				StaticListeners: []Listener{{
					Name:    "test_listener",
					Address: "0.0.0.0:10000",
				}},
			},
		},
		{
			name:       "mixed_config_path_and_yaml_yaml_always_wins",
			configPath: adminEphemeralPath,
			configYaml: `admin: {address: {socket_address: {address: "127.0.0.3", port_value: 9903}}}`,
			expect: &Config{
				Admin: "127.0.0.3:9903",
				StaticListeners: []Listener{{
					Name:    "test_listener",
					Address: "0.0.0.0:10000",
				}},
			},
		},
		{
			name:       "udp_proxy",
			configPath: udpProxyPath,
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
			result, err := ParseListeners(tt.configPath, tt.configYaml)
			if tt.expectedErr != "" {
				require.EqualError(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expect, result)
			}
		})
	}
}

func TestFindAdminAddress(t *testing.T) {
	testdataDir := testDataPath(t)

	noAdminPath := filepath.Join(testdataDir, "no_admin.yaml")
	adminLocalhostPath := filepath.Join(testdataDir, "admin_localhost.yaml")

	tests := []struct {
		name        string
		configPath  string
		configYaml  string
		expect      string
		expectedErr string
	}{
		{
			name:       "file_with_admin",
			configPath: adminLocalhostPath,
			expect:     admin.ServerAddr,
		},
		{
			name:       "file_without_admin",
			configPath: noAdminPath,
			expect:     "",
		},
		{
			name:       "config_with_admin",
			configYaml: `admin: {address: {socket_address: {address: "127.0.0.1", port_value: 9901}}}`,
			expect:     admin.ServerAddr,
		},
		{
			name:       "config_without_admin",
			configYaml: `static_resources: {listeners: [{name: test_listener}]}`,
			expect:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hostPort, err := FindAdminAddress(tt.configPath, tt.configYaml)
			if tt.expectedErr != "" {
				require.EqualError(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expect, hostPort)
			}
		})
	}
}

func TestFindAdminAddressFromArgs(t *testing.T) {
	testdataDir := testDataPath(t)

	adminLocalhostPath := filepath.Join(testdataDir, "admin_localhost.yaml")
	adminYaml := `admin: {address: {socket_address: {address: "127.0.0.3", port_value: 9903}}}`

	tests := []struct {
		name   string
		args   []string
		expect string
	}{
		{
			name:   "config path value",
			args:   []string{"--config-path", adminLocalhostPath},
			expect: admin.ServerAddr,
		},
		{
			name:   "config path equals",
			args:   []string{"--config-path=" + adminLocalhostPath},
			expect: admin.ServerAddr,
		},
		{
			name:   "config yaml value",
			args:   []string{"--config-yaml", adminYaml},
			expect: "127.0.0.3:9903",
		},
		{
			name:   "config yaml equals",
			args:   []string{"--config-yaml=" + adminYaml},
			expect: "127.0.0.3:9903",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hostPort, err := FindAdminAddressFromArgs(tt.args)
			require.NoError(t, err)
			require.Equal(t, tt.expect, hostPort)
		})
	}
}
