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

// Package tar avoids a large (~3MB) dependency on archiver/v3. These are special-cased to the needs of getenvoy.
package tar

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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
		return io.NopCloser(zr), nil
	}
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("not a valid gz stream %s: %w", src, err)
	}
	return zr, nil
}

// Untar unarchives, stripping the base directory inside the "src" archive. Ex on "/foo/bar", "dst" will have "bar/**"
func Untar(dst string, src io.Reader) error { // dst, src order like io.Copy
	tr := tar.NewReader(src)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		srcPath := filepath.Clean(header.Name)
		slash := strings.Index(srcPath, "/")
		if slash == -1 { // strip leading path
			continue
		}
		srcPath = srcPath[slash+1:]

		dstPath := filepath.Join(dst, srcPath)
		info := header.FileInfo()
		if info.IsDir() {
			if e := os.MkdirAll(dstPath, info.Mode()); e != nil {
				return e
			}
			continue
		}

		if e := extractFile(dstPath, tr, info.Mode()); e != nil {
			return e
		}
	}
	return nil
}

func extractFile(dst string, src io.Reader, perm os.FileMode) error {
	file, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm) //nolint:gosec
	if err != nil {
		return err
	}
	defer file.Close() //nolint
	_, err = io.Copy(file, src)
	return err
}

// Tar archives the source, rooted at the given directory.
// Ex Given "src" includes "envoy-2/bin" and "build/bin". If "root" is "envoy-2", the archive includes "envoy-2/bin".
func Tar(dst io.Writer, src fs.FS, root string) error { // dst, src order like io.Copy
	if root == "" {
		return errors.New("root must not be empty")
	}
	tw := tar.NewWriter(dst)
	defer tw.Close() //nolint

	basePath := filepath.Dir(root)
	// Recurse through the path including all files and directories
	return fs.WalkDir(src, root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}

		// Make a header for the file or directory based on the current file
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		// Ensure the destination file starts at the intended basePath
		header.Name = filepath.Join(basePath, path)
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if info.IsDir() {
			return nil // nothing to write
		}
		return copy(tw, src, path)
	})
}

// Copy the contents of the file into the tar without buffering
func copy(dst io.Writer, src fs.FS, path string) error { // dst, src order like io.Copy
	f, err := src.Open(path) //nolint:gosec
	if err != nil {
		return err
	}
	defer f.Close() //nolint
	_, err = io.Copy(dst, f)
	return err
}
