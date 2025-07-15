// Copyright 2025 Tetrate
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

package build

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/tetratelabs/func-e/internal/moreos"
)

// GoBuild builds a go binary from the given source file and outputs it to the specified path.
// Note: Be careful with outDir so that it is at the right scope! For example, you likely don't want to use
// t.TempDir() unless it is really scoped to a single test function.
func GoBuild(src, outDir string) (string, error) {
	goBin, err := findGoBin()
	if err != nil {
		return "", err
	}

	// We change the working directory to be the go module root directory, so the source and
	// dest need to be absolute paths.
	for _, path := range []string{src, outDir} {
		if !filepath.IsAbs(path) {
			return "", fmt.Errorf("must be an absolute path: %s", path)
		} else if _, err = os.Stat(path); errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("must exist: %s", path)
		}
	}

	// Use the same naming convention for the out file as its source directory
	baseName := filepath.Base(filepath.Dir(src))
	out := filepath.Join(outDir, baseName+moreos.Exe)
	moreos.Fprintf(os.Stderr, "Building %s...\n", out)
	// Build from the project root directory
	buildCmd := exec.Command(goBin, "build",
		"-ldflags", "-s -w -X main.version=dev",
		"-o", out, src)
	if buildCmd.Dir, err = findGoModRoot(); err != nil {
		return "", err
	}
	if outBytes, buildErr := buildCmd.CombinedOutput(); buildErr != nil {
		return "", fmt.Errorf("couldn't compile %s: %w\n%s", src, buildErr, string(outBytes))
	}
	if err = os.Chmod(out, 0o755); err != nil {
		return "", fmt.Errorf("could not make %s executable: %w", out, err)
	}
	return out, nil
}

func findGoBin() (string, error) {
	binName := "go"
	goBin := filepath.Join(os.Getenv("GOROOT"), "bin", binName)
	if _, err := os.Stat(goBin); err == nil {
		return goBin, nil
	} else if goBin, err = exec.LookPath(binName); err != nil {
		return "", fmt.Errorf("could not find %s in GOROOT or PATH: %w", binName, err)
	} else {
		return goBin, nil
	}
}

// findGoModRoot walks up the directory tree to find the directory containing go.mod
func findGoModRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("could not determine current file path")
	}
	dir := filepath.Dir(file)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no go.mod found in any parent directory")
		}
		dir = parent
	}
}
