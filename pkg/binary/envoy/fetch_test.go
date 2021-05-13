// Copyright 2019 Tetrate
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

package envoy

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ulikunitz/xz"

	"github.com/tetratelabs/getenvoy/internal/tar"
	"github.com/tetratelabs/getenvoy/pkg/globals"
	manifesttest "github.com/tetratelabs/getenvoy/pkg/test/manifest"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

const (
	// This is only for unit testing: we don't need to use latest.
	version = "1.17.2"
	// Hard code even if the unit test env isn't darwin. This avoids drifts like linux vs linux-glibc
	platform = "darwin"
)

func TestUntarEnvoyError(t *testing.T) {
	tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
	dst := filepath.Join(tempDir, "dst")
	defer removeTempDir()

	var realHandler func(w http.ResponseWriter, r *http.Request)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if realHandler != nil {
			realHandler(w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	url := server.URL + "/file.tar.xz"
	t.Run("error on incorrect URL", func(t *testing.T) {
		e := untarEnvoy(dst, url, io.Discard)
		require.EqualError(t, e, fmt.Sprintf(`received 404 status code from %s`, url))
	})

	realHandler = func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	}
	t.Run("error on empty", func(t *testing.T) {
		e := untarEnvoy(dst, url, io.Discard)
		require.EqualError(t, e, fmt.Sprintf(`not a valid xz stream %s: unexpected EOF`, url))
	})

	realHandler = func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("mary had a little lamb")) //nolint
	}
	t.Run("error on not a tar", func(t *testing.T) {
		e := untarEnvoy(dst, url, io.Discard)
		require.EqualError(t, e, fmt.Sprintf(`not a valid xz stream %s: xz: invalid header magic bytes`, url))
	})
}

func TestUntarEnvoy(t *testing.T) {
	tests := []struct {
		extension string
		path      string
	}{
		{
			extension: "tar.xz", // As of May 2021 all releases were compressed using xz
			path:      "getenvoy-envoy-1.17.1.p0.gd6a4496-1p74.gbb8060d-darwin-release-x86_64",
		},
		{
			extension: "tar.gz", // The very first release of envoy was compressed with gz
			path:      "getenvoy-1.11.0-bf169f9-af8a2e7-darwin-release-x86_64",
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why
		t.Run(test.extension, func(t *testing.T) {
			dstDir, removeDstDir := morerequire.RequireNewTempDir(t)
			defer removeDstDir()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(200)
				var zw io.WriteCloser
				if test.extension == "tar.xz" {
					zw, _ = xz.NewWriter(w)
				} else {
					zw = gzip.NewWriter(w)
				}
				defer zw.Close() //nolint

				tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
				defer removeTempDir()
				binDir := filepath.Join(tempDir, test.path, "bin")
				require.NoError(t, os.MkdirAll(binDir, 0700))
				require.NoError(t, ioutil.WriteFile(filepath.Join(binDir, "envoy"), []byte("fake"), 0700))
				require.NoError(t, tar.Tar(zw, os.DirFS(tempDir), test.path))
			}))
			defer server.Close()

			url := fmt.Sprintf(`%s/tetrate/getenvoy/%s.%s`, server.URL, test.path, test.extension)

			dst := filepath.Join(dstDir, "dst")

			out := new(bytes.Buffer)
			e := untarEnvoy(dst, url, out)
			require.NoError(t, e)
			require.Contains(t, out.String(), `100% |████████████████████████████████████████|`)
			require.FileExists(t, filepath.Join(dst, binEnvoy))
		})
	}
}

func TestFetchIfNeeded_ErrorOnIncorrectURL(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	o.ManifestURL += "/mannyfest.json"
	_, e := FetchIfNeeded(&o.GlobalOpts, o.reference)
	require.EqualError(t, e, "received 404 status code from "+o.ManifestURL)
	require.Empty(t, o.Out.(*bytes.Buffer))
}

