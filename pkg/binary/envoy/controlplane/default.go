package controlplane

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/tetratelabs/getenvoy/pkg/binary"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
)

// DefaultStaticBootstrap can be used when there is no provided bootstrap
// It uses a basic front proxy bootstrap that directs between Google and Bing based on authority/host header
func DefaultStaticBootstrap(r *envoy.Runtime) {
	r.RegisterPreStart(defaultBootstrap)
}

func defaultBootstrap(r binary.Runner) error {
	path := filepath.Join(r.DebugStore(), "bootstrap.yaml")
	if err := ioutil.WriteFile(path, []byte(defaultFrontProxy), 0644); err != nil {
		return fmt.Errorf("error creating a default bootstrap: %v", err)
	}
	r.AppendArgs([]string{"-c", path})
	return nil
}

func HasBootstrapArg(args []string) bool {
	for _, arg := range args {
		if arg == "-c" || arg == "--config-path" || arg == "--config-yaml" {
			return true
		}
	}
	return false
}

var defaultFrontProxy = `
static_resources:
  listeners:
  - address:
      # Tells Envoy to listen on 0.0.0.0:15001
      socket_address:
        address: 0.0.0.0
        port_value: 15001
    filter_chains:
    # Any requests received on this address are sent through this chain of filters
    - filters:
      # If the request is HTTP it will pass through this HTTP filter
      - name: envoy.http_connection_manager 
        typed_config:
          "@type": type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager
          codec_type: auto
          stat_prefix: http
          access_log:
            name: envoy.file_access_log
            typed_config:
              "@type": type.googleapis.com/envoy.config.accesslog.v2.FileAccessLog
              path: /dev/stdout
          route_config:
            name: search_route
            virtual_hosts:
            - name: backend
              domains:
              - "*"
              routes:
              # Match on host (:authority in HTTP2) headers
              - match:
                  prefix: "/"
                  headers:
                    - name: ":authority"
                      exact_match: "google.com"
                route:
                  # Send request to an endpoint in the Google cluster
                  cluster: google
                  host_rewrite: www.google.com
              - match:
                  prefix: "/"
                  headers:
                    - name: ":authority"
                      exact_match: "bing.com"
                route:
                  # Send request to an endpoint in the Bing cluster
                  cluster: bing
                  host_rewrite: www.bing.com
          http_filters:
          - name: envoy.router
            typed_config: {}
  clusters:
  - name: google
    connect_timeout: 1s
    # Instruct Envoy to continouously resolve DNS of www.google.com asynchronously
    type: logical_dns 
    dns_lookup_family: V4_ONLY
    lb_policy: round_robin
    load_assignment:
      cluster_name: google
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: www.google.com
                port_value: 80
  - name: bing
    connect_timeout: 1s
    # Instruct Envoy to continouously resolve DNS of www.bing.com asynchronously
    type: logical_dns
    dns_lookup_family: V4_ONLY
    lb_policy: round_robin
    load_assignment:
      cluster_name: bing
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: www.bing.com
                port_value: 80
admin:
  access_log_path: "/dev/stdout"
  address:
    socket_address:
      address: 0.0.0.0
      port_value: 15000
`
