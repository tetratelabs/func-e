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

// Package tar avoids third-party dependencies (ex archiver/v3) and are special-cased to the needs of getenvoy.
package tar

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
)

// Untar unarchives the compressed "src" which is either a "tar.xz" or "tar.gz" stream.
// This strips the base directory inside the "src" archive. Ex on "/foo/bar", "dst" will have "bar/**"
//
// This is used to decompress Envoy distributions in the "downloadLocationURL" field of "manifest.json".
// To keep the binary size small, only supports compression formats used in practice. As of May 2021, all
// "downloadLocationURL" values were "tar.xz", except the first (1.11.0), which was "tar.gz".
func Untar(dst string, src io.Reader) error { // dst, src order like io.Copy
	if e := os.MkdirAll(dst, 0750); e != nil {
		return e
	}

	zSrc, e := newDecompressor(src)
	if e != nil {
		return e
	}
	defer zSrc.Close() //nolint

	tr := tar.NewReader(zSrc)
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

// newDecompressor returns an "xz" or "gzip" decompression function based on bytes in the stream.
func newDecompressor(r io.Reader) (io.ReadCloser, error) {
	br := bufio.NewReader(r)
	h, err := br.Peek(xz.HeaderLen)
	if err != nil { // This is only used to decompress Envoy, so a valid stream will never be this short.
		return nil, err
	}
	if xz.ValidHeader(h) {
		xzr, e := xz.NewReader(br)
		if e != nil {
			return nil, err
		}
		return io.NopCloser(xzr), nil
	}
	return gzip.NewReader(br)
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

// TarGz tars and gzips "src", rooted at the last directory, into the file named "dst"
// Ex If "src" includes "/tmp/envoy/bin" and "/tmp/build/bin". If "src" is "/tmp/envoy", "dst" includes "envoy/bin".
//
// This is used to compress the working directory of Envoy after it is stopped.
func TarGz(dst, src string) error { //nolint dst, src order like io.Copy
	srcFS := os.DirFS(filepath.Dir(src))
	basePath := filepath.Base(src)

	// create the tar.gz file
	file, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600) //nolint:gosec
	if err != nil {
		return err
	}
	defer file.Close() //nolint
	gzw := gzip.NewWriter(file)
	defer gzw.Close() //nolint
	tw := tar.NewWriter(gzw)
	defer tw.Close() //nolint

	// Recurse through the path including all files and directories
	return fs.WalkDir(srcFS, basePath, func(path string, d os.DirEntry, err error) error {
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

		// Ensure the destination file starts at the intended path
		header.Name = path
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if info.IsDir() {
			return nil // nothing to write
		}
		return copy(tw, srcFS, path)
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
