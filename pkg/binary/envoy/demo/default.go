// Copyright 2019 Tetrate
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

package demo

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/tetratelabs/getenvoy/pkg/binary"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
)

// StaticBootstrap can be used when there is no provided bootstrap
// It uses a basic front proxy bootstrap that directs between Google and Bing based on authority/host header
func StaticBootstrap(r *envoy.Runtime) {
	r.RegisterPreStart(defaultBootstrap)
}

func defaultBootstrap(r binary.Runner) error {
	path := filepath.Join(r.DebugStore(), "bootstrap.yaml")
	if err := ioutil.WriteFile(path, []byte(defaultFrontProxy), 0644); err != nil {
		return fmt.Errorf("error creating a default bootstrap: %v", err)
	}
	r.AppendArgs([]string{"--config-path", path})
	return nil
}

// HasBootstrapArg returns true if any of the Envoy config flags are present
func HasBootstrapArg(args []string) bool {
	for _, arg := range args {
		if arg == "-c" || arg == "--config-path" || arg == "--config-yaml" {
			return true
		}
	}
	return false
}

// Instructions tell users how to make HTTP request to demonstrate the demo bootstrap
var Instructions = `
The demo boostrap runs Envoy as a basic front proxy to Google/Bing

To make a request to Google:
` + "`curl -s -o /dev/null -vvv -H 'Host: google.com' localhost:15001/`" + `

To make a request to Bing:
` + "`curl -s -o /dev/null -vvv -H 'Host: bing.com' localhost:15001/`" + `

Check the access logs below to see the requests being made.
`

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
