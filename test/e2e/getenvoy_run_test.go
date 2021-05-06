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
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/log"

	reference "github.com/tetratelabs/getenvoy/pkg"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoytest"
	e2e "github.com/tetratelabs/getenvoy/test/e2e/util"
	utilenvoy "github.com/tetratelabs/getenvoy/test/e2e/util/envoy"
)

const terminateTimeout = 2 * time.Minute

// TestGetEnvoyRun runs the equivalent of "getenvoy run"
//
// See TestMain for general notes on about the test runtime.
func TestGetEnvoyRun(t *testing.T) {
	t.Parallel() // uses random ports so safe to run parallel

	c := getEnvoy(`run`).Arg(reference.Latest) // TODO allow implicit version #106
	// Below is the minimal config needed to run envoy
	c.Args(`--`, `--config-yaml`, `admin: {access_log_path: '/dev/stdout', address: {socket_address: {address: '127.0.0.1', port_value: 0}}}`)

	stdout, stderr, terminate := c.Start(t, terminateTimeout)

	// The underlying call is conditional to ensure errors that raise before we stop the server, stop it.
	deferredTerminate := terminate
	defer func() {
		if deferredTerminate != nil {
			deferredTerminate()
		}
	}()

	envoyWorkingDir := requireEnvoyWorkingDir(t, stdout, c)
	requireEnvoyReady(t, envoyWorkingDir, stderr, c)

	log.Infof(`stopping Envoy after running [%v]`, c)
	terminate()
	deferredTerminate = nil

	verifyDebugDump(t, envoyWorkingDir, c)
}

func requireEnvoyWorkingDir(t *testing.T, stdout io.Reader, c interface{}) string {
	stdoutLines := e2e.StreamLines(stdout).Named("stdout")
	log.Infof(`waiting for GetEnvoy to log working directory after running [%v]`, c)
	workingDirectoryPattern := regexp.MustCompile(`cd (.*)`)
	line, err := stdoutLines.FirstMatch(workingDirectoryPattern).Wait(10 * time.Minute) // give time to compile the extension
	require.NoError(t, err, `error parsing working directory from stdout of [%v]`, c)
	return workingDirectoryPattern.FindStringSubmatch(line)[1]
}

func requireEnvoyReady(t *testing.T, envoyWorkingDir string, stderr io.Reader, c interface{}) utilenvoy.AdminAPI {
	stderrLines := e2e.StreamLines(stderr).Named("stderr")

	log.Infof(`waiting for Envoy start-up to complete after running [%v]`, c)
	_, err := stderrLines.FirstMatch(regexp.MustCompile(`starting main dispatch loop`)).Wait(1 * time.Minute)
	require.NoError(t, err, `error parsing startup from stderr of [%v]`, c)

	adminAddressPath := filepath.Join(envoyWorkingDir, "admin-address.txt")
	adminAddress, err := os.ReadFile(adminAddressPath) //nolint:gosec
	require.NoError(t, err, `error reading admin address file %q after running [%v]`, adminAddressPath, c)

	log.Infof(`waiting for Envoy client to connect after running [%v]`, c)
	envoyClient, err := utilenvoy.NewClient(string(adminAddress))
	require.NoError(t, err, `error from envoy client %s after running [%v]`, adminAddress, c)
	require.Eventually(t, func() bool {
		ready, e := envoyClient.IsReady()
		return e == nil && ready
	}, 1*time.Minute, 100*time.Millisecond, `envoy client %s never ready after running [%v]`, adminAddress, c)

	return envoyClient
}

func verifyDebugDump(t *testing.T, workingDir string, c interface{}) {
	debugArchive := envoytest.RequireRestoreWorkingDir(t, workingDir, c)

	// ensure the minimum contents exist
	for _, filename := range []string{"config_dump.json", "stats.json"} {
		path := filepath.Join(workingDir, filename)
		f, err := os.Stat(path)
		require.NoError(t, err, `debug archive %s doesn't contain %s after stopping [%v]`, debugArchive, filename, c)
		require.NotEmpty(t, f.Size(), `%s was empty after stopping [%v]`, filename, c)
	}
}
