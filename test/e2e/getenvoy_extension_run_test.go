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
	"io/ioutil"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/mholt/archiver"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/log"

	e2e "github.com/tetratelabs/getenvoy/test/e2e/util"
	utilenvoy "github.com/tetratelabs/getenvoy/test/e2e/util/envoy"
)

// TestGetEnvoyExtensionRun runs the equivalent of "getenvoy extension run" for a matrix of extension.Categories and
// extension.Languages. "getenvoy extension init" is a prerequisite, so run first.
//
// Note: "getenvoy extension run" can be extremely slow due to implicit responsibilities such as downloading modules
// or compilation. This uses Docker, so changes to the Dockerfile or contents like "commands.sh" effect performance.
func TestGetEnvoyExtensionRun(t *testing.T) {
	const extensionName = "getenvoy_extension_run"
	const terminateTimeout = 2 * time.Minute
	requireEnvoyBinaryPath(t) // Ex. After running "make bin", E2E_GETENVOY_BINARY=$PWD/build/bin/darwin/amd64/getenvoy

	for _, test := range e2e.GetCategoryLanguageCombinations() {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.String(), func(t *testing.T) {
			// We override the home directory ~/.getenvoy so as to not overwrite any debug info there. We need to do
			// this to ensure presence of debug files are from this run, not another one.
			homeDir, removeHomeDir := requireNewTempDir(t)
			defer removeHomeDir()

			workDir, removeWorkDir := requireNewTempDir(t)
			defer removeWorkDir()

			revertChDir := requireChDir(t, workDir)
			defer revertChDir()

			// run requires "get envoy extension init" to have succeeded
			requireExtensionInit(t, workDir, test.Category, test.Language, extensionName)
			defer requireExtensionClean(t, workDir)

			// "getenvoy extension run" only returns stdout because `docker run -t` redirects stderr to stdout.
			cmd := GetEnvoy("extension run --envoy-options '-l trace'").
				Args("--home-dir", homeDir).
				Args(e2e.Env.GetBuiltinContainerOptions()...)
			_, stderr, terminate := cmd.Start(t, terminateTimeout)
			defer terminate()

			stderrLines := e2e.StreamLines(stderr).Named("stderr")

			log.Infof(`waiting for Envoy Admin address to get logged after running [%v]`, cmd)
			adminAddressPattern := regexp.MustCompile(`admin address: ([^:]+:[0-9]+)`)
			line, err := stderrLines.FirstMatch(adminAddressPattern).Wait(10 * time.Minute) // give time to compile the extension
			require.NoError(t, err, `error parsing admin address from stderr of [%v]`, cmd)
			adminAddress := adminAddressPattern.FindStringSubmatch(line)[1]

			log.Infof(`waiting for Envoy start-up to complete after running [%v]`, cmd)
			_, err = stderrLines.FirstMatch(regexp.MustCompile(`starting main dispatch loop`)).Wait(1 * time.Minute)
			require.NoError(t, err, `error parsing startup from stderr of [%v]`, cmd)

			log.Infof(`waiting for Envoy client to connect after running [%v]`, cmd)
			envoyClient, err := utilenvoy.NewClient(adminAddress)
			require.NoError(t, err, `error from envoy client %s after running [%v]`, adminAddress, cmd)
			require.Eventually(t, func() bool {
				ready, e := envoyClient.IsReady()
				return e == nil && ready
			}, 1*time.Minute, 100*time.Millisecond, `envoy client %s never ready after running [%v]`, adminAddress, cmd)

			log.Infof(`waiting for Wasm extensions after running [%v]`, cmd)
			require.Eventually(t, func() bool {
				stats, e := envoyClient.GetStats()
				if e != nil {
					return false
				}
				// at the moment, the only available Wasm metric is the number of Wasm VMs
				concurrency := stats.GetMetric("server.concurrency")
				activeWasmVms := stats.GetMetric("wasm.envoy.wasm.runtime.v8.active")
				return concurrency != nil && activeWasmVms != nil && activeWasmVms.Value == concurrency.Value+2
			}, 1*time.Minute, 100*time.Millisecond, `wasm stats never found after running [%v]`, adminAddress, cmd)

			log.Infof(`stopping Envoy after running [%v]`, cmd)
			terminate()

			// verify the debug dump of Envoy state has been taken
			debugDir := filepath.Join(homeDir, "debug")
			files, err := ioutil.ReadDir(debugDir)
			require.NoError(t, err, `error reading %s after stopping [%v]`, debugDir, cmd)
			require.Equal(t, 1, len(files), `expected 1 file in %s after stopping [%v]`, debugDir, cmd)

			// get a listing of the debug archive
			debugArchive := filepath.Join(debugDir, files[0].Name())
			var dumpFiles []string
			err = archiver.Walk(filepath.Join(debugDir, files[0].Name()), func(f archiver.File) error {
				dumpFiles = append(dumpFiles, f.Name())
				return nil
			})
			require.NoError(t, err, `error reading debug archive %s after stopping [%v]`, debugArchive, cmd)

			// ensure the minimum contents exist
			for _, file := range []string{"config_dump.json", "stats.json"} {
				require.Contains(t, dumpFiles, file, `debug archive %s doesn't contain %s after stopping [%v]`, debugArchive, file, cmd)
			}
		})
	}
}
