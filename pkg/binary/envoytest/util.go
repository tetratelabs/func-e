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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mholt/archiver/v3"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/globals"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

// RunAndTerminateWithDebug is like RequireRunTerminate, except returns a directory populated by the debug plugin.
func RunAndTerminateWithDebug(t *testing.T, debugDir string, debug func(r *envoy.Runtime) error, args ...string) string {
	fakeEnvoy, removeFakeEnvoy := morerequire.RequireCaptureScript(t, "envoy")
	defer removeFakeEnvoy()

	fakeTimestamp := "1619574747231823000"
	o := &globals.RunOpts{EnvoyPath: fakeEnvoy, WorkingDir: filepath.Join(debugDir, fakeTimestamp)}
	// InitializeRunOpts creates this directory in a real command run
	require.NoError(t, os.MkdirAll(o.WorkingDir, 0750))

	r := envoy.NewRuntime(o)

	e := debug(r)
	require.NoError(t, e)

	RequireRunTerminate(t, nil, r, args...)
	RequireRestoreWorkingDir(t, o.WorkingDir, r)
	return o.WorkingDir
}

// RequireRunTerminate executes Run on the given Runtime and terminates it after starting.
func RequireRunTerminate(t *testing.T, terminate func(r *envoy.Runtime), r *envoy.Runtime, args ...string) {
	if terminate == nil {
		terminate = func(r *envoy.Runtime) {
			fakeInterrupt := r.FakeInterrupt
			if fakeInterrupt != nil {
				fakeInterrupt()
			}
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := r.Run(ctx, args); err != nil {
			t.Errorf("unable to run %v: %v", r, err)
			return
		}
		cancel()
	}()

	require.Eventually(t, func() bool {
		_, err := r.GetPid()
		return err == nil
	}, 1*time.Second, 100*time.Millisecond, "never started process")

	terminate(r)

	select { // Await run completion
	case <-time.After(10 * time.Second):
		t.Fatal("Run never completed")
	case <-ctx.Done():
	}
}

// RequireRestoreWorkingDir restores the working directory from the debug archive and returns the archive name.
func RequireRestoreWorkingDir(t *testing.T, workingDir string, c interface{}) string {
	// Run deletes the debug store directory after making a tar.gz with the same name.
	// Restore it so assertions can read the contents later.
	debugArchive := filepath.Join(workingDir + ".tar.gz")
	defer os.Remove(debugArchive) //nolint

	e := archiver.Unarchive(debugArchive, filepath.Dir(workingDir)) // Dir strips the RunID directory name
	require.NoError(t, e, "error restoring %s from %s after stopping [%v]", workingDir, debugArchive, c)
	return debugArchive
}
