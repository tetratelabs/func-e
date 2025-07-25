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

// RequireEnvoyVersionsTestServer serves "/envoy-versions.json", containing download links a fake Envoy archive.
func RequireEnvoyVersionsTestServer(t *testing.T, v version.PatchVersion) *httptest.Server {
	s := &server{t: t}
	h := httptest.NewServer(s)
	s.versions = version.ReleaseVersions{
		Versions: map[version.PatchVersion]version.Release{ // hard-code date so that tests don't drift
			v: {ReleaseDate: FakeReleaseDate, Tarballs: map[version.Platform]version.TarballURL{
				version.Platform("linux" + "/" + runtime.GOARCH):  TarballURL(h.URL, "linux", runtime.GOARCH, v),
				version.Platform("darwin" + "/" + runtime.GOARCH): TarballURL(h.URL, "darwin", runtime.GOARCH, v),
			}}},
		SHA256Sums: map[version.Tarball]version.SHA256Sum{},
	}
	fakeEnvoyTarGz, sha256Sum := RequireFakeEnvoyTarGz(s.t, v)
	s.fakeEnvoyTarGz = fakeEnvoyTarGz
	for _, u := range s.versions.Versions[v].Tarballs {
		s.versions.SHA256Sums[version.Tarball(path.Base(string(u)))] = sha256Sum
	}
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
	fakeEnvoyTarGz []byte
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.RequestURI == "/envoy-versions.json":
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(s.funcEVersions())
		require.NoError(s.t, err)
	case strings.HasPrefix(r.RequestURI, versionsPath):
		subpath := r.RequestURI[len(versionsPath):]
		require.True(s.t, strings.HasSuffix(subpath, archiveFormat), "unexpected uri %q: expected archive suffix %q", subpath, archiveFormat)

		v := strings.Split(subpath, "/")[0]
		require.NotNil(s.t, version.NewPatchVersion(v), "unsupported version in uri %q", subpath)

		w.WriteHeader(http.StatusOK)
		_, err := w.Write(s.fakeEnvoyTarGz)
		require.NoError(s.t, err)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *server) funcEVersions() []byte {
	data, err := json.Marshal(s.versions)
	require.NoError(s.t, err)
	return data
}

// RequireFakeEnvoyTarGz makes a fake envoy.tar.gz
func RequireFakeEnvoyTarGz(t *testing.T, v version.PatchVersion) ([]byte, version.SHA256Sum) {
	tempDir := t.TempDir()

	// construct the platform directory based on the input version
	installDir := filepath.Join(tempDir, v.String())
	require.NoError(t, os.MkdirAll(filepath.Join(installDir, "bin"), 0o700))
	fakeEnvoyBin, err := build.GoBuild(internal.FakeEnvoySrcPath, tempDir)
	require.NoError(t, err)
	// Move the binary to the installDir/bin
	err = os.Rename(fakeEnvoyBin, filepath.Join(installDir, "bin", "envoy"))
	require.NoError(t, err)

	// tar.gz the platform dir
	tempGz := filepath.Join(tempDir, "envoy.tar.gz")
	err = tar.TarGz(tempGz, installDir)
	require.NoError(t, err)

	// Read the tar.gz into a byte array. This allows the mock server to set content length correctly
	f, err := os.Open(tempGz)
	require.NoError(t, err)
	defer f.Close() //nolint
	b, err := io.ReadAll(f)
	require.NoError(t, err)
	return b, version.SHA256Sum(fmt.Sprintf("%x", sha256.Sum256(b)))
}
