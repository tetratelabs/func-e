package postgres

import (
	"fmt"
	"github.com/tetratelabs/getenvoy/pkg/flavors"
	valid "github.com/asaskevich/govalidator"
)

// Define template parameter names
const Endpoint string = "Endpoint"
const InPort   string = "InPort"

type PostgresFlavor struct {
	// Location of the postgres server 
	Endpoint string
	// Envoy's listener port
	InPort   string
}

var postgresFlavor PostgresFlavor

func init () {
	// Set default values.
	// Default values are not required to be present in cmd line.
	postgresFlavor.InPort = "5432"
	flavor.AddTemplate("postgres", postgresFlavor)
}


func (PostgresFlavor) CreateConfig(params map[string]string) (error, string) {
	required := map[string]int {Endpoint: 0}

	for  param, _ := range params {
		if _, ok := required[param]; ok {
			required[param]++
			postgresFlavor.Endpoint = params[param]
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
		return fmt.Errorf("Template params %s were not specified", notFound), ""
	}

	return nil, ""
}
 

func (PostgresFlavor) CheckParams(params map[string]string) (error, interface{}) {
	required := map[string]int {Endpoint: 0}

	for  param, value := range params {
		switch param {
		case Endpoint:
			required[param]++
			postgresFlavor.Endpoint = value 
		case InPort:
			if !valid.IsInt(value) {
				return fmt.Errorf("Value for templateArg %s must be integer number", param), nil
			}
			postgresFlavor.InPort = value 
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
		return fmt.Errorf("Required template params %s were not specified", notFound), nil
	}

	return nil, postgresFlavor
}

func (PostgresFlavor) GetTemplate() string {
	return configTemplate
}
 


// Postgres specific config file.
var configTemplate string = 
`static_resources:
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
    type: static
    load_assignment:
      cluster_name: postgres_cluster
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: {{ .Endpoint}} 
                port_value: 5432

admin:
  access_log_path: "/dev/null"
  address:
    socket_address:
      address: 0.0.0.0
      port_value: 8001`

