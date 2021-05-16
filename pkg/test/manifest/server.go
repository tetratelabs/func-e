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
	"github.com/ulikunitz/xz"

	tar "github.com/tetratelabs/getenvoy/pkg/internal"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

const archiveFormat = ".tar.xz"
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
		s.writeFakeEnvoyTarXz(w, ref.Version)
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

func (s *server) writeFakeEnvoyTarXz(w io.Writer, version string) {
	tempDir, removeTempDir := morerequire.RequireNewTempDir(s.t)
	defer removeTempDir()
	require.NoError(s.t, os.MkdirAll(filepath.Join(tempDir, version, "bin"), 0700)) //nolint:gosec
	// go:embed doesn't allow us to retain execute bit, so it is simpler to inline this.
	fakeEnvoy := []byte(`#!/bin/sh
set -ue
# Echo invocation context to stdout and fake stderr to ensure it is not combined into stdout.
echo envoy wd: $PWD
echo envoy bin: $0
echo envoy args: $@
echo >&2 envoy stderr
`)
	require.NoError(s.t, os.WriteFile(filepath.Join(tempDir, version, "bin", "envoy"), fakeEnvoy, 0700)) //nolint:gosec

	zw, err := xz.NewWriter(w)
	require.NoError(s.t, err)
	defer zw.Close() //nolint
	err = tar.Tar(zw, tempDir, version)
	require.NoError(s.t, err)
}
