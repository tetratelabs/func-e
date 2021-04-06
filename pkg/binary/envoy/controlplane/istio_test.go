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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"istio.io/istio/pilot/pkg/bootstrap"
	"istio.io/istio/tests/util"

	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/debug"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoytest"
)

func TestConnectsToMockPilotAsAGateway(t *testing.T) {
	err := envoytest.Fetch()
	require.NoError(t, err, "error running envoytest.Fetch()")
	_, teardown := setupMockPilot()
	defer teardown()

	cfg := envoy.NewConfig(
		func(c *envoy.Config) {
			c.Mode = envoy.ParseMode("loadbalancer")
			c.XDSAddress = util.MockPilotGrpcAddr
			c.IPAddresses = []string{"1.1.1.1"}
		},
	)

	runtime, err := envoy.NewRuntime(
		func(r *envoy.Runtime) { r.Config = cfg },
		debug.EnableEnvoyAdminDataCollection,
		Istio,
	)
	require.NoError(t, err, "error creating envoy runtime")
	defer os.RemoveAll(runtime.DebugStore())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err = envoytest.Run(ctx, runtime, "")
	require.NoError(t, err, "error running envoy")

	time.Sleep(time.Millisecond * 500) // Pilot config propagation

	err = envoytest.Kill(ctx, runtime)
	require.NoError(t, err, "error killing envoy")

	gotListeners, err := ioutil.ReadFile(filepath.Join(runtime.DebugStore(), "listeners.txt"))
	require.NoError(t, err, "error killing envoy listeners")

	require.Contains(t, string(gotListeners), "0.0.0.0_8443::0.0.0.0:8443")
	require.Contains(t, string(gotListeners), "0.0.0.0_8080::0.0.0.0:8080")
}

func setupMockPilot() (*bootstrap.Server, util.TearDownFunc) {
	return util.EnsureTestServer(func(args *bootstrap.PilotArgs) {
		bootstrap.PilotCertDir = "testdata"
		args.Config.FileDir = "testdata"
		args.Plugins = bootstrap.DefaultPlugins
		args.Mesh.MixerAddress = ""
		// In a normal macOS setup, you cannot write to /dev/stdout, which is the default path here.
		// While not Docker-specific, there are related notes here https://github.com/moby/moby/issues/31243
		// Since this test doesn't read access logs anyway, the easier workaround is to disable access logging.
		args.MeshConfig.AccessLogFile = ""
		args.Service.Registries = []string{}
	})
}
