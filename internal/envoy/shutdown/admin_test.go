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

package shutdown

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/envoy"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/test"
)

func TestEnableEnvoyAdminDataCollection(t *testing.T) {
	runDir := t.TempDir()

	require.NoError(t, runWithShutdownHook(t, runDir, enableEnvoyAdminDataCollection))

	for _, filename := range adminAPIPaths {
		path := filepath.Join(runDir, filename)
		f, err := os.Stat(path)
		require.NoError(t, err, "error stating %v", path)
		require.NotEmpty(t, f.Size(), "file %v was empty", path)
	}
}

// runWithShutdownHook is like RequireRun, except invokes the hook on shutdown
func runWithShutdownHook(t *testing.T, runDir string, hook func(r *envoy.Runtime) error) error {
	fakeEnvoy := filepath.Join(runDir, "envoy"+moreos.Exe)
	test.RequireFakeEnvoy(t, fakeEnvoy)

	o := &globals.RunOpts{EnvoyPath: fakeEnvoy, RunDir: runDir, DontArchiveRunDir: true}

	stderr := new(bytes.Buffer)
	r := envoy.NewRuntime(o)
	r.Out = io.Discard
	r.Err = stderr
	require.NoError(t, hook(r))

	return test.RequireRun(t, func() {
		fakeInterrupt := r.FakeInterrupt
		if fakeInterrupt != nil {
			fakeInterrupt()
		}
	}, r, stderr, "-c", "envoy.yaml")
}
