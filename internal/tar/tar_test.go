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

package tar_test

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/tar"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

func TestNewDecompressor_Validates(t *testing.T) {
	tests := []struct{ name, src, path, expectedErr string }{
		{
			name:        "not a gz",
			src:         "empty.tar.gz",
			path:        "testdata/empty.tar.xz",
			expectedErr: "not a valid gz stream empty.tar.gz: gzip: invalid header",
		},
		{
			name:        "not an xz",
			src:         "empty.tar.xz",
			path:        "testdata/empty.tar.gz",
			expectedErr: "not a valid xz stream empty.tar.xz: xz: invalid header magic bytes",
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			f, e := os.Open(tt.path)
			require.NoError(t, e)
			defer f.Close()

			_, e = tar.NewDecompressor(tt.src, f)
			require.EqualError(t, e, tt.expectedErr)
		})
	}
}

func TestNewDecompressor(t *testing.T) {
	for _, p := range []string{
		"testdata/empty.tar.xz",
		"testdata/empty.tar.gz",
		"testdata/test.tar.xz",
		"testdata/test.tar.gz",
	} {
		p := p
		t.Run(p, func(t *testing.T) {
			f, e := os.Open(p)
			require.NoError(t, e)
			defer f.Close()

			want, e := os.ReadFile(strings.TrimSuffix(p, path.Ext(p)))
			require.NoError(t, e)

			d, e := tar.NewDecompressor(p, f)
			require.NoError(t, e)
			defer d.Close()

			have, e := io.ReadAll(d)
			require.NoError(t, e)

			require.Equal(t, want, have)
		})
	}
}

func TestUntar(t *testing.T) {
	tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
	defer removeTempDir()

	f, e := os.Open(filepath.Join("testdata", "test.tar"))
	require.NoError(t, e)
	defer f.Close()

	e = tar.Untar(tempDir, f)
	require.NoError(t, e)

	requireTestFiles(t, tempDir)
}

// requireTestFiles ensures the given directory includes the testdata/foo directory
func requireTestFiles(t *testing.T, tempDir string) {
	require.FileExists(t, filepath.Join(tempDir, "bar.sh"))
	require.FileExists(t, filepath.Join(tempDir, "bar.txt"))

	s, e := os.Stat(filepath.Join(tempDir, "bar.sh"))
	require.NoError(t, e)
	require.Equal(t, int64(755), int64(s.Mode().Perm()))

	s, e = os.Stat(filepath.Join(tempDir, "bar.txt"))
	require.NoError(t, e)
	require.Equal(t, int64(644), int64(s.Mode().Perm()))
}

func TestTar(t *testing.T) {
	tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
	defer removeTempDir()

	f, e := os.Open(requireTarTestData(t, tempDir))
	require.NoError(t, e)

	dst := filepath.Join(tempDir, "out")
	e = tar.Untar(dst, f)
	require.NoError(t, e)

	requireTestFiles(t, dst)
}

func requireTarTestData(t *testing.T, tempDir string) string {
	f, e := os.Create(filepath.Join(tempDir, "test.tar"))
	require.NoError(t, e)
	defer f.Close()

	e = tar.Tar(f, "testdata", "foo")
	require.NoError(t, e)
	return f.Name()
}
