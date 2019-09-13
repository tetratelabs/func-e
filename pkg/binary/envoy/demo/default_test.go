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

package demo

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoytest"
)

// NOTE: This test will fail on macOS due to an issue with Envoy, the same issue as debug logging
func Test_DefaultConfig(t *testing.T) {
	if err := envoytest.Fetch(); err != nil {
		t.Fatalf("error fetching Envoy: %v", err)
	}
	t.Run("writes and uses a default config", func(t *testing.T) {
		runtime, _ := envoy.NewRuntime(StaticBootstrap)
		defer os.RemoveAll(runtime.DebugStore() + ".tar.gz")
		defer os.RemoveAll(runtime.DebugStore())
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		assert.NoError(t, envoytest.Run(ctx, runtime, ""))
		makeRequest(t, "google.com")
		makeRequest(t, "bing.com")
		assert.NoError(t, envoytest.Kill(ctx, runtime))
	})
}

func makeRequest(t *testing.T, host string) {
	req, _ := http.NewRequest("GET", "http://localhost:15001", nil)
	req.Header.Set("Host", host)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got non-200 status code with host header %s: %v", host, resp.StatusCode)
	}
}
