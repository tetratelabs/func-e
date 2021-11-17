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

// Package fakebinary as using morerequire introduces build cycles
package fakebinary

import (
	_ "embed" // Embedding the fakeEnvoySrc is easier than file I/O and ensures it doesn't skew coverage
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// exe is like moreos.Exe, except if we used that it would make a build cycle.
var exe = func() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}()

var (
	// fakeEnvoySrc is a test source file used to simulate Envoy console output and signal processing.
	// This has no other source dependencies.
	//go:embed testdata/fake_envoy.go
	fakeEnvoySrc []byte
	// fakeEnvoyBin is the compiled code of fakeEnvoySrc which will be runtime.GOOS dependent.
	fakeEnvoyBin   []byte
	builtFakeEnvoy sync.Once
)

// RequireFakeEnvoy writes fakeEnvoyBin to the given path. This is embedded here because it is reused in many places.
func RequireFakeEnvoy(t *testing.T, path string) {
	builtFakeEnvoy.Do(func() {
		fakeEnvoyBin = RequireBuildFakeBinary(t, t.TempDir(), "envoy", fakeEnvoySrc)
	})
	require.NoError(t, os.WriteFile(path, fakeEnvoyBin, 0o700)) //nolint:gosec
}

// RequireBuildFakeBinary builds a fake binary and returns its contents.
func RequireBuildFakeBinary(t *testing.T, workDir, name string, mainSrc []byte) []byte {
	goBin := requireGoBin(t)

	bin := name + exe
	goArgs := []string{"build", "-o", bin, "main.go"}
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "main.go"), mainSrc, 0o600))

	// Don't allow any third party dependencies for now.
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "go.mod"),
		[]byte("module github.com/tetratelabs/func-e\n\ngo 1.17\n"), 0o600))

	cmd := exec.Command(goBin, goArgs...) //nolint:gosec
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "couldn't compile %s: %s", bin, string(out))
	bytes, err := os.ReadFile(filepath.Join(workDir, bin)) //nolint:gosec
	require.NoError(t, err)
	return bytes
}

func requireGoBin(t *testing.T) string {
	binName := "go" + exe
	goBin := filepath.Join(runtime.GOROOT(), "bin", binName)
	if _, err := os.Stat(goBin); err == nil {
		return goBin
	}
	// Now, search the path
	goBin, err := exec.LookPath(binName)
	require.NoError(t, err, "couldn't find %s in the PATH", goBin)
	return goBin
}
