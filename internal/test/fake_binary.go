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

package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/func-e/internal/moreos"
)

var (
	_, currentFilePath, _, _ = runtime.Caller(0)

	// funcEGoModuleDir points to the directory where the func-e's go.mod resides.
	funcEGoModuleDir = filepath.Join(filepath.Dir(currentFilePath), "..", "..")
)

type fakeBinarySrc struct {
	content []byte
	path    string
}

// requireBuildFakeBinary builds a fake binary from source, either from its content or path.
func requireBuildFakeBinary(t *testing.T, name string, binarySrc fakeBinarySrc) []byte {
	goBin := requireGoBin(t)
	tempDir := t.TempDir()
	buildDir := funcEGoModuleDir
	bin := filepath.Join(tempDir, name+moreos.Exe)
	src := binarySrc.path
	if src == "" {
		buildDir = tempDir
		src = name + ".go"
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, src), binarySrc.content, 0600))
	}
	cmd := exec.Command(goBin, "build", "-o", bin, src) //nolint:gosec
	cmd.Dir = buildDir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "couldn't compile %s: %s", src, string(out))
	bytes, err := os.ReadFile(bin)
	require.NoError(t, err)
	return bytes
}

func requireGoBin(t *testing.T) string {
	binName := "go" + moreos.Exe
	goBin := filepath.Join(runtime.GOROOT(), "bin", binName)
	if _, err := os.Stat(goBin); err == nil {
		return goBin
	}
	// Now, search the path
	goBin, err := exec.LookPath(binName)
	require.NoError(t, err, "couldn't find %s in the PATH", goBin)
	return goBin
}
