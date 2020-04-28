package os

import (
	"io"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
)

func InferOutputDir(explicitDir string) (string, error) {
	if explicitDir != "" {
		dir := filepath.Clean(explicitDir)
		if filepath.IsAbs(dir) {
			return dir, nil
		}
		explicitDir = dir
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, explicitDir), nil
}

func EnsureDirExists(name string) error {
	if err := os.MkdirAll(name, os.ModeDir|0755); err != nil {
		return err
	}
	return nil
}

func IsEmptyDir(name string) (empty bool, errs error) {
	dir, err := os.Open(filepath.Clean(name))
	if err != nil {
		return false, err
	}
	defer func() {
		if e := dir.Close(); e != nil {
			errs = multierror.Append(errs, e)
		}
	}()
	files, err := dir.Readdirnames(1)
	if err != nil && err != io.EOF {
		return false, err
	}
	return len(files) == 0, nil
}
