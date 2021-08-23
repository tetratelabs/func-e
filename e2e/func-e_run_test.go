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
	"context"
	_ "embed" // We embed the config files to make them easier to copy
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/tar"
	"github.com/tetratelabs/func-e/internal/test/morerequire"
)

var (
	// staticFilesystemConfig shows Envoy reading a file referenced from the current directory
	//go:embed static-filesystem.yaml
	staticFilesystemConfig  []byte
	adminAddressPathPattern = regexp.MustCompile(`--admin-address-path ([^ ]+)`)
	envoyStartedLine        = "starting main dispatch loop"
	// minRunArgs is the minimal config needed to run Envoy 1.18+, non-windows <1.18 need access_log_path: '/dev/stdout'
	minRunArgs = []string{"run", "--config-yaml", "admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}"}
)

// TestFuncERun runs the equivalent of "func-e run"
//
// See TestMain for general notes on about the test runtime.
func TestFuncERun(t *testing.T) {
	t.Parallel() // uses random ports so safe to run parallel

	cmd, cleanup := envoyRunTest(t, nil, minRunArgs...)
	defer cleanup()

	verifyRunArchive(t, cmd)
}

func TestFuncERun_StaticFilesystem(t *testing.T) {
	t.Parallel() // uses random ports so safe to run parallel

	revertWd := morerequire.RequireChdir(t, t.TempDir())
	defer revertWd()

	require.NoError(t, os.WriteFile("envoy.yaml", staticFilesystemConfig, 0600))
	responseFromRunDirectory := []byte("foo")
	require.NoError(t, os.WriteFile("response.txt", responseFromRunDirectory, 0600))

	_, cleanup := envoyRunTest(t, func(ctx context.Context, c *funcE, a *adminClient) {
		mainURL, err := a.getMainListenerURL(ctx)
		require.NoError(t, err, "couldn't read mainURL after running [%v]", c)

		body, err := httpGet(ctx, mainURL)
		require.NoError(t, err, "couldn't read %s after running [%v]", mainURL, c)

		// If this passes, we know Envoy is running in the current directory, so can resolve relative configuration.
		require.Equal(t, responseFromRunDirectory, body, "unexpected content in %s after running [%v]", mainURL, c)
	}, "run", "-c", "envoy.yaml")

	cleanup()
}

// envoyRunTest runs the given args and the test function once envoy is available. This returns the command and a
// function to remove the run archive.
//
// If the process successfully starts, this blocks until the adminClient is available.  The process is interrupted when
// the test completes, or a timeout occurs.
func envoyRunTest(t *testing.T, test func(context.Context, *funcE, *adminClient), args ...string) (*funcE, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), runTimeout)
	defer cancel()

	c := newFuncE(ctx, args...)

	stdout, out := c.cmd.StdoutPipe()
	outLines := bufio.NewScanner(stdout)
	require.NoError(t, out)

	stderr, err := c.cmd.StderrPipe()
	require.NoError(t, err)
	errLines := bufio.NewScanner(stderr)

	err = c.cmd.Start()
	require.NoError(t, err)

	log.Printf("waiting for func-e stdout to match %q after running [%v]", adminAddressPathPattern, c)
	var adminAddressPath string
	go func() {
		for outLines.Scan() {
			l := outLines.Text()
			fmt.Fprintln(os.Stdout, l) // echo to the person watching the test

			// func-e has unit tests that ensure --admin-address-path is always set. However, we can't read the
			// contents until Envoy creates it. This only ensures we know the file name.
			if adminAddressPathPattern.MatchString(l) {
				adminAddressPath = adminAddressPathPattern.FindStringSubmatch(l)[1]
				c.runDir = filepath.Dir(adminAddressPath)
				log.Printf("waiting for Envoy stdout to match %q after running [%v]", envoyStartedLine, c)
			}
		}
		log.Printf("done stdout loop for func-e after running [%v]", c)
	}()

	go func() {
		for errLines.Scan() {
			l := errLines.Text()
			fmt.Fprintln(os.Stderr, l) // echo to the person watching the test

			// When we get to this line, we can assume Envoy started properly. Run the test
			if strings.Contains(l, envoyStartedLine) {
				require.NotEmpty(t, c.runDir, "expected adminAddressPath to be set")
				requireEnvoyPid(t, c)                        // check here to avoid race condition
				go runTestAndInterruptEnvoy(ctx, t, c, test) // don't block printing stderr!
			}
		}
		log.Printf("done stderr loop for func-e after running [%v]", c)
	}()

	err = c.cmd.Wait() // This won't hang forever because newFuncE started it with a context timeout!
	log.Printf("done waiting for func-e after running [%v]", c)
	require.NoError(t, err)

	// Ensure the Envoy process was terminated
	_, err = process.NewProcessWithContext(ctx, c.envoyPid) // because os.FindProcess is no-op in Linux!
	require.Error(t, err, "expected func-e to terminate Envoy after running [%v]", c)

	return c, func() {
		// this may not be present if the process was kill -9'd so don't error
		os.Remove(c.runDir + ".tar.gz") //nolint
	}
}

