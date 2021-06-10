// Copyright 2021 Tetrate
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
	"bufio"
	_ "embed" // We embed the config files to make them easier to copy
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/tar"
	"github.com/tetratelabs/getenvoy/internal/test/morerequire"
)

var (
	// staticFilesystemConfig shows Envoy reading a file referenced from the CWD
	//go:embed static-filesystem.yaml
	staticFilesystemConfig  []byte
	adminAddressPathPattern = regexp.MustCompile(`--admin-address-path ([^ ]+)`)
	envoyStartedLine        = "starting main dispatch loop"
)

// TestGetEnvoyRun runs the equivalent of "getenvoy run"
//
// See TestMain for general notes on about the test runtime.
func TestGetEnvoyRun(t *testing.T) {
	t.Parallel() // uses random ports so safe to run parallel

	// Below is the minimal config needed to run Envoy 1.18+, non-windows <1.18 need access_log_path: '/dev/stdout'
	args := []string{"run", "--config-yaml", "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}"}
	c := envoyRunTest(t, nil, args...)
	verifyDebugDump(t, filepath.Dir(c.adminAddressPath), c)
}

func TestGetEnvoyRun_StaticFilesystem(t *testing.T) {
	t.Parallel() // uses random ports so safe to run parallel

	revertTempWd := morerequire.RequireChdirIntoTemp(t)
	defer revertTempWd()

	require.NoError(t, os.WriteFile("envoy.yaml", staticFilesystemConfig, 0600))
	responseFromRunDirectory := []byte("foo")
	require.NoError(t, os.WriteFile("response.txt", responseFromRunDirectory, 0600))

	envoyRunTest(t, func(c *getEnvoy, a *adminClient) {
		mainURL, err := a.getMainListenerURL()
		require.NoError(t, err, `couldn't read mainURL after running [%v]`, c)

		body, err := httpGet(mainURL)
		require.NoError(t, err, `couldn't read %s after running [%v]`, mainURL, c)

		// If this passes, we know Envoy is running in the current directory, so can resolve relative configuration.
		require.Equal(t, responseFromRunDirectory, body, `unexpected content in %s after running [%v]`, mainURL, c)
	}, "run", "-c", "envoy.yaml")
}

// envoyRunTest runs the given args. If the process successfully starts, it blocks until the adminClient is available.
// The process is interrupted when the test completes, or a timeout occurs.
func envoyRunTest(t *testing.T, test func(*getEnvoy, *adminClient), args ...string) *getEnvoy {
	c, cancel := newGetEnvoy(args...)
	defer cancel()

	stdout, out := c.cmd.StdoutPipe()
	outLines := bufio.NewScanner(stdout)
	require.NoError(t, out)

	stderr, err := c.cmd.StderrPipe()
	require.NoError(t, err)
	errLines := bufio.NewScanner(stderr)

	err = c.cmd.Start()
	require.NoError(t, err)

	log.Printf(`waiting for getenvoy stdout to match %q after running [%v]`, adminAddressPathPattern, c)
	go func() {
		for outLines.Scan() {
			l := outLines.Text()
			fmt.Fprintln(os.Stdout, l) // echo to the person watching the test

			// getenvoy has unit tests that ensure --admin-address-path is always set. However, we can't read the
			// contents until Envoy creates it. This only ensures we know the file name.
			if strings.Contains(l, "--admin-address-path") {
				require.True(t, adminAddressPathPattern.MatchString(l), `error parsing admin address path from %s of [%v]`, l, c)
				c.adminAddressPath = adminAddressPathPattern.FindStringSubmatch(l)[1]
				log.Printf(`waiting for Envoy stdout to match %q after running [%v]`, envoyStartedLine, c)
			}
		}
	}()

	go func() {
		for errLines.Scan() {
			l := errLines.Text()
			fmt.Fprintln(os.Stderr, l) // echo to the person watching the test

			// When we get to this line, we can assume Envoy started properly. Run the test
			if strings.Contains(l, envoyStartedLine) {
				require.NotEmpty(t, c.adminAddressPath, "expected adminAddressPath to be set")
				go runTestAndInterruptEnvoy(t, c, test) // don't block stderr when the test is running
			}
		}
	}()

	err = c.cmd.Wait() // This won't hang forever because newGetEnvoy started it with a context timeout!
	require.NoError(t, err)
	return c
}

func runTestAndInterruptEnvoy(t *testing.T, c *getEnvoy, test func(*getEnvoy, *adminClient)) {
	defer func() {
		log.Printf(`shutting down Envoy after running [%v]`, c)
		_ = c.cmd.Process.Signal(syscall.SIGTERM)
	}()
	a := requireEnvoyReady(t, c.adminAddressPath, c)
	if test != nil {
		test(c, a)
	}
}

func requireEnvoyReady(t *testing.T, adminAddressPath string, c interface{}) *adminClient {
	adminAddress, err := os.ReadFile(adminAddressPath) //nolint:gosec
	require.NoError(t, err, `error reading admin address file %q after running [%v]`, adminAddressPath, c)

	log.Printf(`waiting for Envoy adminClient to connect after running [%v]`, c)
	envoyClient, err := newAdminClient(string(adminAddress))
	require.NoError(t, err, `error from Envoy adminClient %s after running [%v]`, adminAddress, c)
	require.Eventually(t, func() bool {
		ready, err := envoyClient.isReady()
		return err == nil && ready
	}, 1*time.Minute, 100*time.Millisecond, `Envoy adminClient %s never ready after running [%v]`, adminAddress, c)

	return envoyClient
}

func verifyDebugDump(t *testing.T, runDir string, c interface{}) {
	// Run deletes the working directory after making a tar.gz with the same name.
	// Restore it so assertions can read the contents later.
	runArchive := filepath.Join(runDir + ".tar.gz")
	defer os.Remove(runArchive) //nolint

	src, err := os.Open(runArchive)
	require.NoError(t, err, "error opening %s after shutdown [%v]", runArchive, c)
	err = tar.Untar(runDir, src)
	require.NoError(t, err, "error restoring %s from %s after shutdown [%v]", runDir, runArchive, c)

	// ensure the minimum contents exist
	for _, filename := range []string{"stdout.log", "stderr.log", "config_dump.json", "stats.json"} {
		path := filepath.Join(runDir, filename)
		f, err := os.Stat(path)
		require.NoError(t, err, `run archive %s doesn't contain %s after shutdown [%v]`, runArchive, filename, c)
		require.NotEmpty(t, f.Size(), `%s was empty after shutdown [%v]`, filename, c)
	}
}
