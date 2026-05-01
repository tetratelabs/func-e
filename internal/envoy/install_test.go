// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/admin"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/test"
	"github.com/tetratelabs/func-e/internal/test/httptest"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestUntarEnvoyError(t *testing.T) {
	tarball, tarballSHA256sum := test.RequireFakeEnvoyTarGz(t, version.LastKnownEnvoy)

	tests := []struct {
		name        string
		handler     http.HandlerFunc
		sha256Sum   version.SHA256Sum
		expectedErr string
	}{
		{
			name: "error on incorrect URL",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.NotFound(w, r)
			},
			sha256Sum:   tarballSHA256sum,
			expectedErr: `received 404 status code from $URL`,
		},
		{
			name: "error on empty",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			sha256Sum:   tarballSHA256sum,
			expectedErr: `error untarring $URL: EOF`,
		},
		{
			name: "error on not a tar",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("mary had a little lamb"))
			},
			sha256Sum:   tarballSHA256sum,
			expectedErr: `error untarring $URL: gzip: invalid header`,
		},
		{
			name: "error on wrong sha256sum a tar",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(tarball)
			},
			sha256Sum:   "cafebabe",
			expectedErr: fmt.Sprintf(`error untarring $URL: expected SHA-256 sum "cafebabe", but have %q`, tarballSHA256sum),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := filepath.Join(t.TempDir(), "dst")
			url := version.TarballURL("http://" + admin.ServerAddr + "/file.tar.gz")

			err := untarEnvoy(t.Context(), httptest.HandlerFactory(tt.handler), dst, url, tt.sha256Sum, globals.DefaultDevUserAgent)
			expectedErr := strings.ReplaceAll(tt.expectedErr, "$URL", string(url))
			require.EqualError(t, err, expectedErr)
		})
	}
}

// TestUntarEnvoy doesn't test compression formats because that logic is in tar.Tar
func TestUntarEnvoy(t *testing.T) {
	tempDir := t.TempDir()

	tarball, tarballSHA256sum := test.RequireFakeEnvoyTarGz(t, version.LastKnownEnvoy)
	written := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		written, _ = w.Write(tarball)
	})

	err := untarEnvoy(t.Context(), httptest.HandlerFactory(handler), tempDir, version.TarballURL("http://"+admin.ServerAddr), tarballSHA256sum, globals.DefaultDevUserAgent)
	require.NoError(t, err)
	require.Equal(t, len(tarball), written)
	require.FileExists(t, filepath.Join(tempDir, binEnvoy))
}

func TestInstallIfNeeded_ErrorOnIncorrectURL(t *testing.T) {
	o := setupInstallTest(t)

	o.EnvoyVersionsURL += "/varsionz.json"
	o.GetEnvoyVersions = NewGetVersions(o.HTTPClientFunc, o.EnvoyVersionsURL, o.UserAgent)
	o.EnvoyVersion = version.LastKnownEnvoy
	_, err := InstallIfNeeded(o.ctx, &o.GlobalOpts)
	require.EqualError(t, err, "received 404 status code from "+o.EnvoyVersionsURL)
	require.Empty(t, o.Out.(*bytes.Buffer))
}

func TestInstallIfNeeded_Validates(t *testing.T) {
	tests := []struct {
		name        string
		p           version.Platform
		v           version.PatchVersion
		expectedErr string
	}{
		{
			name:        "invalid version",
			p:           "darwin/amd64",
			v:           version.PatchVersion("1.1.1"),
			expectedErr: `couldn't find version "1.1.1" for platform "darwin/amd64"`,
		},
		{
			name:        "unsupported OS",
			p:           "solaris/amd64",
			v:           version.LastKnownEnvoy,
			expectedErr: fmt.Sprintf(`couldn't find version %q for platform "solaris/amd64"`, version.LastKnownEnvoy),
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			o := setupInstallTest(t)
			o.Platform = tc.p
			o.Out = new(bytes.Buffer)
			o.EnvoyVersion = tc.v
			_, e := InstallIfNeeded(o.ctx, &o.GlobalOpts)
			require.EqualError(t, e, tc.expectedErr)
			require.Empty(t, o.Out.(*bytes.Buffer))
		})
	}
}

