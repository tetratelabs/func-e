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

package test

import (
	"bufio"
	"context"
	_ "embed" // Embedding the fakeEnvoySrc is easier than file I/O and ensures it doesn't skew coverage
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/test/morerequire"
)

// Runner allows us to not introduce dependency cycles on envoy.Runtime
type Runner interface {
	Run(ctx context.Context, args []string) error
}

// RequireRun executes Run on the given Runtime and calls shutdown after it started.
func RequireRun(t *testing.T, shutdown func(), r Runner, stderr io.Reader, args ...string) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	if shutdown == nil {
		shutdown = cancel
	}
	go func() {
		err = r.Run(ctx, args)
		cancel()
	}()

	reader := bufio.NewReader(stderr)
	require.Eventually(t, func() bool {
		b, err := reader.Peek(512)
		return err != nil && strings.Contains(string(b), "initializing epoch 0\n")
	}, 2*time.Second, 100*time.Millisecond, "never started process")

	shutdown()

	select { // Await run completion
	case <-time.After(10 * time.Second):
		t.Fatal("Run never completed")
	case <-ctx.Done():
	}
	return //nolint
}

var (
	// fakeEnvoySrc is a test source file used to simulate Envoy console output and signal processing.
	//go:embed testdata/fake_envoy.go
	fakeEnvoySrc []byte
	// fakeEnvoyBin is the compiled code of fakeEnvoySrc which will be runtime.GOOS dependent.
	fakeEnvoyBin   []byte
	builtFakeEnvoy sync.Once
)

// RequireFakeEnvoy writes fakeEnvoyBin to the given path
func RequireFakeEnvoy(t *testing.T, path string) {
	builtFakeEnvoy.Do(func() {
		fakeEnvoyBin = requireBuildFakeEnvoy(t)
	})
	require.NoError(t, os.WriteFile(path, fakeEnvoyBin, 0700)) //nolint:gosec
}

// requireBuildFakeEnvoy builds a fake envoy binary and returns its contents.
func requireBuildFakeEnvoy(t *testing.T) []byte {
	goBin := requireGoBin(t)
	tempDir, deleteTempDir := morerequire.RequireNewTempDir(t)
	defer deleteTempDir()

	name := "envoy"
	src := name + ".go"
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, src), fakeEnvoySrc, 0600))
	cmd := exec.Command(goBin, "build", "-o", name, src) //nolint:gosec
	cmd.Dir = tempDir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "couldn't compile %s: %s", src, string(out))
	bytes, err := os.ReadFile(filepath.Join(tempDir, name))
	require.NoError(t, err)
	return bytes
}

func requireGoBin(t *testing.T) string {
	binName := "go"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	goBin := filepath.Join(runtime.GOROOT(), "bin", binName)
	if _, err := os.Stat(goBin); err == nil {
		return goBin
	}
	// Now, search the path
	goBin, err := exec.LookPath(binName)
	require.NoError(t, err, "couldn't find %s in the PATH", goBin)
	return goBin
}
