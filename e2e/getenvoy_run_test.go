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

package e2e

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/tar"
)

const terminateTimeout = 2 * time.Minute

// TestGetEnvoyRun runs the equivalent of "getenvoy run"
//
// See TestMain for general notes on about the test runtime.
func TestGetEnvoyRun(t *testing.T) {
	t.Parallel() // uses random ports so safe to run parallel

	c := getEnvoy(`run`)
	// Below is the minimal config needed to run envoy
	c.args("--config-yaml", "admin: {access_log_path: '/dev/stdout', address: {socket_address: {address: '127.0.0.1', port_value: 0}}}")

	stdout, stderr, terminate := c.start(t, terminateTimeout)

	// The underlying call is conditional to ensure errors that raise before we stop the server, stop it.
	deferredTerminate := terminate
	defer func() {
		if deferredTerminate != nil {
			deferredTerminate()
		}
	}()

	envoyWorkingDir := requireEnvoyWorkingDir(t, stdout, c)
	requireEnvoyReady(t, envoyWorkingDir, stderr, c)

	log.Printf(`stopping Envoy after running [%v]`, c)
	terminate()
	deferredTerminate = nil

	verifyDebugDump(t, envoyWorkingDir, c)
}

func requireEnvoyWorkingDir(t *testing.T, stdout io.Reader, c interface{}) string {
	stdoutLines := streamLines(stdout).named("stdout")
	log.Printf(`waiting for GetEnvoy to log working directory after running [%v]`, c)
	workingDirectoryPattern := regexp.MustCompile(`working directory: (.*)`)
	line, err := stdoutLines.FirstMatch(workingDirectoryPattern).Wait(1 * time.Minute)
	require.NoError(t, err, `error parsing working directory from stdout of [%v]`, c)
	return workingDirectoryPattern.FindStringSubmatch(line)[1]
}

func requireEnvoyReady(t *testing.T, envoyWorkingDir string, stderr io.Reader, c interface{}) adminAPI {
	stderrLines := streamLines(stderr).named("stderr")

	log.Printf(`waiting for Envoy start-up to complete after running [%v]`, c)
	_, err := stderrLines.FirstMatch(regexp.MustCompile(`starting main dispatch loop`)).Wait(1 * time.Minute)
	require.NoError(t, err, `error parsing startup from stderr of [%v]`, c)

	adminAddressPath := filepath.Join(envoyWorkingDir, "admin-address.txt")
	adminAddress, err := os.ReadFile(adminAddressPath) //nolint:gosec
	require.NoError(t, err, `error reading admin address file %q after running [%v]`, adminAddressPath, c)

	log.Printf(`waiting for Envoy client to connect after running [%v]`, c)
	envoyClient, err := newClient(string(adminAddress))
	require.NoError(t, err, `error from envoy client %s after running [%v]`, adminAddress, c)
	require.Eventually(t, func() bool {
		ready, err := envoyClient.isReady()
		return err == nil && ready
	}, 1*time.Minute, 100*time.Millisecond, `envoy client %s never ready after running [%v]`, adminAddress, c)

	return envoyClient
}

func verifyDebugDump(t *testing.T, workingDir string, c interface{}) {
	// Run deletes the working directory after making a tar.gz with the same name.
	// Restore it so assertions can read the contents later.
	runArchive := filepath.Join(workingDir + ".tar.gz")
	defer os.Remove(runArchive) //nolint

	src, err := os.Open(runArchive)
	require.NoError(t, err, "error opening %s after stopping [%v]", runArchive, c)
	err = tar.Untar(workingDir, src)
	require.NoError(t, err, "error restoring %s from %s after stopping [%v]", workingDir, runArchive, c)

	// ensure the minimum contents exist
	for _, filename := range []string{"config_dump.json", "stats.json"} {
		path := filepath.Join(workingDir, filename)
		f, err := os.Stat(path)
		require.NoError(t, err, `run archive %s doesn't contain %s after stopping [%v]`, runArchive, filename, c)
		require.NotEmpty(t, f.Size(), `%s was empty after stopping [%v]`, filename, c)
	}
}
