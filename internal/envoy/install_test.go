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
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/test"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestUntarEnvoyError(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	dst := filepath.Join(tempDir, "dst")

	tarball, tarballSHA256sum := test.RequireFakeEnvoyTarGz(t, version.LastKnownEnvoy)

	var realHandler func(w http.ResponseWriter, r *http.Request)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if realHandler != nil {
			realHandler(w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	url := version.TarballURL(server.URL + "/file.tar.gz")
	t.Run("error on incorrect URL", func(t *testing.T) {
		err := untarEnvoy(ctx, dst, url, tarballSHA256sum, globals.DefaultPlatform, "dev")
		require.EqualError(t, err, fmt.Sprintf(`received 404 status code from %s`, url))
	})

	realHandler = func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	}
	t.Run("error on empty", func(t *testing.T) {
		err := untarEnvoy(ctx, dst, url, tarballSHA256sum, globals.DefaultPlatform, "dev")
		require.EqualError(t, err, fmt.Sprintf(`error untarring %s: EOF`, url))
	})

	realHandler = func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("mary had a little lamb"))
	}
	t.Run("error on not a tar", func(t *testing.T) {
		err := untarEnvoy(ctx, dst, url, tarballSHA256sum, globals.DefaultPlatform, "dev")
		require.EqualError(t, err, fmt.Sprintf(`error untarring %s: gzip: invalid header`, url))
	})

	realHandler = func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write(tarball)
	}
	t.Run("error on wrong sha256sum a tar", func(t *testing.T) {
		err := untarEnvoy(ctx, dst, url, "cafebabe", globals.DefaultPlatform, "dev")
		require.EqualError(t, err, fmt.Sprintf(`error untarring %s: expected SHA-256 sum "cafebabe", but have "%s"`, url, tarballSHA256sum))
	})
}

// TestUntarEnvoy doesn't test compression formats because that logic is in tar.Tar
func TestUntarEnvoy(t *testing.T) {
	tempDir := t.TempDir()

	tarball, tarballSHA256sum := test.RequireFakeEnvoyTarGz(t, version.LastKnownEnvoy)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		l, err := w.Write(tarball)
		require.NoError(t, err)
		require.Equal(t, len(tarball), l)
	}))
	defer server.Close()

	err := untarEnvoy(context.Background(), tempDir, version.TarballURL(server.URL), tarballSHA256sum, globals.DefaultPlatform, "dev")
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(tempDir, binEnvoy))
}

func TestInstallIfNeeded_ErrorOnIncorrectURL(t *testing.T) {
	o := setupInstallTest(t)

	o.EnvoyVersionsURL += "/varsionz.json"
	o.GetEnvoyVersions = NewGetVersions(o.EnvoyVersionsURL, o.Platform, o.Version)
	o.EnvoyVersion = version.LastKnownEnvoy
	_, err := InstallIfNeeded(o.ctx, &o.GlobalOpts)
	require.EqualError(t, err, "received 404 status code from "+o.EnvoyVersionsURL)
	require.Empty(t, o.Out.(*bytes.Buffer))
}

func TestInstallIfNeeded_Validates(t *testing.T) {
	o := setupInstallTest(t)

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
		o.Platform = tt.p
		t.Run(tc.name, func(t *testing.T) {
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
	out := o.Out.(*bytes.Buffer)

	o.EnvoyVersion = version.LastKnownEnvoy
	envoyPath, e := InstallIfNeeded(o.ctx, &o.GlobalOpts)
	require.NoError(t, e)
	require.Equal(t, o.EnvoyPath, envoyPath)
	require.FileExists(t, envoyPath)

	// The version directory timestamp matches the fake release date, not the current time
	versionDir := strings.Replace(envoyPath, binEnvoy, "", 1)
	f, err := os.Stat(versionDir)
	require.NoError(t, err)
	require.Equal(t, f.ModTime().UTC().Format("2006-01-02"), string(test.FakeReleaseDate))

	require.Equal(t, moreos.Sprintf("downloading %s\n", o.tarballURL), out.String())
}

func TestInstallIfNeeded_NotFound(t *testing.T) {
	o := setupInstallTest(t)

	t.Run("unknown version", func(t *testing.T) {
		o.Platform = "darwin/amd64"
		o.EnvoyVersion = version.PatchVersion("1.1.1")
		_, e := InstallIfNeeded(o.ctx, &o.GlobalOpts)
		require.EqualError(t, e, `couldn't find version "1.1.1" for platform "darwin/amd64"`)
	})
	t.Run("unknown platform", func(t *testing.T) {
		o.Platform = "solaris/amd64"
		o.EnvoyVersion = version.LastKnownEnvoy
		_, e := InstallIfNeeded(o.ctx, &o.GlobalOpts)
		require.EqualError(t, e, fmt.Sprintf(`couldn't find version "%s" for platform "solaris/amd64"`, version.LastKnownEnvoy))
	})
}

func TestInstallIfNeeded_AlreadyExists(t *testing.T) {
	o := setupInstallTest(t)
	out := o.Out.(*bytes.Buffer)

	require.NoError(t, os.MkdirAll(filepath.Dir(o.EnvoyPath), 0o700))
	require.NoError(t, os.WriteFile(o.EnvoyPath, []byte("fake"), 0o700))

	envoyStat, err := os.Stat(o.EnvoyPath)
	require.NoError(t, err)

	o.EnvoyVersion = version.LastKnownEnvoy
	envoyPath, e := InstallIfNeeded(o.ctx, &o.GlobalOpts)
	require.NoError(t, e)
	require.Equal(t, moreos.Sprintf("%s is already downloaded\n", version.LastKnownEnvoy), out.String())

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
		EnvoyPath, e := verifyEnvoy(envoyPath)
		require.Empty(t, EnvoyPath)
		require.Contains(t, e.Error(), "file")
	})

	expectedEnvoyPath := filepath.Join(envoyPath, binEnvoy)
	require.NoError(t, os.WriteFile(expectedEnvoyPath, []byte{}, 0o700))
	t.Run("envoy binary ok", func(t *testing.T) {
		EnvoyPath, e := verifyEnvoy(envoyPath)
		require.Equal(t, expectedEnvoyPath, EnvoyPath)
		require.Nil(t, e)
	})

	require.NoError(t, os.Chmod(expectedEnvoyPath, 0o600))
	t.Run("envoy binary not executable", func(t *testing.T) {
		if runtime.GOOS == moreos.OSWindows {
			t.Skip("execute bit isn't visible on windows")
		}
		EnvoyPath, e := verifyEnvoy(envoyPath)
		require.Empty(t, EnvoyPath)
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
	versionsServer := test.RequireEnvoyVersionsTestServer(t, version.LastKnownEnvoy)
	homeDir := t.TempDir()
	setup := &installTest{
		ctx:        context.Background(),
		tempDir:    t.TempDir(),
		tarballURL: test.TarballURL(versionsServer.URL, runtime.GOOS, runtime.GOARCH, version.LastKnownEnvoy),
		GlobalOpts: globals.GlobalOpts{
			HomeDir:          homeDir,
			EnvoyVersionsURL: versionsServer.URL + "/envoy-versions.json",
			Out:              new(bytes.Buffer),
			Platform:         globals.DefaultPlatform,
			RunOpts: globals.RunOpts{
				EnvoyPath: filepath.Join(homeDir, "versions", version.LastKnownEnvoy.String(), binEnvoy),
			},
		},
	}
	setup.GetEnvoyVersions = NewGetVersions(setup.EnvoyVersionsURL, setup.Platform, setup.Version)
	t.Cleanup(versionsServer.Close)
	return setup
}
