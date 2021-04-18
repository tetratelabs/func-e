// Copyright 2020 Tetrate
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

package postgres

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"text/template"

	valid "github.com/asaskevich/govalidator"
	"github.com/tetratelabs/log"

	"github.com/tetratelabs/getenvoy/pkg/flavors"
)

// Define template parameter names
const endpointsParam string = "endpoints"
const inportParam string = "inport"

// Flavor implements flavor.FlavorConfigTemplate interface
// and stores config data specific to Postgres template.
type Flavor struct {
	// Location of the postgres server
	endpoints []*clusterEndpoint
	// Envoy's listener port
	InPort      string
	ClusterType string
}

type clusterEndpoint struct {
	EndpointAddr string
	EndpointPort string
	IsIP         bool
}

var flavor = Flavor{
	InPort: "5432",
}

func init() {
	// Register postgres flavor.
	flavors.AddFlavor("postgres", &flavor)
}

// GenerateConfig method takes command line parameters and creates
// Postgres specific Envoy config.
func (f *Flavor) GenerateConfig(params map[string]string) (string, error) {
	var err error
	var config string

	err = f.parseInputParams(params)
	if err != nil {
		return "", err
	}
	// Now process the main template. This includes everything except endpoints.
	// NOw run the template substitution
	mainConfig, err := f.generateMainConfig()
	if err != nil {
		return "", err
	}
	config += mainConfig

	endpointsConfig, err := f.generateEndpointSetConfig()
	if err != nil {
		return "", err
	}
	config += endpointsConfig

	return config, nil
}

// Utility functions.

// Function parses and verifies inout params for consistency
// and correctness.
func (f *Flavor) parseInputParams(params map[string]string) error {
	var err error

	for param, value := range params {
		switch param {
		case endpointsParam:
			f.endpoints, err = parseEndpointSet(value)
			if err != nil {
				return err
			}
		case inportParam:
			if !valid.IsInt(value) {
				return fmt.Errorf("Value for templateArg %s must be integer number", param)
			}
			f.InPort = value
		default:
			log.Warnf("Ignoring unrecognized template parameter: %s", param)
		}
	}

	// Check if all required params have been found in the parameter list
	if len(f.endpoints) == 0 {
		return fmt.Errorf("Parameter endpoints must be specified")
	}

	// Verify that all endpoints are of the same type:
	// All are IP address or all are domain name.
	clusterType := f.endpoints[0].IsIP
	for _, singleEndpoint := range f.endpoints[1:] {
		if singleEndpoint.IsIP != clusterType {
			return fmt.Errorf("Endpoints must be of the same type type: IP addresses or hostnames")
		}
	}
	return nil
}

// Endpoint may have the following forms:
// IP address: like 127.0.0.1
// IP address and port: 127.0.0.1:5432
// domain name: postgres
// domain name and port: postgres:5432
func parseSingleEndpoint(endpoint string) (*clusterEndpoint, error) {
	var host, port string
	var err error
	parts := strings.Split(endpoint, ":")
	if len(parts) > 2 {
		return nil, fmt.Errorf("%s endpoint has incorrect format. Should be endpoint[:port]", endpoint)
	}

	if len(parts) == 2 {
		host, port, err = net.SplitHostPort(endpoint)
		if err != nil {
			return nil, fmt.Errorf("%s endpoint has incorrect format. Should be endpoint[:port]", endpoint)
		}
		if !valid.IsInt(port) {
			return nil, fmt.Errorf("Port value in endpoint:port must be integer not %s", port)
		}
	} else {
		host = parts[0]
		port = "5432"
	}

	// This is just IP address or domain name.
	isIP := net.ParseIP(host) != nil

	return &clusterEndpoint{
		EndpointAddr: host,
		EndpointPort: port,
		IsIP:         isIP}, nil

}

// Function parses a string of comma separated endpoints.
func parseEndpointSet(endpoints string) ([]*clusterEndpoint, error) {
	// endpoints is comma separated list of endpoints
	singleEndpoints := strings.Split(endpoints, ",")

	clusterEndpoints := make([]*clusterEndpoint, 0, len(singleEndpoints))

	for _, endpoint := range singleEndpoints {
		singleClusterEndpoint, err := parseSingleEndpoint(endpoint)
		if err != nil {
			return nil, err
		}
		clusterEndpoints = append(clusterEndpoints, singleClusterEndpoint)
	}

	return clusterEndpoints, nil
}

func (f *Flavor) generateEndpointSetConfig() (string, error) {
	var buf bytes.Buffer

	tmpl := template.New("postgres-endpoint")
	tmpl, err := tmpl.Parse(clusterEndpointTemplate)
	if err != nil {
		// Template is not supplied by a user, but is compiled-in, so this error should
		// happen only during development time.
		return "", fmt.Errorf("Cannot parse postgres endpoint template")
	}

	for _, singleEndpoint := range f.endpoints {
		err = tmpl.Execute(&buf, singleEndpoint)
		if err != nil {
			return "", fmt.Errorf("Cannot execute postgres endpoint template")
		}
	}

	return buf.String(), nil
}

func (f *Flavor) generateMainConfig() (string, error) {
	var buf bytes.Buffer

	if f.endpoints[0].IsIP {
		f.ClusterType = "static"
	} else {
		f.ClusterType = "strict_dns"
	}
	tmpl := template.New("postgres-main")
	tmpl, err := tmpl.Parse(configTemplate)
	if err != nil {
		// Template is not supplied by a user, but is compiled-in, so this error should
		// happen only during development time.
		return "", fmt.Errorf("Cannot parse postgres main template")
	}
	err = tmpl.Execute(&buf, f)
	if err != nil {
		return "", fmt.Errorf("Cannot execute postgres main template")
	}

	return buf.String(), nil
}

// Postgres specific config file.
var configTemplate = `
admin:
  access_log_path: "/dev/null"
  address:
    socket_address:
      address: 0.0.0.0
      port_value: 0

static_resources:
  listeners:
  - name: postgres_listener
    address:
      socket_address:
        address: 0.0.0.0
        port_value: {{ .InPort }}
    filter_chains:
    - filters:
      - name: envoy.filters.network.postgres_proxy
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.postgres_proxy.v3alpha.PostgresProxy
          stat_prefix: egress_postgres
      - name: envoy.tcp_proxy
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
          stat_prefix: postgres_tcp
          cluster: postgres_cluster

  clusters:
  - name: postgres_cluster
    connect_timeout: 1s
    type: {{ .ClusterType }}
    load_assignment:
      cluster_name: postgres_cluster
      endpoints:
      - lb_endpoints:`

var clusterEndpointTemplate = `
        - endpoint:
            address:
              socket_address:
                address: {{ .EndpointAddr }} 
                port_value: {{ .EndpointPort }}`
