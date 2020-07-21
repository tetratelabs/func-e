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
	"fmt"

	valid "github.com/asaskevich/govalidator"
	"github.com/tetratelabs/getenvoy/pkg/flavors"
)

// Define template parameter names
const endpoint string = "endpoint"
const inport string = "inport"

// Flavor implements flavor.FlavorConfigTemplate interface
// and stores config data specific to Postgres template.
type Flavor struct {
	// Location of the postgres server
	endpoint string
	// Envoy's listener port
	inport string
}

var flavor Flavor

func init() {
	// Set default values.
	// Default values are not required to be present in cmd line.
	flavor.inport = "5432"
	flavors.AddTemplate("postgres", flavor)
}

// CheckParams verifies that passed template arguments are correct and
// are sufficient for creating a valid config from template.
func (Flavor) CheckParams(params map[string]string) (interface{}, error) {
	required := map[string]int{endpoint: 0}

	for param, value := range params {
		switch param {
		case endpoint:
			required[param]++
			flavor.endpoint = value
		case inport:
			if !valid.IsInt(value) {
				return nil, fmt.Errorf("Value for templateArg %s must be integer number", param)
			}
			flavor.inport = value
		default:
			fmt.Printf("Ignoring unrecognized template parameter: %s", param)
		}
	}

	// Check if all required params have been found in the parameter list
	var notFound string
	for key, count := range required {
		if count == 0 {
			notFound += key + " "
		}
	}
	if len(notFound) != 0 {
		return nil, fmt.Errorf("Required template params %s were not specified", notFound)
	}

	return flavor, nil
}

// GetTemplate returns unprocessed template for Envoy.
func (Flavor) GetTemplate() string {
	return configTemplate
}

// Postgres specific config file.
var configTemplate = `static_resources:
  listeners:
  - name: postgres_listener
    address:
      socket_address:
        address: 0.0.0.0
        port_value: {{ .inport }}
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
    type: static
    load_assignment:
      cluster_name: postgres_cluster
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: {{ .endpoint}} 
                port_value: 5432

admin:
  access_log_path: "/dev/null"
  address:
    socket_address:
      address: 0.0.0.0
      port_value: 8001`