func TestFetchIfNeeded(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()
	out := o.Out.(*bytes.Buffer)

	envoyPath, e := FetchIfNeeded(&o.GlobalOpts, o.reference)
	require.NoError(t, e)
	require.Contains(t, out.String(), o.envoyURL)
	require.Contains(t, out.String(), "100% |████████████████████████████████████████|")

	require.Equal(t, o.EnvoyPath, envoyPath)
	require.FileExists(t, envoyPath)
}

func TestFetchIfNeeded_AlreadyExists(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()
	out := o.Out.(*bytes.Buffer)

	require.NoError(t, os.MkdirAll(filepath.Dir(o.EnvoyPath), 0700))
	require.NoError(t, ioutil.WriteFile(o.EnvoyPath, []byte("fake"), 0700))

	envoyStat, err := os.Stat(o.EnvoyPath)
	require.NoError(t, err)

	envoyPath, e := FetchIfNeeded(&o.GlobalOpts, o.reference)
	require.NoError(t, e)
	require.Equal(t, fmt.Sprintf("%s/%s is already downloaded\n", version, platform), out.String())

	newStat, e := os.Stat(envoyPath)
	require.NoError(t, e)

	// didn't overwrite
	require.Equal(t, envoyStat, newStat)
}

func TestVerifyEnvoy(t *testing.T) {
	tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
	defer removeTempDir()

	platformPath := filepath.Join(tempDir, "platform")
	require.NoError(t, os.MkdirAll(filepath.Join(platformPath, "bin"), 0755))
	t.Run("envoy binary doesn't exist", func(t *testing.T) {
		EnvoyPath, e := verifyEnvoy(platformPath)
		require.Empty(t, EnvoyPath)
		require.Contains(t, e.Error(), "no such file or directory")
	})

	expectedEnvoyPath := filepath.Join(platformPath, "bin", "envoy")
	require.NoError(t, os.WriteFile(expectedEnvoyPath, []byte{}, 0600))
	t.Run("envoy binary not executable", func(t *testing.T) {
		EnvoyPath, e := verifyEnvoy(platformPath)
		require.Empty(t, EnvoyPath)
		require.EqualError(t, e, fmt.Sprintf(`envoy binary not executable at %q`, expectedEnvoyPath))
	})

	require.NoError(t, os.Chmod(expectedEnvoyPath, 0750))
	t.Run("envoy binary ok", func(t *testing.T) {
		EnvoyPath, e := verifyEnvoy(platformPath)
		require.Equal(t, expectedEnvoyPath, EnvoyPath)
		require.Nil(t, e)
	})
}

type manifestTest struct {
	globals.GlobalOpts
	reference, envoyURL string
}

func setupTest(t *testing.T) (*manifestTest, func()) {
	var tearDown []func()

	homeDir, removeHomeDir := morerequire.RequireNewTempDir(t)
	tearDown = append(tearDown, removeHomeDir)

	ref := fmt.Sprintf("standard:%s/%s", version, platform)
	m, err := manifesttest.NewSimpleManifest(ref)
	require.NoError(t, err)
	manifestServer := manifesttest.RequireManifestTestServer(t, m)
	tearDown = append(tearDown, manifestServer.Close)

	return &manifestTest{
			reference: ref,
			envoyURL:  fmt.Sprintf("%s/builds/%s/%s.tar.xz", manifestServer.URL, version, platform),
			GlobalOpts: globals.GlobalOpts{
				HomeDir:     homeDir,
				ManifestURL: manifestServer.URL + "/manifest.json",
				Out:         new(bytes.Buffer),
				RunOpts: globals.RunOpts{
					EnvoyPath: filepath.Join(homeDir, "builds", "standard", version, platform, "bin", "envoy"),
				},
			},
		}, func() {
			for i := len(tearDown) - 1; i >= 0; i-- {
				tearDown[i]()
			}
		}
}
