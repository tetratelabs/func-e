// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal"
	"github.com/tetratelabs/func-e/internal/tar"
	"github.com/tetratelabs/func-e/internal/test/build"
	"github.com/tetratelabs/func-e/internal/version"
)

const (
	// FakeReleaseDate helps us make sure main code doesn't accidentally write the system time instead of the expected.
	FakeReleaseDate = version.ReleaseDate("2020-12-31")
	// Even though currently binaries are compressed with "xz" more than "gz", using "gz" in tests allows us to re-use
	// tar.TarGz instead of complicating internal utilities or adding dependencies only for tests.
	archiveFormat = ".tar.gz"
	versionsPath  = "/versions/"
)

// NewEnvoyVersionsHandler serves "/envoy-versions.json" and fake Envoy
// archives from the given base URL.
func NewEnvoyVersionsHandler(t *testing.T, baseURL string, v version.PatchVersion) http.Handler {
	t.Helper()
	s := &server{t: t}
	s.init(baseURL, v)
	return s
}

// RequireEnvoyVersionsTestServer serves "/envoy-versions.json", containing download links a fake Envoy archive.
func RequireEnvoyVersionsTestServer(t *testing.T, v version.PatchVersion) *httptest.Server {
	t.Helper()
	s := &server{t: t}
	// NOTE: Real TCP, not the pipe-backed one. Callers exec the fake binary, so can't run in synctest.
	h := httptest.NewServer(s)
	t.Cleanup(h.Close)
	s.init(h.URL, v)
	return h
}

// TarballURL gives the expected download URL for the given runtime.GOOS and Envoy version.
func TarballURL(baseURL, goos, goarch string, v version.PatchVersion) version.TarballURL {
	arch := "x86_64"
	if goarch != "arm64" {
		arch = goarch
	}
	return version.TarballURL(fmt.Sprintf("%s%s%s/envoy-%s-%s-%s%s", baseURL, versionsPath, v, v, goos, arch, archiveFormat))
}

// server represents an HTTP server serving func-e versions.
type server struct {
	t              *testing.T
	versions       version.ReleaseVersions
	versionsJSON   []byte
	fakeEnvoyTarGz []byte
}

func (s *server) init(baseURL string, v version.PatchVersion) {
	s.versions = version.ReleaseVersions{
		Versions: map[version.PatchVersion]version.Release{ // hard-code date so that tests don't drift
			v: {ReleaseDate: FakeReleaseDate, Tarballs: map[version.Platform]version.TarballURL{
				version.Platform("linux" + "/" + runtime.GOARCH):  TarballURL(baseURL, "linux", runtime.GOARCH, v),
				version.Platform("darwin" + "/" + runtime.GOARCH): TarballURL(baseURL, "darwin", runtime.GOARCH, v),
			}}},
		SHA256Sums: map[version.Tarball]version.SHA256Sum{},
	}
	fakeEnvoyTarGz, sha256Sum := RequireFakeEnvoyTarGz(s.t, v)
	s.fakeEnvoyTarGz = fakeEnvoyTarGz
	for _, u := range s.versions.Versions[v].Tarballs {
		s.versions.SHA256Sums[version.Tarball(path.Base(string(u)))] = sha256Sum
	}
	versionsJSON, err := json.Marshal(s.versions)
	require.NoError(s.t, err)
	s.versionsJSON = versionsJSON
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/envoy-versions.json":
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(s.versionsJSON)
	case strings.HasPrefix(r.URL.Path, versionsPath):
		subpath := r.URL.Path[len(versionsPath):]
		if !strings.HasSuffix(subpath, archiveFormat) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		v := strings.Split(subpath, "/")[0]
		if version.NewPatchVersion(v) == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(s.fakeEnvoyTarGz)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// RequireFakeEnvoyTarGz builds a fake Envoy archive.
func RequireFakeEnvoyTarGz(t *testing.T, v version.PatchVersion) ([]byte, version.SHA256Sum) {
	t.Helper()
	tempDir := t.ArtifactDir()

	installDir := filepath.Join(tempDir, v.String())
	require.NoError(t, os.MkdirAll(filepath.Join(installDir, "bin"), 0o700))
	fakeEnvoyBin, err := build.GoBuild(internal.FakeEnvoySrcPath, tempDir)
	require.NoError(t, err)
	require.NoError(t, os.Rename(fakeEnvoyBin, filepath.Join(installDir, "bin", "envoy")))

	tempGz := filepath.Join(tempDir, "envoy.tar.gz")
	require.NoError(t, tar.TarGz(tempGz, installDir))

	// Keep the body in memory so test servers can set Content-Length.
	f, err := os.Open(tempGz)
	require.NoError(t, err)
	defer f.Close()
	b, err := io.ReadAll(f)
	require.NoError(t, err)
	return b, version.SHA256Sum(fmt.Sprintf("%x", sha256.Sum256(b)))
}
