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
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/test/morerequire"
)

func TestNewDecompressor_Validates(t *testing.T) {
	tests := []struct {
		name        string
		junk        []byte
		expectedErr string
	}{
		{
			name:        "empty",
			junk:        []byte{},
			expectedErr: "EOF",
		},
		{
			name:        "short and invalid",
			junk:        []byte{1, 2, 3, 4},
			expectedErr: "EOF",
		},
		{
			name:        "longer than xz header and invalid",
			junk:        []byte("mary had a little lamb"),
			expectedErr: "gzip: invalid header",
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			_, e := newDecompressor(bytes.NewReader(tc.junk))
			require.EqualError(t, e, tc.expectedErr)
		})
	}
}

// TestNewDecompressor shows we can handle all compressed variants of Envoy, even accidentally empty files.
//
// As of May 2021, here are example values:
// * "getenvoy-envoy-1.17.3.p0.g46bf743-1p74.gbb8060d-darwin-release-x86_64.tar.xz"
// * "getenvoy-1.11.0-bf169f9-af8a2e7-darwin-release-x86_64.tar.gz"
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

			d, e := newDecompressor(f)
			require.NoError(t, e)
			defer d.Close()

			have, e := io.ReadAll(d)
			require.NoError(t, e)

			require.Equal(t, want, have)
		})
	}
}

// For simplicity, TestUntar only tests "xz" format. TestNewDecompressor already shows it handles "gz"
func TestUntar(t *testing.T) {
	for _, tt := range []struct {
		dstExists bool
		emptyTar  bool
	}{{true, true}, {true, false}, {false, true}, {false, false}} {
		tt := tt
		t.Run(fmt.Sprintf("%+v", tt), func(t *testing.T) {
			tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
			defer removeTempDir()

			dst := tempDir
			if !tt.dstExists {
				dst = filepath.Join(tempDir, "new")
			}

			srcFile := filepath.Join("testdata", "test.tar.xz")
			if tt.emptyTar {
				srcFile = filepath.Join("testdata", "empty.tar.xz")
			}
			f, e := os.Open(srcFile)
			require.NoError(t, e)
			defer f.Close()

			e = Untar(dst, f)
			require.NoError(t, e)

			if tt.emptyTar {
				requireEmptyDirectory(t, dst)
			} else {
				requireTestFiles(t, dst)
			}
		})
	}
}

// requireTestFiles ensures the given directory includes the testdata/foo directory
func requireTestFiles(t *testing.T, dst string) {
	for _, path := range []string{"bar.sh", "bar/baz.txt"} {
		want, e := os.Stat(filepath.Join("testdata", "foo", path))
		require.NoError(t, e)
		have, e := os.Stat(filepath.Join(dst, path))
		require.NoError(t, e)
		require.Equal(t, want.Mode(), have.Mode())
	}
}

// requireTestFiles ensures the given directory is empty
func requireEmptyDirectory(t *testing.T, dst string) {
	d, e := os.ReadDir(dst)
	require.NoError(t, e)
	require.Empty(t, d, e)
}

func TestTarGZ(t *testing.T) {
	tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
	defer removeTempDir()

	src := filepath.Join("testdata", "foo")
	dst := filepath.Join(tempDir, "test.tar.gz")
	e := TarGz(dst, src)
	require.NoError(t, e)

	f, e := os.Open(dst)
	require.NoError(t, e)
	defer f.Close() //nolint

	e = Untar(tempDir, f)
	require.NoError(t, e)

	requireTestFiles(t, tempDir)
}
