// Copyright 2020 Tetrate
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

package os

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/tetratelabs/multierror"
)

// IsEmptyDir checks whether a given directory is empty.
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

// IsRegularFile returns an error if there is no regular file at a given path.
func IsRegularFile(name string) error {
	info, err := os.Stat(name)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New("there is a directory at a given path instead of a regular file")
	}
	return nil
}
