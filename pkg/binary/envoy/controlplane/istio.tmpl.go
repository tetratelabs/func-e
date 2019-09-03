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

package controlplane

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// IstioVersion details the version of Istio from which the bootstrap comes from
const IstioVersion = "1.2.5"

// TODO: (maybe?) Refactor the Istio write.Bootstrap upstream so we can pass it a byte slice instead
func writeIstioTemplate(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}
	return ioutil.WriteFile(path, []byte(istioBootStrapTemplate), 0644)
}

var istioBootStrapTemplate = `{
  "node": {
    "id": "{{ .nodeID }}",
    "cluster": "{{ .cluster }}",
    "locality": {
      {{ if .region }}
      "region": "{{ .region }}",
      {{ end }}
      {{ if .zone }}
      "zone": "{{ .zone }}",
      {{ end }}
      {{ if .sub_zone }}
      "sub_zone": "{{ .sub_zone }}",
      {{ end }}
    },
    "metadata": {{ .meta_json_str }}
  },
  "stats_config": {
    "use_all_default_tags": false,
    "stats_tags": [
      {
        "tag_name": "cluster_name",
        "regex": "^cluster\\.((.+?(\\..+?\\.svc\\.cluster\\.local)?)\\.)"
      },
      {
        "tag_name": "tcp_prefix",
        "regex": "^tcp\\.((.*?)\\.)\\w+?$"
      },
      {
        "tag_name": "response_code",
        "regex": "_rq(_(\\d{3}))$"
      },
      {
        "tag_name": "response_code_class",
        "regex": "_rq(_(\\dxx))$"
      },
      {
        "tag_name": "http_conn_manager_listener_prefix",
        "regex": "^listener(?=\\.).*?\\.http\\.(((?:[_.[:digit:]]*|[_\\[\\]aAbBcCdDeEfF[:digit:]]*))\\.)"
      },
      {
        "tag_name": "http_conn_manager_prefix",
        "regex": "^http\\.(((?:[_.[:digit:]]*|[_\\[\\]aAbBcCdDeEfF[:digit:]]*))\\.)"
      },
      {
        "tag_name": "listener_address",
        "regex": "^listener\\.(((?:[_.[:digit:]]*|[_\\[\\]aAbBcCdDeEfF[:digit:]]*))\\.)"
      },
      {
        "tag_name": "mongo_prefix",
        "regex": "^mongo\\.(.+?)\\.(collection|cmd|cx_|op_|delays_|decoding_)(.*?)$"
      }
    ],
    "stats_matcher": {
      "inclusion_list": {
        "patterns": [
          {{- range $a, $s := .inclusionPrefix }}
          {
          "prefix": "{{$s}}"
          },
          {{- end }}
          {{- range $a, $s := .inclusionSuffix }}
          {
          "suffix": "{{$s}}"
          },
          {{- end }}
          {{- range $a, $s := .inclusionRegexps }}
          {
          "regex": "{{js $s}}"
          },
          {{- end }}
        ]
      }
    }
  },
  "admin": {
    "access_log_path": "/dev/null",
    "address": {
      "socket_address": {
        "address": "{{ .localhost }}",
        "port_value": {{ .config.ProxyAdminPort }}
      }
    }
  },
  "dynamic_resources": {
    "lds_config": {
      "ads": {}
    },
    "cds_config": {
      "ads": {}
    },
    "ads_config": {
      "api_type": "GRPC",
      "grpc_services": [
        {
          "envoy_grpc": {
            "cluster_name": "xds-grpc"
          }
        }
      ]
    }
  },
  "static_resources": {
    "clusters": [
      {
        "name": "prometheus_stats",
        "type": "STATIC",
        "connect_timeout": "0.250s",
        "lb_policy": "ROUND_ROBIN",
        "hosts": [
          {
            "socket_address": {
              "protocol": "TCP",
              "address": "{{ .localhost }}",
              "port_value": {{ .config.ProxyAdminPort }}
            }
          }
        ]
      },
      {
        "name": "xds-grpc",
        "type": "STRICT_DNS",
        "dns_refresh_rate": "{{ .dns_refresh_rate }}",
        "dns_lookup_family": "{{ .dns_lookup_family }}",
        "connect_timeout": "{{ .connect_timeout }}",
        "lb_policy": "ROUND_ROBIN",
        {{ if eq .config.ControlPlaneAuthPolicy 1 }}
        "tls_context": {
          "common_tls_context": {
            "alpn_protocols": [
              "h2"
            ],
            "tls_certificates": [
              {
                "certificate_chain": {
                  "filename": "/etc/certs/cert-chain.pem"
                },
                "private_key": {
                  "filename": "/etc/certs/key.pem"
                }
              }
            ],
            "validation_context": {
              "trusted_ca": {
                "filename": "/etc/certs/root-cert.pem"
              },
              "verify_subject_alt_name": [
                {{- range $a, $s := .pilot_SAN }}
                "{{$s}}"
                {{- end}}
              ]
            }
          }
        },
        {{ end }}
        "hosts": [
          {
            "socket_address": {{ .pilot_grpc_address }}
          }
        ],
        "circuit_breakers": {
          "thresholds": [
            {
              "priority": "DEFAULT",
              "max_connections": 100000,
              "max_pending_requests": 100000,
              "max_requests": 100000
            },
            {
              "priority": "HIGH",
              "max_connections": 100000,
              "max_pending_requests": 100000,
              "max_requests": 100000
            }
          ]
        },
        "upstream_connection_options": {
          "tcp_keepalive": {
            "keepalive_time": 300
          }
        },
        "http2_protocol_options": { }
      }
      {{ if .zipkin }}
      ,
      {
        "name": "zipkin",
        "type": "STRICT_DNS",
        "dns_refresh_rate": "{{ .dns_refresh_rate }}",
        "dns_lookup_family": "{{ .dns_lookup_family }}",
        "connect_timeout": "1s",
        "lb_policy": "ROUND_ROBIN",
        "hosts": [
          {
            "socket_address": {{ .zipkin }}
          }
        ]
      }
      {{ else if .lightstep }}
      ,
      {
        "name": "lightstep",
        "http2_protocol_options": {},
        {{ if .lightstepSecure }}
        "tls_context": {
          "common_tls_context": {
            "alpn_protocols": [
              "h2"
            ],
            "validation_context": {
              "trusted_ca": {
                "filename": "{{ .lightstepCacertPath }}"
              }
            }
          }
        },
        {{ end }}
        "type": "STRICT_DNS",
        "dns_refresh_rate": "{{ .dns_refresh_rate }}",
        "dns_lookup_family": "{{ .dns_lookup_family }}",
        "connect_timeout": "1s",
        "lb_policy": "ROUND_ROBIN",
        "hosts": [
          {
            "socket_address": {{ .lightstep }}
          }
        ]
      }
      {{ else if .datadog }}
      ,
      {
        "name": "datadog_agent",
        "connect_timeout": "1s",
        "type": "STRICT_DNS",
        "dns_refresh_rate": "{{ .dns_refresh_rate }}",
        "dns_lookup_family": "{{ .dns_lookup_family }}",
        "lb_policy": "ROUND_ROBIN",
        "hosts": [
          {
            "socket_address": {{ .datadog }}
          }
        ]
      }
      {{ end }}
      {{ if .envoy_metrics_service }}
      ,
      {
        "name": "envoy_metrics_service",
        "type": "STRICT_DNS",
        "dns_refresh_rate": "{{ .dns_refresh_rate }}",
        "dns_lookup_family": "{{ .dns_lookup_family }}",
        "connect_timeout": "1s",
        "lb_policy": "ROUND_ROBIN",
        "http2_protocol_options": {},
        "hosts": [
          {
            "socket_address": {{ .envoy_metrics_service }}
          }
        ]
      }
      {{ end }}
    ],
    "listeners":[
      {
        "address": {
          "socket_address": {
            "protocol": "TCP",
            "address": "{{ .wildcard }}",
            "port_value": 15090
          }
        },
        "filter_chains": [
          {
            "filters": [
              {
                "name": "envoy.http_connection_manager",
                "config": {
                  "codec_type": "AUTO",
                  "stat_prefix": "stats",
                  "route_config": {
                    "virtual_hosts": [
                      {
                        "name": "backend",
                        "domains": [
                          "*"
                        ],
                        "routes": [
                          {
                            "match": {
                              "prefix": "/stats/prometheus"
                            },
                            "route": {
                              "cluster": "prometheus_stats"
                            }
                          }
                        ]
                      }
                    ]
                  },
                  "http_filters": {
                    "name": "envoy.router"
                  }
                }
              }
            ]
          }
        ]
      }
    ]
  }
  {{ if .zipkin }}
  ,
  "tracing": {
    "http": {
      "name": "envoy.zipkin",
      "config": {
        "collector_cluster": "zipkin",
        "collector_endpoint": "/api/v1/spans",
        "trace_id_128bit": "true",
        "shared_span_context": "false"
      }
    }
  }
  {{ else if .lightstep }}
  ,
  "tracing": {
    "http": {
      "name": "envoy.lightstep",
      "config": {
        "collector_cluster": "lightstep",
        "access_token_file": "{{ .lightstepToken}}"
      }
    }
  }
  {{ else if .datadog }}
  ,
  "tracing": {
    "http": {
      "name": "envoy.tracers.datadog",
      "config": {
        "collector_cluster": "datadog_agent",
        "service_name": "{{ .cluster }}"
      }
    }
  }
  {{ end }}
  {{ if or .envoy_metrics_service .statsd }}
  ,
  "stats_sinks": [
    {{ if .envoy_metrics_service }}
    {
      "name": "envoy.metrics_service",
      "config": {
        "grpc_service": {
          "envoy_grpc": {
            "cluster_name": "envoy_metrics_service"
          }
        }
      }
    },
    {{ end }}
    {{ if .statsd }}
    {
      "name": "envoy.statsd",
      "config": {
        "address": {
          "socket_address": {{ .statsd }}
        }
      }
    },
    {{ end }}
  ]
  {{ end }}
}
`
