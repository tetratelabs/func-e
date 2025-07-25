// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	out := filepath.Join(outDir, baseName)
	fmt.Fprintf(os.Stderr, "Building %s...\n", out) //nolint:errcheck
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
