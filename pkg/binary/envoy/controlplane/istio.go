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
	_ "embed" //nolint
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	meshconfig "istio.io/api/mesh/v1alpha1"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pkg/bootstrap"
	"istio.io/istio/pkg/config/mesh"

	"github.com/tetratelabs/getenvoy/pkg/binary"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
)

const (
	defaultControlplane = "istio-pilot:15010"
	// boostrap.Config.CreateFileForEpoch(1) creates a file named envoy-rev1.json
	initialEpochBootstrap = "envoy-rev1.json"
)

// envoyBootstrapTemplate is the "envoy_bootstrap.json" from the Istio release tag or distribution
//go:embed istio-1.8.4/tools/packaging/common/envoy_bootstrap.json
var envoyBootstrapTemplate []byte

//  ^^ ex source: https://raw.githubusercontent.com/istio/istio/1.8.4/tools/packaging/common/envoy_bootstrap.json

// EnableIstioBootstrap tells GetEnvoy that it's using Istio for xDS and should bootstrap accordingly
func EnableIstioBootstrap(r *envoy.Runtime) {
	if r.Config.XDSAddress == "" {
		r.Config.XDSAddress = defaultControlplane
	}
	if len(r.Config.IPAddresses) == 0 {
		ips, err := retrieveIPs()
		if err != nil {
			panic(fmt.Sprintf("unable to retrieve IPs to be used in Istio bootstrap: %v", err))
		}
		r.Config.IPAddresses = ips
	}
	r.RegisterPreStart(writeBootstrap)
	r.RegisterPreStart(appendArgs)
}

func appendArgs(r binary.Runner) error {
	// Type assert as we're using Envoy specific config
	e, ok := r.(*envoy.Runtime)
	if !ok {
		return errors.New("unable to append Istio args to Envoy as binary.Runner is not an Envoy runtime")
	}
	args := []string{
		"--config-path", filepath.Join(e.DebugStore(), initialEpochBootstrap),
		"--drain-time-s", fmt.Sprint(e.Config.DrainDuration.Seconds),
	}
	r.AppendArgs(args)
	return nil
}

func writeBootstrap(r binary.Runner) error {
	// Type assert as we're using Envoy specific config
	e, ok := r.(*envoy.Runtime)
	if !ok {
		return errors.New("unable to write Istio bootstrap: binary.Runner is not an Envoy runtime")
	}
	cfg := generateIstioConfig(e)
	if err := writeProxyBootstrapTemplate(cfg.ProxyBootstrapTemplatePath); err != nil {
		return fmt.Errorf("unable to write Istio proxy bootstrap template: %w", err)
	}
	if _, err := bootstrap.New(bootstrap.Config{
		Node:     istioNode(e.Config),
		Proxy:    &cfg,
		LocalEnv: os.Environ(),
		NodeIPs:  e.Config.IPAddresses,
	}).CreateFileForEpoch(1); err != nil {
		return fmt.Errorf("unable to write Istio bootstrap: %v", err)
	}
	return nil
}

// Until Istio 1.10, Envoy bootstrap hard-codes tracing configuration. This parameter allows tests to override defaults.
// If set to nil, Envoy's /ready admin endpoint won't stick at PRE_INITIALIZING due to an unavailable Zipkin host.
// See https://github.com/istio/istio/issues/31553#issuecomment-802427832
var tracingConfig = mesh.DefaultProxyConfig().Tracing

func generateIstioConfig(e *envoy.Runtime) meshconfig.ProxyConfig {
	cfg := mesh.DefaultProxyConfig()
	cfg.ConfigPath = e.DebugStore()
	cfg.DiscoveryAddress = e.Config.XDSAddress
	cfg.ProxyAdminPort = e.Config.AdminPort
	cfg.ProxyBootstrapTemplatePath = filepath.Join(e.TmplDir, "istio_bootstrap_tmpl.json")
	cfg.EnvoyAccessLogService = &meshconfig.RemoteService{Address: e.Config.ALSAddresss}
	// Required: Defaults to MUTUAL_TLS, but we don't configure auth, yet, so it has to be set to NONE
	cfg.ControlPlaneAuthPolicy = meshconfig.AuthenticationPolicy_NONE
	cfg.Tracing = tracingConfig
	return cfg
}

func retrieveIPs() ([]string, error) {
	host, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	addrs, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	res := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		res = append(res, addr.String())
	}
	return res, nil
}

func istioNode(cfg *envoy.Config) string {
	mode := cfg.Mode
	if mode == envoy.LoadBalancer {
		mode = "router"
	}
	p := &model.Proxy{
		Type:        model.NodeType(mode),
		IPAddresses: cfg.IPAddresses,
		ID:          "unset",
		DNSDomain:   "unset",
	}
	return p.ServiceNode()
}

// TODO: (maybe?) Refactor the Istio write.Bootstrap upstream so we can pass it a byte slice instead
func writeProxyBootstrapTemplate(proxyBootstrapTemplatePath string) error {
	if err := os.MkdirAll(filepath.Dir(proxyBootstrapTemplatePath), os.ModePerm); err != nil {
		return err
	}
	if err := os.WriteFile(proxyBootstrapTemplatePath, envoyBootstrapTemplate, 0600); err != nil {
		return err
	}
	return nil
}
