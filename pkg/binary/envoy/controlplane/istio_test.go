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
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/debug"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoytest"
	"istio.io/istio/pilot/pkg/bootstrap"
	"istio.io/istio/tests/util"
)

// more of an integration test than a unit test
func Test_IstioGateway(t *testing.T) {
	t.Run("connects to mock Pilot as a gateway", func(t *testing.T) {
		_, teardown := setupMockPilot()
		defer teardown()
		cfg := envoy.NewConfig(
			func(c *envoy.Config) {
				c.Mode = envoy.ParseMode("router")
				c.XDSAddress = util.MockPilotGrpcAddr
				c.IPAddresses = []string{"1.1.1.1"}
			},
		)
		runtime, _ := envoy.NewRuntime(
			func(r *envoy.Runtime) { r.Config = cfg },
			debug.EnableEnvoyAdminDataCollection,
			Istio,
		)
		defer os.RemoveAll(runtime.DebugStore() + ".tar.gz")
		defer os.RemoveAll(runtime.DebugStore())
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		assert.NoError(t, envoytest.Run(ctx, runtime, ""))
		assert.NoError(t, envoytest.Kill(ctx, runtime))
	})

	// t.Run("mutates Envoy to a runnable state", func(t *testing.T) {
	// 	r := defaultIstioRuntime()
	// 	writeBootstrap(r)
	// 	envoy, _ := envoy.NewRuntime()
	// 	defer os.RemoveAll(r.DebugStore() + ".tar.gz")
	// 	defer os.RemoveAll(r.DebugStore())
	// 	assert.NoError(t, envoytest.Run(envoy, filepath.Join(r.DebugStore(), initialEpochBootstrap)))
	// })
}

func setupMockPilot() (*bootstrap.Server, util.TearDownFunc) {
	return util.EnsureTestServer(func(args *bootstrap.PilotArgs) {
		bootstrap.PilotCertDir = "testdata"
		// args.DiscoveryOptions.SecureGrpcAddr = ""
		args.Plugins = bootstrap.DefaultPlugins
		args.Config.FileDir = "testdata"
		args.Mesh.MixerAddress = ""
		args.Mesh.RdsRefreshDelay = nil
		// args.Mesh.ConfigFile = "testdata/ingress-gateway/mesh.yaml"
		args.Service.Registries = []string{}
	})
}