func TestInstallIfNeeded(t *testing.T) {
	o := setupInstallTest(t)

	o.EnvoyVersion = version.LastKnownEnvoy
	envoyPath, e := InstallIfNeeded(o.ctx, &o.GlobalOpts)
	require.NoError(t, e)
	require.Equal(t, o.EnvoyPath, envoyPath)
	require.FileExists(t, envoyPath)

	versionDir := filepath.Dir(filepath.Dir(envoyPath))
	f, err := os.Stat(versionDir)
	require.NoError(t, err)
	require.Equal(t, f.ModTime().UTC().Format("2006-01-02"), string(test.FakeReleaseDate))

	require.Equal(t, fmt.Sprintf("downloading %s\n", o.tarballURL), o.Out.(*bytes.Buffer).String())
}

func TestInstallIfNeeded_AlreadyExists(t *testing.T) {
	o := setupInstallTest(t)

	require.NoError(t, os.MkdirAll(filepath.Dir(o.EnvoyPath), 0o700))
	require.NoError(t, os.WriteFile(o.EnvoyPath, []byte("fake"), 0o700))

	envoyStat, err := os.Stat(o.EnvoyPath)
	require.NoError(t, err)

	o.EnvoyVersion = version.LastKnownEnvoy
	envoyPath, e := InstallIfNeeded(o.ctx, &o.GlobalOpts)
	require.NoError(t, e)
	require.Equal(t, fmt.Sprintf("%s is already downloaded\n", version.LastKnownEnvoy), o.Out.(*bytes.Buffer).String())

	newStat, e := os.Stat(envoyPath)
	require.NoError(t, e)

	// didn't overwrite
	require.Equal(t, envoyStat, newStat)
}

func TestVerifyEnvoy(t *testing.T) {
	tempDir := t.TempDir()

	envoyPath := filepath.Join(tempDir, "versions", version.LastKnownEnvoy.String())
	require.NoError(t, os.MkdirAll(filepath.Join(envoyPath, "bin"), 0o755))
	t.Run("envoy binary doesn't exist", func(t *testing.T) {
		actualEnvoyPath, e := verifyEnvoy(envoyPath)
		require.Empty(t, actualEnvoyPath)
		require.ErrorContains(t, e, "file")
	})

	expectedEnvoyPath := filepath.Join(envoyPath, binEnvoy)
	require.NoError(t, os.WriteFile(expectedEnvoyPath, []byte{}, 0o700))
	t.Run("envoy binary ok", func(t *testing.T) {
		actualEnvoyPath, e := verifyEnvoy(envoyPath)
		require.Equal(t, expectedEnvoyPath, actualEnvoyPath)
		require.NoError(t, e)
	})

	require.NoError(t, os.Chmod(expectedEnvoyPath, 0o600))
	t.Run("envoy binary not executable", func(t *testing.T) {
		actualEnvoyPath, e := verifyEnvoy(envoyPath)
		require.Empty(t, actualEnvoyPath)
		require.EqualError(t, e, fmt.Sprintf(`envoy binary not executable at %q`, expectedEnvoyPath))
	})
}

type installTest struct {
	ctx context.Context
	globals.GlobalOpts
	tempDir    string
	tarballURL version.TarballURL
}

func setupInstallTest(t *testing.T) *installTest {
	t.Helper()
	baseURL := "http://" + admin.ServerAddr
	handler := test.NewEnvoyVersionsHandler(t, baseURL, version.LastKnownEnvoy)
	homeDir := t.TempDir()
	setup := &installTest{
		ctx:        t.Context(),
		tempDir:    t.TempDir(),
		tarballURL: test.TarballURL(baseURL, runtime.GOOS, runtime.GOARCH, version.LastKnownEnvoy),
		GlobalOpts: globals.GlobalOpts{
			ConfigHome:       homeDir,
			DataHome:         homeDir,
			StateHome:        homeDir,
			RuntimeDir:       homeDir,
			EnvoyVersionsURL: baseURL + "/envoy-versions.json",
			Out:              new(bytes.Buffer),
			Platform:         globals.DefaultPlatform,
			RunOpts: globals.RunOpts{
				EnvoyPath:      filepath.Join(homeDir, "envoy-versions", version.LastKnownEnvoy.String(), binEnvoy),
				HTTPClientFunc: httptest.HandlerFactory(handler),
			},
		},
	}
	setup.GetEnvoyVersions = NewGetVersions(setup.HTTPClientFunc, setup.EnvoyVersionsURL, setup.UserAgent)
	return setup
}
