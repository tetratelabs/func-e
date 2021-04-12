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
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
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
// 1.15.3 is the last version to match the Istio version we are using in go.mod:
//   https://github.com/istio/istio/blob/1.7.8/istio.deps ->
//   https://github.com/istio/proxy/blob/d68172a37cb87c52d683a906bd2fba90060f8d82/WORKSPACE ->
//   https://github.com/istio/envoy/commit/4abbbc0394b247d4ea37fd8f9137732c3d0a2b91 ->
//   https://github.com/istio/envoy/pull/290 -> ~ 1.15
var reference = "standard:1.15.3"

var once sync.Once
var errorFetchingEnvoy error

// FetchEnvoyAndRun retrieves the Envoy indicated by reference only once. This is intended to be used with TestMain.
// In CI, you can execute this to obviate latency during test runs: "go run cmd/getenvoy/main.go fetch standard:1.15.3"
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
type RunKillOptions struct {
	Bootstrap            string
	ExpectedStatus       int
	RetainDebugStore     bool
	SleepBeforeTerminate time.Duration
}

// RequireRunTerminate executes envoy, waits for ready, sends sigint, waits for termination, then unarchives the debug directory.
// It should be used when you just want to cycle through an Envoy lifecycle.
func RequireRunTerminate(t *testing.T, r binary.Runner, options RunKillOptions) {
	key, err := manifest.NewKey(reference)
	require.NoError(t, err)
	var args []string
	if options.Bootstrap != "" {
		args = append(args, "-c", options.Bootstrap)
	}

	args = append(args,
		// Use ephemeral admin port to avoid test conflicts.
		// Enable admin access logging to help debug test failures. (minimum Envoy 1.12 for macOS support)
		"--config-yaml", "admin: {access_log_path: '/dev/stdout', address: {socket_address: {address: '127.0.0.1', port_value: 0}}}",
		// Generate base id to allow concurrent envoys in tests. (minimum Envoy 1.15)
		"--use-dynamic-base-id",
	)
	// Allows us the status checker to read the resolved admin port after envoy starts
	envoy.EnableAdminAddressDetection(r.(*envoy.Runtime))

	// This ensures on any panic the envoy process is terminated, which can prevent test hangs.
	deferredTerminate := func() {
		// envoy.waitForTerminationSignals() registers SIGINT and SIGTERM
		r.SendSignal(syscall.SIGTERM)
	}

	defer func() {
		if deferredTerminate != nil {
			deferredTerminate()
		}
	}()

	// Ensure we don't leave tar.gz files around after the test completes
	defer os.RemoveAll(r.DebugStore() + ".tar.gz") // nolint

	go func() {
		if err := r.Run(key, args); err != nil {
			log.Errorf("unable to run key %s: %v", key, err)
		}
	}()

	// Look for terminated or ready, so that we fail faster than polling for status ready
	expectedStatus := binary.StatusReady
	if options.ExpectedStatus > 0 {
		expectedStatus = options.ExpectedStatus
	}
	require.Eventually(t, func() bool {
		return r.Status() == expectedStatus || r.Status() == binary.StatusTerminated
	}, 30*time.Second, 100*time.Millisecond, "never achieved status(%d) or StatusTerminated", expectedStatus)
	require.Equal(t, expectedStatus, r.Status(), "never achieved status(%d)", expectedStatus)

	time.Sleep(options.SleepBeforeTerminate)

	require.NotEqual(t, binary.StatusTerminated, r.Status(), "already StatusTerminated")

	r.SendSignal(syscall.SIGTERM)
	require.Eventually(t, func() bool {
		return r.Status() == binary.StatusTerminated
	}, 10*time.Second, 50*time.Millisecond, "never achieved StatusTerminated")

	deferredTerminate = nil // We succeeded, so no longer need to kill the envoy process

	// RunPath deletes the debug store directory after making a tar.gz with the same name.
	// Restore it so assertions can read the contents later.
	if options.RetainDebugStore {
		err := archiver.Unarchive(r.DebugStore()+".tar.gz", filepath.Dir(r.DebugStore()))
		require.NoError(t, err, "error extracting DebugStore")
	}
}
