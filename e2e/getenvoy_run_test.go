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
	_ "embed" // We embed the config files to make them easier to copy
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/tar"
	"github.com/tetratelabs/getenvoy/internal/test/morerequire"
)

const shutdownTimeout = 2 * time.Minute

// TestGetEnvoyRun runs the equivalent of "getenvoy run"
//
// See TestMain for general notes on about the test runtime.
func TestGetEnvoyRun(t *testing.T) {
	t.Parallel() // uses random ports so safe to run parallel

	// Below is the minimal config needed to run envoy
	c := getEnvoy(`run`, "--config-yaml", "admin: {access_log_path: '/dev/stdout', address: {socket_address: {address: '127.0.0.1', port_value: 0}}}")

	stdout, stderr, shutdown := c.start(t, shutdownTimeout)

	// The underlying call is conditional to ensure errors that raise before we stop the server, stop it.
	deferredShutdown := shutdown
	defer func() {
		if deferredShutdown != nil {
			deferredShutdown()
		}
	}()

	runDir := requireRunDir(t, stdout, c)
	requireEnvoyReady(t, runDir, stderr, c)

	log.Printf(`shutting down Envoy after running [%v]`, c)
	shutdown()
	deferredShutdown = nil

	verifyDebugDump(t, runDir, c)
}

// staticFilesystemConfig shows envoy reading a file referenced from the CWD
//go:embed static-filesystem.yaml
var staticFilesystemConfig []byte

func TestGetEnvoyRun_StaticFilesystem(t *testing.T) {
	t.Parallel() // uses random ports so safe to run parallel

	revertTempWd := morerequire.RequireChdirIntoTemp(t)
	defer revertTempWd()

	require.NoError(t, os.WriteFile("envoy.yaml", staticFilesystemConfig, 0600))
	responseFromWorkingDirectory := []byte("foo")
	require.NoError(t, os.WriteFile("response.txt", responseFromWorkingDirectory, 0600))
	c := getEnvoy(`run`, "-c", "envoy.yaml")

	stdout, stderr, shutdown := c.start(t, shutdownTimeout)
	defer shutdown()

	runDir := requireRunDir(t, stdout, c)
	admin := requireEnvoyReady(t, runDir, stderr, c)

	mainURL, err := admin.getMainListenerURL()
	require.NoError(t, err, `couldn't read mainURL after running [%v]`, c)

	body, err := httpGet(mainURL)
	require.NoError(t, err, `couldn't read %s after running [%v]`, mainURL, c)

	// If this passes, we know Envoy is running in the current directory, so can resolve relative configuration.
	require.Equal(t, responseFromWorkingDirectory, body, `unexpected content in %s after running [%v]`, mainURL, c)
}

func requireRunDir(t *testing.T, stdout io.Reader, c interface{}) string {
	stdoutLines := streamLines(stdout).named("stdout")
	log.Printf(`waiting for GetEnvoy to log admin address path after running [%v]`, c)
	adminAddressPathPattern := regexp.MustCompile(`--admin-address-path ([^ ]+)`)
	line, err := stdoutLines.FirstMatch(adminAddressPathPattern).Wait(1 * time.Minute)
	require.NoError(t, err, `error parsing admin address path from stdout of [%v]`, c)
	return filepath.Dir(adminAddressPathPattern.FindStringSubmatch(line)[1])
}

func requireEnvoyReady(t *testing.T, runDir string, stderr io.Reader, c interface{}) *adminClient {
	stderrLines := streamLines(stderr).named("stderr")

	log.Printf(`waiting for Envoy start-up to complete after running [%v]`, c)
	_, err := stderrLines.FirstMatch(regexp.MustCompile(`starting main dispatch loop`)).Wait(1 * time.Minute)
	require.NoError(t, err, `error parsing startup from stderr of [%v]`, c)

	adminAddressPath := filepath.Join(runDir, "admin-address.txt")
	adminAddress, err := os.ReadFile(adminAddressPath) //nolint:gosec
	require.NoError(t, err, `error reading admin address file %q after running [%v]`, adminAddressPath, c)

	log.Printf(`waiting for Envoy adminClient to connect after running [%v]`, c)
	envoyClient, err := newAdminClient(string(adminAddress))
	require.NoError(t, err, `error from envoy adminClient %s after running [%v]`, adminAddress, c)
	require.Eventually(t, func() bool {
		ready, err := envoyClient.isReady()
		return err == nil && ready
	}, 1*time.Minute, 100*time.Millisecond, `envoy adminClient %s never ready after running [%v]`, adminAddress, c)

	return envoyClient
}

func verifyDebugDump(t *testing.T, workingDir string, c interface{}) {
	// Run deletes the working directory after making a tar.gz with the same name.
	// Restore it so assertions can read the contents later.
	runArchive := filepath.Join(workingDir + ".tar.gz")
	defer os.Remove(runArchive) //nolint

	src, err := os.Open(runArchive)
	require.NoError(t, err, "error opening %s after shutdown [%v]", runArchive, c)
	err = tar.Untar(workingDir, src)
	require.NoError(t, err, "error restoring %s from %s after shutdown [%v]", workingDir, runArchive, c)

	// ensure the minimum contents exist
	for _, filename := range []string{"stdout.log", "stderr.log", "config_dump.json", "stats.json"} {
		path := filepath.Join(workingDir, filename)
		f, err := os.Stat(path)
		require.NoError(t, err, `run archive %s doesn't contain %s after shutdown [%v]`, runArchive, filename, c)
		require.NotEmpty(t, f.Size(), `%s was empty after shutdown [%v]`, filename, c)
	}
}
