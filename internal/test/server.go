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

package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/tar"
	"github.com/tetratelabs/getenvoy/internal/test/morerequire"
	"github.com/tetratelabs/getenvoy/internal/version"
)

const (
	// FakeReleaseDate helps us make sure main code doesn't accidentally write the system time instead of the expected.
	FakeReleaseDate = "2020-12-31"
	// Even though currently binaries are compressed with "xz" more than "gz", using "gz" in tests allows us to re-use
	// tar.TarGz instead of complicating internal utilities or adding dependencies only for tests.
	archiveFormat = ".tar.gz"
	versionsPath  = "/versions/"
)

// RequireEnvoyVersionsTestServer serves "/envoy-versions.json", containing download links a fake Envoy archive.
func RequireEnvoyVersionsTestServer(t *testing.T, v string) *httptest.Server {
	s := &server{t: t}
	h := httptest.NewServer(s)
	s.versions = version.EnvoyVersions{
		LatestVersion: v,
		Versions: map[string]version.EnvoyVersion{ // hard-code date so that tests don't drift
			v: {ReleaseDate: FakeReleaseDate, Tarballs: map[string]string{
				"linux/" + runtime.GOARCH:  TarballURL(h.URL, "linux", runtime.GOARCH, v),
				"darwin/" + runtime.GOARCH: TarballURL(h.URL, "darwin", runtime.GOARCH, v),
			}}},
	}
	return h
}

// TarballURL gives the expected download URL for the given runtime.GOOS and Envoy version.
func TarballURL(baseURL, goos, goarch, v string) string {
	var arch = "x86_64"
	if goarch != "arm64" {
		arch = goarch
	}
	return fmt.Sprintf("%s%s%s/envoy-%s-%s-%s%s", baseURL, versionsPath, v, v, goos, arch, archiveFormat)
}

// server represents an HTTP server serving GetEnvoy versions.
type server struct {
	t        *testing.T
	versions version.EnvoyVersions
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.RequestURI == "/envoy-versions.json":
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(s.getEnvoyVersions())
		require.NoError(s.t, err)
	case strings.HasPrefix(r.RequestURI, versionsPath):
		subpath := r.RequestURI[len(versionsPath):]
		require.True(s.t, strings.HasSuffix(subpath, archiveFormat), "unexpected uri %q: expected archive suffix %q", subpath, archiveFormat)

		v := strings.Split(subpath, "/")[0]
		require.Regexpf(s.t, `^1\.[1-9][0-9]\.[0-9]+$`, v, "unsupported version in uri %q", subpath)

		w.WriteHeader(http.StatusOK)
		_, err := w.Write(requireFakeEnvoyTarGz(s.t, v))
		require.NoError(s.t, err)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *server) getEnvoyVersions() []byte {
	data, err := json.Marshal(s.versions)
	require.NoError(s.t, err)
	return data
}

func requireFakeEnvoyTarGz(t *testing.T, v string) []byte {
	tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
	defer removeTempDir()

	// construct the platform directory based on the input version
	installDir := filepath.Join(tempDir, v)
	require.NoError(t, os.MkdirAll(filepath.Join(installDir, "bin"), 0700)) //nolint:gosec
	RequireFakeEnvoy(t, filepath.Join(installDir, "bin", "envoy"))

	// tar.gz the platform dir
	tempGz := filepath.Join(tempDir, "envoy.tar.gz")
	err := tar.TarGz(tempGz, installDir)
	require.NoError(t, err)

	// Read the tar.gz into a byte array. This allows the mock server to set content length correctly
	f, err := os.Open(tempGz) //nolint:gosec
	require.NoError(t, err)
	defer f.Close() // nolint
	b, err := io.ReadAll(f)
	require.NoError(t, err)
	return b
}
