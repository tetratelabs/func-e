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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/debug"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoytest"
)

// NOTE: This test will fail on macOS due to an issue with Envoy, the same issue as debug logging
func Test_DefaultConfig(t *testing.T) {
	if err := envoytest.Fetch(); err != nil {
		t.Fatalf("error fetching Envoy: %v", err)
	}
	t.Run("writes and uses a default config", func(t *testing.T) {
		runtime, _ := envoy.NewRuntime(
			debug.EnableEnvoyAdminDataCollection,
			DefaultStaticBootstrap,
		)
		defer os.RemoveAll(runtime.DebugStore() + ".tar.gz")
		defer os.RemoveAll(runtime.DebugStore())
		assert.NoError(t, envoytest.RunKill(runtime, "", time.Second*5))
		gotListeners, _ := ioutil.ReadFile(filepath.Join(runtime.DebugStore(), "listeners.txt"))
		assert.Contains(t, string(gotListeners), "::0.0.0.0:15001")
	})
}
