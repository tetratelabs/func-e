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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/globals"
	manifesttest "github.com/tetratelabs/getenvoy/internal/test/manifest"
	"github.com/tetratelabs/getenvoy/internal/test/morerequire"
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

	url := server.URL + "/file.tar.gz"
	t.Run("error on incorrect URL", func(t *testing.T) {
		e := untarEnvoy(dst, url, io.Discard)
		require.EqualError(t, e, fmt.Sprintf(`received 404 status code from %s`, url))
	})

	realHandler = func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	}
	t.Run("error on empty", func(t *testing.T) {
		e := untarEnvoy(dst, url, io.Discard)
		require.EqualError(t, e, fmt.Sprintf(`error untarring %s: EOF`, url))
	})

	realHandler = func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("mary had a little lamb")) //nolint
	}
	t.Run("error on not a tar", func(t *testing.T) {
		e := untarEnvoy(dst, url, io.Discard)
		require.EqualError(t, e, fmt.Sprintf(`error untarring %s: gzip: invalid header`, url))
	})
}

// TestUntarEnvoy doesn't test compression formats because that logic is in tar.Tar
func TestUntarEnvoy(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	out := new(bytes.Buffer)
	e := untarEnvoy(o.tempDir, o.envoyURL, out)
	require.NoError(t, e)
	require.FileExists(t, filepath.Join(o.tempDir, binEnvoy))
	require.Contains(t, out.String(), `100% |████████████████████████████████████████|`)
}

func TestFetchIfNeeded_ErrorOnIncorrectURL(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	o.ManifestURL += "/mannyfest.json"
	_, e := FetchIfNeeded(&o.GlobalOpts, o.reference)
	require.EqualError(t, e, "received 404 status code from "+o.ManifestURL)
	require.Empty(t, o.Out.(*bytes.Buffer))
}

// progressReader emits a progress bar, when the length is known
func TestProgressReader(t *testing.T) {
	var out bytes.Buffer
	b := []byte{1, 2, 3, 4}
	br := progressReader(&out, bytes.NewReader(b), int64(len(b)))
	_, e := io.ReadAll(br)

	require.NoError(t, e)
	require.Contains(t, out.String(), "100% |████████████████████████████████████████|")
}

// progressReader emits a spinner with download rate when it doesn't know the length
func TestProgressReader_UnknownLength(t *testing.T) {
	var out bytes.Buffer
	b := []byte{1, 2, 3, 4}
	br := progressReader(&out, bytes.NewReader(b), -1)
	_, e := io.ReadAll(br)

	require.NoError(t, e)
	require.NotContains(t, out.String(), "100% |████████████████████████████████████████|")
}

func TestFetchIfNeeded(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()
	out := o.Out.(*bytes.Buffer)

	envoyPath, e := FetchIfNeeded(&o.GlobalOpts, o.reference)
	require.NoError(t, e)
	require.Equal(t, o.EnvoyPath, envoyPath)
	require.FileExists(t, envoyPath)

	require.Contains(t, out.String(), o.envoyURL)
	require.Contains(t, out.String(), "100% |████████████████████████████████████████|")
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
	tempDir, reference, envoyURL string
}

func setupTest(t *testing.T) (*manifestTest, func()) {
	var tearDown []func()

	tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
	tearDown = append(tearDown, removeTempDir)

	ref := fmt.Sprintf("standard:%s/%s", version, platform)
	m, err := manifesttest.NewSimpleManifest(ref)
	require.NoError(t, err)
	manifestServer := manifesttest.RequireManifestTestServer(t, m)
	tearDown = append(tearDown, manifestServer.Close)

	return &manifestTest{
			tempDir:   tempDir,
			reference: ref,
			envoyURL:  fmt.Sprintf("%s/builds/%s/%s.tar.gz", manifestServer.URL, version, platform),
			GlobalOpts: globals.GlobalOpts{
				HomeDir:     tempDir,
				ManifestURL: manifestServer.URL + "/manifest.json",
				Out:         new(bytes.Buffer),
				RunOpts: globals.RunOpts{
					EnvoyPath: filepath.Join(tempDir, "builds", "standard", version, platform, "bin", "envoy"),
				},
			},
		}, func() {
			for i := len(tearDown) - 1; i >= 0; i-- {
				tearDown[i]()
			}
		}
}
