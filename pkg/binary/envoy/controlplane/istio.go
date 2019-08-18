package controlplane

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/tetratelabs/getenvoy/pkg/binary"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	meshconfig "istio.io/api/mesh/v1alpha1"

	"istio.io/istio/pilot/pkg/model"
	agent "istio.io/istio/pkg/bootstrap"
	"istio.io/istio/pkg/config/mesh"
)

const defaultControlplane = "istio-pilot:15010"

// Istio tells GetEnvoy that it's using Istio for xDS and should bootstrap accordingly
var Istio = func(r *envoy.Runtime) {
	if len(r.Config.XDSAddress) == 0 {
		r.Config.XDSAddress = defaultControlplane
	}
	ips, err := retrieveIPs()
	if err != nil {
		panic(fmt.Sprintf("unable to retrieve IPs to be used in Istio bootstrap: %v", err))
	}
	r.Config.IPAddresses = ips
	r.RegisterPreStart(writeBootstrap)
	r.RegisterPreStart(appendArgs)
}

func appendArgs(r binary.Runner) error {
	// Type assert as we're using Envoy specific config
	envoy, ok := r.(*envoy.Runtime)
	if !ok {
		return errors.New("unable to append Istio args to Envoy as binary.Runner is not an Envoy runtime")
	}
	args := []string{
		"--config-path", filepath.Join(envoy.DebugStore(), "envoy-rev1.json"),
		"--drain-time-s", fmt.Sprint(int(convertDuration(envoy.Config.DrainDuration) / time.Second)),
		"--max-obj-name-len", fmt.Sprint(envoy.Config.StatNameLength),
	}
	r.AppendArgs(args)
	return nil
}

func convertDuration(d *types.Duration) time.Duration {
	if d == nil {
		return 0
	}
	dur, _ := types.DurationFromProto(d)
	return dur
}

func writeBootstrap(r binary.Runner) error {
	// Type assert as we're using Envoy specific config
	envoy, ok := r.(*envoy.Runtime)
	if !ok {
		return errors.New("unable to write Istio bootstrap: binary.Runner is not an Envoy runtime")
	}
	cfg := mesh.DefaultProxyConfig()
	cfg.ConfigPath = envoy.DebugStore()
	cfg.DiscoveryAddress = envoy.Config.XDSAddress
	cfg.ProxyAdminPort = envoy.Config.AdminPort
	cfg.ProxyBootstrapTemplatePath = filepath.Join(envoy.TmplDir, "istio_bootstrap_tmpl.json")
	if err := writeIstioTemplate(cfg.ProxyBootstrapTemplatePath); err != nil {
		return fmt.Errorf("unable to write Istio bootstrap template: %v", err)
	}
	cfg.EnvoyAccessLogService = &meshconfig.RemoteService{Address: envoy.Config.ALSAddresss}
	// cfg.ControlPlaneAuthPolicy = v1alpha1.AuthenticationPolicy_MUTUAL_TLS // TODO: turn on!
	if _, err := agent.WriteBootstrap(&cfg, istioNode(envoy.Config), 1, []string{}, nil, os.Environ(), envoy.Config.IPAddresses, "60s"); err != nil {
		return fmt.Errorf("unable to write Istio bootstrap: %v", err)
	}
	return nil
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
	p := &model.Proxy{
		Type:        model.NodeType(cfg.Mode),
		IPAddresses: cfg.IPAddresses,
		ID:          "unset",
		DNSDomain:   "unset",
	}
	return p.ServiceNode()
}
