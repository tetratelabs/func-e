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
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/manifest"
)

const archiveFormat = ".tar.gz"

// RequireManifestTestServer serves "/manifest.json", which contains download links a compressed archive of artifactDir
func RequireManifestTestServer(t *testing.T, m *manifest.Manifest) *httptest.Server {
	s := &server{t: t}
	h := httptest.NewServer(s)
	s.manifest = rewriteDownloadLocations(h.URL, *m)
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
	case strings.HasPrefix(r.RequestURI, "/builds/"):
		uri := r.RequestURI[len("/builds/"):]
		require.True(s.t, strings.HasSuffix(uri, archiveFormat),
			"unexpected uri %q: expected archive suffix %q", uri, archiveFormat)
		s.validateReference(uri)
		w.WriteHeader(http.StatusOK)
		writeFakeEnvoyTarGz(s.t, w)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *server) validateReference(uri string) {
	ref, err := manifest.ParseReference(uri)
	require.NoError(s.t, err, "could not parse reference from uri %q", uri)
	require.Regexpf(s.t, `^1\.[1-9][0-9]\.[0-9]+$`, ref.Version, "unsupported version in uri %q", uri)
}

func (s *server) GetManifest() []byte {
	data, err := json.Marshal(s.manifest)
	require.NoError(s.t, err)
	return data
}

// writeFakeEnvoyTarGz writes a tar.gz containing a "bin/envoy" that echos the commandline, output and stderr.
// TODO: fake via exec.Run in unit tests because it is less complicated and error-prone than faking via shell scripts.
func writeFakeEnvoyTarGz(t *testing.T, buf io.Writer) {
	// Create script literal of $envoyHome/bin/envoy which copies the current directory to $envoyCapture when invoked.
	// stdout and stderr are prefixed "envoy " to differentiate them from other command output.
	fakeEnvoy := []byte(`#!/bin/sh
set -ue
# Echo invocation context to stdout and fake stderr to ensure it is not combined into stdout.
echo envoy wd: $PWD
echo envoy bin: $0
echo envoy args: $@
echo >&2 envoy stderr
`)
	gw := gzip.NewWriter(buf)
	defer gw.Close() // nolint
	tw := tar.NewWriter(gw)
	defer tw.Close() // nolint

	err := tw.WriteHeader(&tar.Header{Name: "bin", Mode: 0750, Typeflag: tar.TypeDir})
	require.NoError(t, err)
	err = tw.WriteHeader(&tar.Header{Name: "bin/envoy", Mode: 0750, Size: int64(len(fakeEnvoy)), Typeflag: tar.TypeReg})
	require.NoError(t, err)
	_, err = tw.Write(fakeEnvoy)
	require.NoError(t, err)
}

func rewriteDownloadLocations(baseURL string, m manifest.Manifest) *manifest.Manifest {
	for _, flavor := range m.Flavors {
		for _, version := range flavor.Versions {
			for _, build := range version.Builds {
				build.DownloadLocationURL = fmt.Sprintf("%s/builds/%s%s", baseURL, build.DownloadLocationURL, archiveFormat)
			}
		}
	}
	return &m
}
