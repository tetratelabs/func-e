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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/require"
	"istio.io/istio/pilot/pkg/bootstrap"
	"istio.io/istio/pilot/pkg/networking/plugin"
	"istio.io/istio/pkg/config/mesh"
	"istio.io/istio/pkg/keepalive"

	"github.com/tetratelabs/getenvoy/pkg/binary"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/debug"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoytest"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

func TestMain(m *testing.M) {
	envoytest.FetchEnvoyAndRun(m)
}

func TestConnectsToMockPilotAsAGateway(t *testing.T) {
	pilotGrpcAddr, stopPilot := requireMockPilot(t, "testdata")
	defer stopPilot()

	runtime, err := envoy.NewRuntime(
		func(r *envoy.Runtime) {
			r.Config.Mode = envoy.ParseMode("loadbalancer")
			r.Config.XDSAddress = pilotGrpcAddr
			r.Config.IPAddresses = []string{"1.1.1.1"}
		},
		Istio,                                // this integrates and ends up reading pilotGrpcAddr
		debug.EnableEnvoyAdminDataCollection, // this allows us to read clusters.txt
	)
	require.NoError(t, err, "error creating envoy runtime")
	defer os.RemoveAll(runtime.DebugStore())

	envoytest.RequireRunKill(t, runtime, envoytest.RunKillOptions{
		// Envoy won't achieve ready status because endpoints in testdata/configs.yaml are unreachable
		ExpectedStatus: binary.StatusInitializing,
		// We sleep to allow the configuration to populate, so that we can read it back.
		SleepBeforeKill: 5 * time.Second,
		// Assertions below inspect files in the debug store
		RetainDebugStore: true,
	})

	// Verify configuration from istio testdata/configs.yaml ended up in envoy
	gotClusters, err := ioutil.ReadFile(filepath.Join(runtime.DebugStore(), "clusters.txt"))
	require.NoError(t, err, "error getting envoy clusters")
	require.Contains(t, string(gotClusters), "istio-ingressgateway.istio-system.svc.cluster.local::1.1.1.1:8443")
}

// requireMockPilot will ensure a pilot server and returns its gRPC address and a function to stop it.
func requireMockPilot(t *testing.T, configFileDir string) (string, func()) {
	configFileDir = morerequire.RequireAbsDir(t, configFileDir)
	cleanups := make([]func(), 0)

	cleanup := func() {
		for i := len(cleanups) - 1; i >= 0; i-- {
			cleanups[i]()
		}
	}

	// This ensures on any panic the pilot process is stopped, which can prevent test hangs.
	deferredCleanup := cleanup
	defer func() {
		if deferredCleanup != nil {
			deferredCleanup()
		}
	}()

	// In case pilot tries to open any relative paths, make sure they are writeable
	workDir, removeWorkDir := morerequire.RequireNewTempDir(t)
	cleanups = append(cleanups, removeWorkDir)

	_, revertChDir := morerequire.RequireChDir(t, workDir)
	cleanups = append(cleanups, revertChDir)

	meshConfig := mesh.DefaultMeshConfig()
	meshConfig.EnableTracing = false
	// TODO: figure out how to set telemetry enabled = false as this reduces the amount of listeners
	meshConfig.EnablePrometheusMerge = &types.BoolValue{Value: false}
	meshConfig.DnsRefreshRate = types.DurationProto(time.Minute)
	meshConfig.DisableMixerHttpReports = true
	// Create a test pilot discovery service configured to watch the tempDir.
	args := bootstrap.PilotArgs{
		Namespace: "testing",
		Config:    bootstrap.ConfigArgs{FileDir: configFileDir},
		DiscoveryOptions: bootstrap.DiscoveryServiceOptions{
			HTTPAddr: "127.0.0.1:", // instead of listening on all addresses
			GrpcAddr: "127.0.0.1:",
		},
		MeshConfig:       &meshConfig,
		MCPOptions:       bootstrap.MCPOptions{MaxMessageSize: 1024 * 1024 * 4},
		KeepaliveOptions: keepalive.DefaultOption(),
		ForceStop:        true,
		Plugins:          []string{plugin.Health}, // only health, no mixer or authorization
		Service:          bootstrap.ServiceArgs{}, // no service registries
	}

	s, err := bootstrap.NewServer(&args)
	require.NoError(t, err, "failed to bootstrap mock pilot server with args: %s", args)

	// Start the mock pilot server, passing a stop channel used for cleanup later
	stop := make(chan struct{})
	err = s.Start(stop)
	require.NoError(t, err, "failed to start mock pilot server: %s", s)
	cleanups = append(cleanups, func() { close(stop) })

	// await /ready endpoint
	client := &http.Client{Timeout: 1 * time.Second}
	checkURL := fmt.Sprintf("http://%s/ready", s.HTTPListener.Addr())
	require.Eventually(t, func() bool {
		resp, err := client.Get(checkURL)
		if err != nil {
			return false
		}
		defer resp.Body.Close()

		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 100*time.Millisecond, "error waiting for pilot to be /ready")

	deferredCleanup = nil // We succeeded, so don't need to tear down before returning

	return s.GRPCListener.Addr().String(), cleanup
}
