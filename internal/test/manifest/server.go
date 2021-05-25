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

package manifest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/manifest"
	"github.com/tetratelabs/getenvoy/internal/tar"
	"github.com/tetratelabs/getenvoy/internal/test/morerequire"
)

// Even though currently binaries are compressed with "xz" more than "gz", using "gz" in tests allows us to re-use
// tar.TarGz instead of complicating internal utilities or adding dependencies only for tests.
const archiveFormat = ".tar.gz"
const versionsPath = "/versions/"

// RequireManifestTestServer serves "/manifest.json", which contains download links a compressed archive of artifactDir
func RequireManifestTestServer(t *testing.T, version string) *httptest.Server {
	s := &server{t: t}
	h := httptest.NewServer(s)
	s.manifest = &manifest.Manifest{
		ManifestVersion: "v0.1.0",
		Flavors: map[string]*manifest.Flavor{
			"standard": {
				Name: "standard",
				Versions: map[string]*manifest.Version{
					version: {
						Name: version,
						Builds: map[string]*manifest.Build{
							"LINUX_GLIBC": {
								Platform:            "LINUX_GLIBC",
								DownloadLocationURL: TarballURL(h.URL, "linux", version),
							},
							"DARWIN": {
								Platform:            "DARWIN",
								DownloadLocationURL: TarballURL(h.URL, "darwin", version),
							},
						},
					},
				},
			},
		},
	}
	return h
}

// TarballURL gives the expected download URL for the given runtime.GOOS and Envoy version.
func TarballURL(baseURL, goos, v string) string {
	return fmt.Sprintf("%s%s%s/envoy-%s-%s-x86_64%s", baseURL, versionsPath, v, v, goos, archiveFormat)
}

// server represents an HTTP server serving GetEnvoy manifest.
type server struct {
	t        *testing.T
	manifest *manifest.Manifest
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.RequestURI == "/manifest.json":
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(s.GetManifest())
		require.NoError(s.t, err)
	case strings.HasPrefix(r.RequestURI, versionsPath):
		subpath := r.RequestURI[len(versionsPath):]
		require.True(s.t, strings.HasSuffix(subpath, archiveFormat), "unexpected uri %q: expected archive suffix %q", subpath, archiveFormat)

		v := strings.Split(subpath, "/")[0]
		require.Regexpf(s.t, `^1\.[1-9][0-9]\.[0-9]+$`, v, "unsupported version in uri %q", subpath)

		w.WriteHeader(http.StatusOK)
		_, e := w.Write(requireFakeEnvoyTarGz(s.t, v))
		require.NoError(s.t, e)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *server) GetManifest() []byte {
	data, err := json.Marshal(s.manifest)
	require.NoError(s.t, err)
	return data
}

func requireFakeEnvoyTarGz(t *testing.T, v string) []byte {
	tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
	defer removeTempDir()

	// construct the platform directory based on the input version
	installDir := filepath.Join(tempDir, v)
	require.NoError(t, os.MkdirAll(filepath.Join(installDir, "bin"), 0700)) //nolint:gosec
	// go:embed doesn't allow us to retain execute bit, so it is simpler to inline this.
	fakeEnvoy := []byte(`#!/bin/sh
set -ue
# Echo invocation context to stdout and fake stderr to ensure it is not combined into stdout.
echo envoy wd: $PWD
echo envoy bin: $0
echo envoy args: $@
echo >&2 envoy stderr
`)
	require.NoError(t, os.WriteFile(filepath.Join(installDir, "bin", "envoy"), fakeEnvoy, 0700)) //nolint:gosec

	// tar.gz the platform dir
	tempGz := filepath.Join(tempDir, "envoy.tar.gz")
	e := tar.TarGz(tempGz, installDir)
	require.NoError(t, e)

	// Read the tar.gz into a byte array. This allows the mock server to set content length correctly
	f, e := os.Open(tempGz) //nolint:gosec
	require.NoError(t, e)
	defer f.Close() // nolint
	b, e := io.ReadAll(f)
	require.NoError(t, e)
	return b
}
