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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/debug"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoytest"
	"istio.io/istio/pilot/pkg/bootstrap"
	"istio.io/istio/tests/util"
)

func TestMain(m *testing.M) {
	if err := envoytest.Fetch(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// NOTE: This test will fail on macOS due to an issue with Envoy, the same issue as debug logging
func Test_IstioGateway(t *testing.T) {
	t.Run("connects to mock Pilot as a gateway", func(t *testing.T) {
		_, teardown := setupMockPilot()
		defer teardown()
		cfg := envoy.NewConfig(
			func(c *envoy.Config) {
				c.Mode = envoy.ParseMode("loadbalancer")
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
		time.Sleep(time.Millisecond * 500) // Pilot config propagation
		assert.NoError(t, envoytest.Kill(ctx, runtime))
		gotListeners, _ := ioutil.ReadFile(filepath.Join(runtime.DebugStore(), "listeners.txt"))
		assert.Contains(t, string(gotListeners), "0.0.0.0_8443::0.0.0.0:8443")
		assert.Contains(t, string(gotListeners), "0.0.0.0_8080::0.0.0.0:8080")
	})
}

func setupMockPilot() (*bootstrap.Server, util.TearDownFunc) {
	return util.EnsureTestServer(func(args *bootstrap.PilotArgs) {
		bootstrap.PilotCertDir = "testdata"
		args.Config.FileDir = "testdata"
		args.Plugins = bootstrap.DefaultPlugins
		args.Mesh.MixerAddress = ""
		args.Mesh.RdsRefreshDelay = nil
		args.Service.Registries = []string{}
	})
}
