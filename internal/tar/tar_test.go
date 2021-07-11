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
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/test/morerequire"
	"github.com/tetratelabs/func-e/internal/version"
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
			_, err := newDecompressor(bytes.NewReader(tc.junk))
			require.EqualError(t, err, tc.expectedErr)
		})
	}
}

// TestNewDecompressor shows we can handle all compressed variants of Envoy, even accidentally empty files.
//
// As of May 2021, here are example values:
// * "func-e-envoy-1.17.3.p0.g46bf743-1p74.gbb8060d-darwin-release-x86_64.tar.xz"
// * "func-e-1.11.0-bf169f9-af8a2e7-darwin-release-x86_64.tar.gz"
func TestNewDecompressor(t *testing.T) {
	for _, p := range []string{
		"testdata/empty.tar.xz",
		"testdata/empty.tar.gz",
		"testdata/test.tar.xz",
		"testdata/test.tar.gz",
	} {
		p := p
		t.Run(p, func(t *testing.T) {
			f, err := os.Open(p)
			require.NoError(t, err)
			defer f.Close()

			want, err := os.ReadFile(strings.TrimSuffix(p, path.Ext(p)))
			require.NoError(t, err)

			d, err := newDecompressor(f)
			require.NoError(t, err)
			defer d.Close()

			have, err := io.ReadAll(d)
			require.NoError(t, err)

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
			f, err := os.Open(srcFile)
			require.NoError(t, err)
			defer f.Close()

			err = Untar(dst, f)
			require.NoError(t, err)

			if tt.emptyTar {
				requireEmptyDirectory(t, dst)
			} else {
				requireTestFiles(t, dst)
			}
		})
	}
}

// TestUntarAndVerify ensures SHA-256 are valid regardless of platform running these tests.
func TestUntarAndVerify(t *testing.T) {
	for k, v := range map[string]version.SHA256Sum{
		"testdata/empty.tar.xz": version.SHA256Sum("0ff74a47ceef95ffaf6e629aac7e54d262300e5ee318830b41da1f809fc71afd"),
		"testdata/empty.tar.gz": version.SHA256Sum("0d4b6fdb100ea7581e94fbfb5d69ad42c902db6c12c4d16c298576df209c4d1e"),
		"testdata/test.tar.xz":  version.SHA256Sum("65a3a72fcd6455e464e8f2196b7594eef73f7574b57b0cc88baa96be61ca74e4"),
		"testdata/test.tar.gz":  version.SHA256Sum("e3d54b02088eb7e485c43120916644c485627c7336adee945f39be67533e1a75"),
	} {
		file := k
		sha256 := v
		t.Run(file, func(t *testing.T) {
			tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
			defer removeTempDir()

			f, err := os.Open(file)
			require.NoError(t, err)
			defer f.Close()

			err = UntarAndVerify(tempDir, f, sha256)
			require.NoError(t, err)
		})
	}
}

type errorReader struct {
	err error
}

func (r errorReader) Read(_ []byte) (n int, err error) {
	return 0, r.err
}

func TestUntarAndVerify_ErrorReading(t *testing.T) {
	tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
	defer removeTempDir()

	expectedErr := errors.New("ice cream")
	err := UntarAndVerify(tempDir, &errorReader{expectedErr}, "1234")
	require.Same(t, err, expectedErr)
}

func TestUntarAndVerify_InvalidSignature(t *testing.T) {
	tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
	defer removeTempDir()

	f, err := os.Open("testdata/empty.tar.xz")
	require.NoError(t, err)
	defer f.Close()

	err = UntarAndVerify(tempDir, f, "cafebabe")
	require.EqualError(t, err, `expected SHA-256 sum "cafebabe", but have "0ff74a47ceef95ffaf6e629aac7e54d262300e5ee318830b41da1f809fc71afd"`)
}

// requireTestFiles ensures the given directory includes the testdata/foo directory
func requireTestFiles(t *testing.T, dst string) {
	// NOTE: this will not include empty.txt as we don't want to clutter the tar with empty files
	for _, path := range []string{"bar.sh", filepath.Join("bar", "baz.txt")} {
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