func runTestAndInterruptEnvoy(ctx context.Context, t *testing.T, c *funcE, test func(context.Context, *funcE, *adminClient)) {
	defer func() {
		log.Printf("interrupting func-e after running [%v]", c)
		require.NoError(t, moreos.Interrupt(c.cmd.Process), "error shutting down Envoy after running [%v]", c)
	}()
	a := requireEnvoyReady(ctx, t, c)
	if test != nil {
		test(ctx, c, a)
	}
}

func requireEnvoyReady(ctx context.Context, t *testing.T, c *funcE) *adminClient {
	adminAddressPath := filepath.Join(c.runDir, "admin-address.txt")
	adminAddress, err := os.ReadFile(adminAddressPath) //nolint:gosec
	require.NoError(t, err, "error reading admin address file %q after running [%v]", adminAddressPath, c)

	log.Printf("waiting for Envoy adminClient to connect after running [%v]", c)
	envoyClient, err := newAdminClient(string(adminAddress))
	require.NoError(t, err, "error from Envoy adminClient %s after running [%v]", adminAddress, c)
	require.Eventually(t, func() bool {
		return envoyClient.isReady(ctx)
	}, 1*time.Minute, 100*time.Millisecond, "Envoy adminClient %s never ready after running [%v]", adminAddress, c)

	return envoyClient
}

// requireEnvoyPid ensures $runDir/envoy.pid was written and is valid
func requireEnvoyPid(t *testing.T, c *funcE) {
	pidPath := filepath.Join(c.runDir, "envoy.pid")
	pidTxt, err := os.ReadFile(pidPath)
	require.NoError(t, err, "couldn't read %s after running [%v]", pidPath, c)
	pid, err := strconv.Atoi(string(pidTxt))
	require.NoError(t, err, "invalid Envoy pid in %s after running [%v]", pidPath, c)
	require.Greater(t, pid, 1, "invalid Envoy pid %s after running [%v]", pid, c)
	c.envoyPid = int32(pid)
}

// Run deletes the run directory after making a tar.gz with the same name. This extracts it and tests the contents.
func verifyRunArchive(t *testing.T, c *funcE) {
	runDir := t.TempDir()

	runArchive := c.runDir + ".tar.gz"
	src, err := os.Open(runArchive)
	require.NoError(t, err, "error opening %s after shutdown [%v]", runArchive, c)

	err = tar.Untar(runDir, src)
	require.NoError(t, err, "error restoring %s from %s after shutdown [%v]", runDir, runArchive, c)

	// ensure the minimum contents exist
	for _, filename := range []string{"stdout.log", "stderr.log", "config_dump.json", "stats.json"} {
		path := filepath.Join(runDir, filename)
		f, err := os.Stat(path)
		require.NoError(t, err, "run archive %s doesn't contain %s after shutdown [%v]", runArchive, filename, c)
		require.NotEmpty(t, f.Size(), "%s was empty after shutdown [%v]", filename, c)
	}
}
