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

	"github.com/tetratelabs/getenvoy/internal/tar"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

// Even though currently binaries are compressed with "xz" more than "gz", using "gz" in tests allows us to re-use
// tar.TarGz instead of complicating internal utilities or adding dependencies only for tests.
const archiveFormat = ".tar.gz"
const buildsPath = "/builds/"

// RequireManifestTestServer serves "/manifest.json", which contains download links a compressed archive of artifactDir
func RequireManifestTestServer(t *testing.T, m *manifest.Manifest) *httptest.Server {
	s := &server{t: t}
	h := httptest.NewServer(s)
	s.rewriteDownloadLocations(h.URL, *m)
	return h
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
	case strings.HasPrefix(r.RequestURI, buildsPath):
		reference := r.RequestURI[len(buildsPath):]
		require.True(s.t, strings.HasSuffix(reference, archiveFormat),
			"unexpected uri %q: expected archive suffix %q", reference, archiveFormat)
		ref, err := manifest.ParseReference(reference)
		require.NoError(s.t, err, "could not parse reference from uri %q", reference)
		require.Regexpf(s.t, `^1\.[1-9][0-9]\.[0-9]+$`, ref.Version, "unsupported version in uri %q", reference)

		w.WriteHeader(http.StatusOK)
		_, e := w.Write(requireFakeEnvoyTarGz(s.t, ref.Version))
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

func (s *server) rewriteDownloadLocations(baseURL string, m manifest.Manifest) {
	for _, flavor := range m.Flavors {
		for _, version := range flavor.Versions {
			for _, build := range version.Builds {
				build.DownloadLocationURL = fmt.Sprintf("%s%s%s%s", baseURL, buildsPath, build.DownloadLocationURL, archiveFormat)
			}
		}
	}
	s.manifest = &m
}

func requireFakeEnvoyTarGz(t *testing.T, version string) []byte {
	tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
	defer removeTempDir()

	// construct the platform directory based on the input version
	platformDir := filepath.Join(tempDir, version)
	require.NoError(t, os.MkdirAll(filepath.Join(platformDir, "bin"), 0700)) //nolint:gosec
	// go:embed doesn't allow us to retain execute bit, so it is simpler to inline this.
	fakeEnvoy := []byte(`#!/bin/sh
set -ue
# Echo invocation context to stdout and fake stderr to ensure it is not combined into stdout.
echo envoy wd: $PWD
echo envoy bin: $0
echo envoy args: $@
echo >&2 envoy stderr
`)
	require.NoError(t, os.WriteFile(filepath.Join(platformDir, "bin", "envoy"), fakeEnvoy, 0700)) //nolint:gosec

	// tar.gz the platform dir
	tempGz := filepath.Join(tempDir, "envoy.tar.gz")
	e := tar.TarGz(tempGz, platformDir)
	require.NoError(t, e)

	// Read the tar.gz into a byte array. This allows the mock server to set content length correctly
	f, e := os.Open(tempGz) //nolint:gosec
	require.NoError(t, e)
	defer f.Close() // nolint
	b, e := io.ReadAll(f)
	require.NoError(t, e)
	return b
}
