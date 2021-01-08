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
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/golang/protobuf/ptypes"
	durationpb "github.com/golang/protobuf/ptypes/duration"

	"github.com/tetratelabs/getenvoy/pkg/binary"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/log"
	meshconfig "istio.io/api/mesh/v1alpha1"

	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pkg/bootstrap"
	"istio.io/istio/pkg/config/mesh"
)

const (
	defaultControlplane   = "istio-pilot:15010"
	initialEpochBootstrap = "envoy-rev1.json"
)

// Istio tells GetEnvoy that it's using Istio for xDS and should bootstrap accordingly
func Istio(r *envoy.Runtime) {
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
		"--drain-time-s", fmt.Sprint(int(convertDuration(e.Config.DrainDuration) / time.Second)),
		"--max-obj-name-len", fmt.Sprint(e.Config.StatNameLength),
	}
	r.AppendArgs(args)
	return nil
}

func convertDuration(d *durationpb.Duration) time.Duration {
	if d == nil {
		return 0
	}
	dur, err := ptypes.Duration(d)
	if err != nil {
		log.Warnf("unable to convert proto duration %v to time.Duration", d)
	}
	return dur
}

func writeBootstrap(r binary.Runner) error {
	// Type assert as we're using Envoy specific config
	e, ok := r.(*envoy.Runtime)
	if !ok {
		return errors.New("unable to write Istio bootstrap: binary.Runner is not an Envoy runtime")
	}
	cfg := generateIstioConfig(e)
	if err := writeIstioTemplate(cfg.ProxyBootstrapTemplatePath); err != nil {
		return fmt.Errorf("unable to write Istio bootstrap template: %v", err)
	}
	if _, err := bootstrap.New(bootstrap.Config{
		Node:           istioNode(e.Config),
		DNSRefreshRate: "60s",
		Proxy:          &cfg,
		LocalEnv:       os.Environ(),
		NodeIPs:        e.Config.IPAddresses,
	}).CreateFileForEpoch(1); err != nil {
		return fmt.Errorf("unable to write Istio bootstrap: %v", err)
	}
	return nil
}

func generateIstioConfig(e *envoy.Runtime) meshconfig.ProxyConfig {
	cfg := mesh.DefaultProxyConfig()
	cfg.ConfigPath = e.DebugStore()
	cfg.DiscoveryAddress = e.Config.XDSAddress
	cfg.ProxyAdminPort = e.Config.AdminPort
	cfg.ProxyBootstrapTemplatePath = filepath.Join(e.TmplDir, "istio_bootstrap_tmpl.json")
	cfg.EnvoyAccessLogService = &meshconfig.RemoteService{Address: e.Config.ALSAddresss}
	// cfg.ControlPlaneAuthPolicy = v1alpha1.AuthenticationPolicy_MUTUAL_TLS // TODO: turn on!
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
