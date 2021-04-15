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

package envoytest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/mholt/archiver/v3"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/log"

	"github.com/tetratelabs/getenvoy/pkg/binary"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
)

// reference indicates the default Envoy version to be used for testing.
// This currently latest, but we may support a range at some point.
var reference = "standard:1.17.1"

var once sync.Once
var errorFetchingEnvoy error

// FetchEnvoyAndRun retrieves the Envoy indicated by reference only once. This is intended to be used with TestMain.
// In CI, you can execute this to obviate latency during test runs: "go run cmd/getenvoy/main.go fetch standard:1.17.1"
func FetchEnvoyAndRun(m *testing.M) {
	once.Do(func() {
		errorFetchingEnvoy = fetchEnvoy()
	})

	if errorFetchingEnvoy != nil {
		fmt.Println(errorFetchingEnvoy)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func fetchEnvoy() error {
	key, err := manifest.NewKey(reference)
	if err != nil {
		return fmt.Errorf("unable to make manifest key %v: %w", reference, err)
	}

	r, err := envoy.NewRuntime()
	if err != nil {
		return fmt.Errorf("unable to make new envoy runtime: %w", err)
	}

	if !r.AlreadyDownloaded(key) {
		location, err := manifest.Locate(key)
		if err != nil {
			return fmt.Errorf("unable to retrieve manifest from %v: %w", manifest.GetURL(), err)
		}

		err = r.Fetch(key, location)
		if err != nil {
			return fmt.Errorf("unable to retrieve binary from %v: %w", location, err)
		}
	}
	return nil
}

// RunKillOptions allows customization of Envoy lifecycle.
type RunKillOptions struct{ Bootstrap string }

// RequireRunTerminate executes envoy, waits for ready, sends sigint, waits for termination, then unarchives the debug
// directory. It should be used when you just want to cycle through an Envoy lifecycle.
//
// When the configPath parameter is non-empty, it becomes the "--config-path" argument to envoy.
func RequireRunTerminate(t *testing.T, r binary.Runner, configPath string) {
	key, err := manifest.NewKey(reference)
	require.NoError(t, err)
	var args []string
	if configPath != "" {
		args = append(args, "--config-path", configPath)
	}

	args = append(args,
		// Generate base id to allow concurrent envoys in tests. (minimum Envoy 1.15)
		"--use-dynamic-base-id",
		// Use ephemeral admin port to avoid test conflicts.
		// Enable admin access logging to help debug test failures. (minimum Envoy 1.12 for macOS support)
		"--config-yaml", "admin: {access_log_path: '/dev/stdout', address: {socket_address: {address: '127.0.0.1', port_value: 0}}}",
	)

	// This ensures on any panic the envoy process is terminated, which can prevent test hangs.
	deferredInterrupt := func() {
		r.(*envoy.Runtime).FakeInterrupt()
	}

	defer func() {
		if deferredInterrupt != nil {
			deferredInterrupt()
		}
	}()

	// Ensure we don't leave tar.gz files around after the test completes
	defer os.RemoveAll(r.DebugStore() + ".tar.gz") // nolint

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := r.Run(key, args); err != nil {
			log.Errorf("unable to run key %s: %v", key, err)
		}
		cancel()
	}()

	// Look for terminated or ready, so that we fail faster than polling for status ready
	expectedStatus := binary.StatusReady
	require.Eventually(t, func() bool {
		return r.Status() == expectedStatus || r.Status() == binary.StatusTerminated
	}, 30*time.Second, 100*time.Millisecond, "never achieved status(%d) or StatusTerminated", expectedStatus)
	require.Equal(t, expectedStatus, r.Status(), "never achieved status(%d)", expectedStatus)

	// Now, terminate the server.
	r.(*envoy.Runtime).FakeInterrupt()
	deferredInterrupt = nil

	select { // Await run completion
	case <-time.After(10 * time.Second):
		t.Fatal("Run never completed")
	case <-ctx.Done():
	}

	// RunPath deletes the debug store directory after making a tar.gz with the same name.
	// Restore it so assertions can read the contents later.
	e := archiver.Unarchive(r.DebugStore()+".tar.gz", filepath.Dir(r.DebugStore()))
	require.NoError(t, e, "error extracting DebugStore")
}
