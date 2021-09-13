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

package moreos

import (
	"bytes"
	"embed"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/tetratelabs/func-e/internal/test/fakebinary"

	"github.com/stretchr/testify/require"
)

var (
	// fakeFuncESrc is a test source file used to simulate how func-e manages its child process
	//go:embed testdata/fake_func-e.go
	fakeFuncESrc []byte // Embedding the fakeFuncESrc is easier than file I/O and ensures it doesn't skew coverage

	// Include the source imported by fakeFuncESrc directly and indirectly
	//go:embed moreos.go
	//go:embed proc_*.go
	moreosSrcDir embed.FS
)

// Test_X tests multiple features include moreos.EnsureProcessDone and moreos.ProcessGroupAttr to ensure X occurs when Y
func Test_X(t *testing.T) {
	tempDir := t.TempDir()

	// Build a fake envoy and pass the ENV hint so that fake func-e uses it
	fakeEnvoy := filepath.Join(tempDir, "envoy"+Exe)
	fakebinary.RequireFakeEnvoy(t, fakeEnvoy)
	t.Setenv("ENVOY_PATH", fakeEnvoy)

	fakeFuncE := filepath.Join(tempDir, "func-e"+Exe)
	requireFakeFuncE(t, fakeFuncE)

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	// intentionally no args, as we expect it to not block and also fail
	cmd := exec.Command(fakeFuncE, "run")
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	require.Error(t, cmd.Run())
	require.Equal(t, Sprintf("starting: %s\n", fakeEnvoy), stdout.String())
	require.Contains(t, stderr.String(), Sprintf("envoy exited with status: 1\n"))
}

// requireFakeFuncE builds a func-e binary only depends on fakeFuncESrc and the sources in this package.
// This is used to test integrated use of tools like ProcessGroupAttr without mixing unrelated concerns or dependencies.
func requireFakeFuncE(t *testing.T, path string) {
	workDir := t.TempDir()

	// Copy the sources needed for fake func-e, but nothing else
	moreosDir := filepath.Join(workDir, "internal", "moreos")
	require.NoError(t, os.MkdirAll(moreosDir, 0700))
	moreosSrcs, err := moreosSrcDir.ReadDir(".")
	require.NoError(t, err)
	for _, src := range moreosSrcs {
		data, err := moreosSrcDir.ReadFile(src.Name())
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(moreosDir, src.Name()), data, 0600))
	}

	fakeFuncEBin := fakebinary.RequireBuildFakeBinary(t, workDir, "func-e", fakeFuncESrc)
	require.NoError(t, os.WriteFile(path, fakeFuncEBin, 0700)) //nolint:gosec
}
