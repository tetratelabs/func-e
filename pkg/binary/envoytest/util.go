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
	"math/rand"
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

// Reference indicates the default Envoy version to be used for testing.
// 1.12.7 is the last version to match the Istio version we are using (see go.mod)
var Reference = "standard:1.12.7"

var once sync.Once
var errorFetchingEnvoy error

// FetchEnvoyAndRun retrieves the Envoy indicated by Reference only once. This is intended to be used with TestMain.
// In CI, you can execute this to obviate latency during test runs: "go run cmd/getenvoy/main.go fetch standard:1.12.7"
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
	key, err := manifest.NewKey(Reference)
	if err != nil {
		return fmt.Errorf("unable to make manifest key %v: %w", Reference, err)
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
	Bootstrap        string
	ExpectedStatus   int
	RetainDebugStore bool
	SleepBeforeKill  time.Duration
}

// RequireRunKill executes envoy, waits for ready, sends sigint, waits for termination, then unarchives the debug directory.
// It should be used when you just want to cycle through an Envoy lifecycle.
func RequireRunKill(t *testing.T, r binary.Runner, options RunKillOptions) {
	key, _ := manifest.NewKey(Reference)
	var args []string
	if options.Bootstrap != "" {
		args = append(args, "-c", options.Bootstrap)
	}

	args = append(args,
		// Use ephemeral admin port to avoid test conflicts. Enable admin access logging to help debug test failures.
		"--config-yaml", "admin: {access_log_path: '/dev/stdout', address: {socket_address: {address: '127.0.0.1', port_value: 0}}}",
		// Generate base id to allow concurrent envoys in tests. When envoy is v1.15+, switch to --use-dynamic-base-id
		// This prevents "unable to bind domain socket with id=0" on Linux hosts.
		"--base-id", fmt.Sprintf("%d", rand.Int31()), // nolint it isn't important to use a secure random here
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

	time.Sleep(options.SleepBeforeKill)

	require.NotEqual(t, binary.StatusTerminated, r.Status(), "already StatusTerminated")

	r.SendSignal(syscall.SIGINT)
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
