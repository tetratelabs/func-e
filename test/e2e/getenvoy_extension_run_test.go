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
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/mholt/archiver/v3"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/log"

	"github.com/tetratelabs/getenvoy/pkg/common"
	. "github.com/tetratelabs/getenvoy/pkg/test/morerequire"
	e2e "github.com/tetratelabs/getenvoy/test/e2e/util"
	utilenvoy "github.com/tetratelabs/getenvoy/test/e2e/util/envoy"
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

			// "getenvoy extension run" only returns stdout because `docker run -t` redirects stderr to stdout.
			c := getEnvoy("extension run --envoy-options '-l trace'").Args(getToolchainContainerOptions()...)
			_, stderr, terminate := c.Start(t, terminateTimeout)

			// The underlying call is conditional to ensure errors that raise before we stop the server, stop it.
			deferredTerminate := terminate
			defer func() {
				deferredTerminate()
			}()

			stderrLines := e2e.StreamLines(stderr).Named("stderr")

			log.Infof(`waiting for Envoy Admin address to get logged after running [%v]`, c)
			adminAddressPattern := regexp.MustCompile(`admin address: ([^:]+:[0-9]+)`)
			line, err := stderrLines.FirstMatch(adminAddressPattern).Wait(10 * time.Minute) // give time to compile the extension
			require.NoError(t, err, `error parsing admin address from stderr of [%v]`, c)
			adminAddress := adminAddressPattern.FindStringSubmatch(line)[1]

			log.Infof(`waiting for Envoy start-up to complete after running [%v]`, c)
			_, err = stderrLines.FirstMatch(regexp.MustCompile(`starting main dispatch loop`)).Wait(1 * time.Minute)
			require.NoError(t, err, `error parsing startup from stderr of [%v]`, c)

			log.Infof(`waiting for Envoy client to connect after running [%v]`, c)
			envoyClient, err := utilenvoy.NewClient(adminAddress)
			require.NoError(t, err, `error from envoy client %s after running [%v]`, adminAddress, c)
			require.Eventually(t, func() bool {
				ready, e := envoyClient.IsReady()
				return e == nil && ready
			}, 1*time.Minute, 100*time.Millisecond, `envoy client %s never ready after running [%v]`, adminAddress, c)

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
			}, 1*time.Minute, 100*time.Millisecond, `wasm stats never found after running [%v]`, adminAddress, c)

			log.Infof(`stopping Envoy after running [%v]`, c)
			terminate()
			deferredTerminate = func() {
				// no-op as we already terminated
			}

			// verify the debug dump of Envoy state has been taken
			files, err := os.ReadDir(debugDir)
			require.NoError(t, err, `error reading %s after stopping [%v]`, debugDir, c)
			require.Equal(t, 1, len(files), `expected 1 file in %s after stopping [%v]`, debugDir, c)
			defer func() {
				e := os.RemoveAll(debugDir)
				require.NoError(t, e, `error removing debug dir %s after stopping [%v]`, debugDir, c)
			}()

			// get a listing of the debug archive
			debugArchive := filepath.Join(debugDir, files[0].Name())
			var dumpFiles []string
			err = archiver.Walk(filepath.Join(debugDir, files[0].Name()), func(f archiver.File) error {
				dumpFiles = append(dumpFiles, f.Name())
				return nil
			})
			require.NoError(t, err, `error reading debug archive %s after stopping [%v]`, debugArchive, c)

			// ensure the minimum contents exist
			for _, file := range []string{"config_dump.json", "stats.json"} {
				require.Contains(t, dumpFiles, file, `debug archive %s doesn't contain %s after stopping [%v]`, debugArchive, file, c)
			}
		})
	}
}

// backupDebugDir backs up ${GETENVOY_HOME}/debug in case the test hasn't overridden it and the developer has existing
// data there. The function returned reverts this directory.
//
// Typically, this will run in the default ~/.getenvoy directory, as a means to avoid re-downloads of files such as
// .getenvoy/builds/standard/1.17.0/darwin/bin/envoy (~100MB)
//
// While CI usually overrides the `HOME` variable with E2E_TOOLCHAIN_CONTAINER_OPTIONS, a developer may be
// running this on their laptop. To avoid clobbering their old debug data, backup the
func backupDebugDir(t *testing.T) (string, func()) {
	debugDir := filepath.Join(common.HomeDir, "debug")

	if _, err := os.Lstat(debugDir); err != nil && os.IsNotExist(err) {
		return debugDir, func() {} // do nothing on remove, if there was no debug backup
	}

	// get a name of a new temp directory, which we'll rename the existing debug to
	backupDir, _ := RequireNewTempDir(t)
	err := os.RemoveAll(backupDir)
	require.NoError(t, err, `error removing temp directory: %s`, backupDir)

	err = os.Rename(debugDir, backupDir)
	require.NoError(t, err, `error renaming debug dir %s to %s`, debugDir, backupDir)

	return debugDir, func() {
		err = os.Rename(backupDir, debugDir)
		require.NoError(t, err, `error renaming backup dir %s to %s`, debugDir, backupDir)
	}
}
