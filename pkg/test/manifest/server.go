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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/tetratelabs/getenvoy/api"
	"github.com/tetratelabs/getenvoy/pkg/types"
)

const archiveFormat = ".tar.gz"

// RequireManifestTestServer serves "/manifest.json", which contains download links a compressed archive of artifactDir
func RequireManifestTestServer(t *testing.T, manifest *api.Manifest) *httptest.Server {
	s := &server{t: t}
	h := httptest.NewServer(s)
	s.manifest = rewriteDownloadLocations(t, h.URL, manifest)
	return h
}

// server represents an HTTP server serving GetEnvoy manifest.
type server struct {
	t        *testing.T
	manifest *api.Manifest
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
	ref, err := types.ParseReference(uri)
	require.NoError(s.t, err, "could not parse reference from uri %q", uri)
	if ref.Flavor == "wasm" {
		return // don't validate
	}
	require.Equal(s.t, "standard", ref.Flavor, "unsupported flavor in uri %q", uri)
	ver, err := semver.NewVersion(ref.Version)
	require.NoError(s.t, err, "could not parse version from uri %q", uri)
	require.GreaterOrEqualf(s.t, uint64(1), ver.Major(), "unsupported major version in uri %q", uri)
	require.GreaterOrEqualf(s.t, uint64(17), ver.Minor(), "unsupported minor version in uri %q", uri)
}

func (s *server) GetManifest() []byte {
	data, err := protojson.Marshal(s.manifest)
	require.NoError(s.t, err)
	return data
}

// writeFakeEnvoyTarGz writes a tar.gz containing a "bin/envoy" that echos the commandline, output and stderr.
// TODO: fake via exec.Run in unit tests because it is less complicated and error-prone than faking via shell scripts.
func writeFakeEnvoyTarGz(t *testing.T, buf io.Writer) {
	// Create script literal of $envoyHome/bin/envoy which copies the current directory to $envoyCapture when invoked.
	// stdout and stderr are prefixed "envoy " to differentiate them from other command output, namely docker.
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

func rewriteDownloadLocations(t *testing.T, baseURL string, m *api.Manifest) *api.Manifest {
	manifest, ok := proto.Clone(m).(*api.Manifest) // safe copy
	require.True(t, ok)

	for _, flavor := range manifest.Flavors {
		for _, version := range flavor.Versions {
			for _, build := range version.Builds {
				build.DownloadLocationUrl = fmt.Sprintf("%s/builds/%s%s", baseURL, build.DownloadLocationUrl, archiveFormat)
			}
		}
	}
	return manifest
}
