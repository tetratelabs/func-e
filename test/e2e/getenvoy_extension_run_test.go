// Copyright 2020 Tetrate
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

package e2e_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/log"

	. "github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

const extensionName = "getenvoy_extension_run"
const terminateTimeout = 2 * time.Minute

// TestGetEnvoyExtensionRun runs the equivalent of "getenvoy extension run" for a matrix of extension.Categories and
// extension.Languages. "getenvoy extension init" is a prerequisite, so run first.
//
// "getenvoy extension run" uses Docker. See TestMain for general notes on about the test runtime.
func TestGetEnvoyExtensionRun(t *testing.T) {
	debugDir, revertOriginalDebugDir := backupDebugDir(t)
	defer revertOriginalDebugDir()

	for _, test := range getExtensionTestMatrix() {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.String(), func(t *testing.T) {
			workDir, removeWorkDir := RequireNewTempDir(t)
			defer removeWorkDir()

			_, revertChDir := RequireChDir(t, workDir)
			defer revertChDir()

			// run requires "get envoy extension init" to have succeeded
			requireExtensionInit(t, workDir, test.Category, test.Language, extensionName)
			defer requireExtensionClean(t, workDir)

			c := getEnvoy("extension run").Args(getToolchainContainerOptions()...)
			stdout, stderr, terminate := c.Start(t, terminateTimeout)

			// The underlying call is conditional to ensure errors that raise before we stop the server, stop it.
			deferredTerminate := terminate
			defer func() {
				if deferredTerminate != nil {
					deferredTerminate()
				}
			}()

			envoyClient := requireEnvoyReady(t, stdout, stderr, c)

			log.Infof(`waiting for Wasm extensions after running [%v]`, c)
			require.Eventually(t, func() bool {
				stats, e := envoyClient.GetStats()
				if e != nil {
					return false
				}
				// at the moment, the only available Wasm metric is the number of Wasm VMs
				concurrency := stats.GetMetric("server.concurrency")
				activeWasmVms := stats.GetMetric("wasm.envoy.wasm.runtime.v8.active")
				return concurrency != nil && activeWasmVms != nil && activeWasmVms.Value == concurrency.Value+2
			}, 1*time.Minute, 100*time.Millisecond, `wasm stats never found after running [%v]`, c)

			log.Infof(`stopping Envoy after running [%v]`, c)
			terminate()
			deferredTerminate = nil

			verifyDebugDump(t, debugDir, c)
		})
	}
}
