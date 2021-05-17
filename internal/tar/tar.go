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

package tar

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver/v3"
	"github.com/ulikunitz/xz"
)

// NewDecompressor returns a compression function based on the "src" filename.
// NOTE: As of May 2021, all Envoy binaries except the first are tar.xz
func NewDecompressor(src string, r io.Reader) (io.ReadCloser, error) {
	if strings.HasSuffix(src, "tar.xz") {
		zr, err := xz.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("not a valid xz stream %s: %w", src, err)
		}
		return ioutil.NopCloser(zr), nil
	}
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("not a valid gz stream %s: %w", src, err)
	}
	return zr, nil
}

// Untar unarchives, stripping the base directory inside the "src" archive. Ex on "/foo/bar", "dst" will have "bar/**"
func Untar(dst string, src io.Reader) error { // dst, src order like io.Copy
	// No support for streaming https://github.com/mholt/archiver/pull/199
	tmpDir, err := ioutil.TempDir("", "getenvoy-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir) //nolint

	archive := filepath.Join(tmpDir, "archive.tar")
	tarball, err := os.OpenFile(archive, os.O_CREATE|os.O_WRONLY, 0600) //nolint:gosec
	if err != nil {
		return err
	}
	defer tarball.Close() //nolint
	if _, err = io.Copy(tarball, src); err != nil {
		return err
	}

	tar := &archiver.Tar{MkdirAll: true, StripComponents: 1}
	defer tar.Close() //nolint
	return tar.Unarchive(archive, dst)
}

// Tar archives the source, rooted at the given directory.
// Ex Given "src" includes "envoy-2/bin" and "build/bin". If "root" is "envoy-2", the archive includes "envoy-2/bin".
func Tar(dst io.Writer, src, root string) error { // dst, src order like io.Copy
	if root == "" {
		return errors.New("root must not be empty")
	}

	// No support for streaming https://github.com/mholt/archiver/pull/199
	archive := filepath.Join(src, "archive.tar")
	err := archiver.DefaultTar.Archive([]string{filepath.Join(src, root)}, archive)
	if err != nil {
		return err
	}
	defer os.Remove(archive) //nolint

	archiveFile, err := os.Open(archive) //nolint:gosec
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, archiveFile)
	return err
}
