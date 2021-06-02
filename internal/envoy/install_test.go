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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/globals"
	"github.com/tetratelabs/getenvoy/internal/test"
	"github.com/tetratelabs/getenvoy/internal/test/morerequire"
	"github.com/tetratelabs/getenvoy/internal/version"
)

func TestUntarEnvoyError(t *testing.T) {
	userAgent := "test"
	tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
	dst := filepath.Join(tempDir, "dst")
	defer removeTempDir()

	var realHandler func(w http.ResponseWriter, r *http.Request)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if realHandler != nil {
			realHandler(w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	url := server.URL + "/file.tar.gz"
	t.Run("error on incorrect URL", func(t *testing.T) {
		err := untarEnvoy(dst, url, userAgent)
		require.EqualError(t, err, fmt.Sprintf(`received 404 status code from %s`, url))
	})

	realHandler = func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	}
	t.Run("error on empty", func(t *testing.T) {
		err := untarEnvoy(dst, url, userAgent)
		require.EqualError(t, err, fmt.Sprintf(`error untarring %s: EOF`, url))
	})

	realHandler = func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("mary had a little lamb")) //nolint
	}
	t.Run("error on not a tar", func(t *testing.T) {
		err := untarEnvoy(dst, url, userAgent)
		require.EqualError(t, err, fmt.Sprintf(`error untarring %s: gzip: invalid header`, url))
	})
}

// TestUntarEnvoy doesn't test compression formats because that logic is in tar.Tar
func TestUntarEnvoy(t *testing.T) {
	o, cleanup := setupInstallTest(t)
	defer cleanup()

	err := untarEnvoy(o.tempDir, o.tarballURL, o.UserAgent)
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(o.tempDir, binEnvoy))
}

func TestInstallIfNeeded_ErrorOnIncorrectURL(t *testing.T) {
	o, cleanup := setupInstallTest(t)
	defer cleanup()

	o.EnvoyVersionsURL += "/varsionz.json"

	_, err := InstallIfNeeded(&o.GlobalOpts, runtime.GOOS, version.LastKnownEnvoy)
	require.EqualError(t, err, "received 404 status code from "+o.EnvoyVersionsURL)
	require.Empty(t, o.Out.(*bytes.Buffer))
}

func TestInstallIfNeeded_Validates(t *testing.T) {
	o, cleanup := setupInstallTest(t)
	defer cleanup()

	tests := []struct{ name, goos, version, expectedErr string }{
		{
			name:        "invalid version",
			goos:        "darwin",
			version:     `1.1.1`,
			expectedErr: `couldn't find version "1.1.1" for platform "darwin"`,
		},
		{
			name:        "unsupported OS",
			goos:        "windows",
			version:     version.LastKnownEnvoy,
			expectedErr: fmt.Sprintf(`couldn't find version %q for platform "windows"`, version.LastKnownEnvoy),
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			o.Out = new(bytes.Buffer)
			_, e := InstallIfNeeded(&o.GlobalOpts, tc.goos, tc.version)
			require.EqualError(t, e, tc.expectedErr)
			require.Empty(t, o.Out.(*bytes.Buffer))
		})
	}
}

func TestInstallIfNeeded(t *testing.T) {
	o, cleanup := setupInstallTest(t)
	defer cleanup()
	out := o.Out.(*bytes.Buffer)

	envoyPath, e := InstallIfNeeded(&o.GlobalOpts, globals.CurrentPlatform, version.LastKnownEnvoy)
	require.NoError(t, e)
	require.Equal(t, o.EnvoyPath, envoyPath)
	require.FileExists(t, envoyPath)

	require.Equal(t, fmt.Sprintln("downloading", o.tarballURL), out.String())
}

func TestInstallIfNeeded_NotFound(t *testing.T) {
	o, cleanup := setupInstallTest(t)
	defer cleanup()

	t.Run("unknown version", func(t *testing.T) {
		_, e := InstallIfNeeded(&o.GlobalOpts, "darwin/amd64", "1.1.1")
		require.EqualError(t, e, `couldn't find version "1.1.1" for platform "darwin/amd64"`)
	})
	t.Run("unknown platform", func(t *testing.T) {
		_, e := InstallIfNeeded(&o.GlobalOpts, "windows/arm64", version.LastKnownEnvoy)
		require.EqualError(t, e, fmt.Sprintf(`couldn't find version "%s" for platform "windows/arm64"`, version.LastKnownEnvoy))
	})
}

func TestInstallIfNeeded_AlreadyExists(t *testing.T) {
	o, cleanup := setupInstallTest(t)
	defer cleanup()
	out := o.Out.(*bytes.Buffer)

	require.NoError(t, os.MkdirAll(filepath.Dir(o.EnvoyPath), 0700))
	require.NoError(t, ioutil.WriteFile(o.EnvoyPath, []byte("fake"), 0700))

	envoyStat, err := os.Stat(o.EnvoyPath)
	require.NoError(t, err)

	envoyPath, e := InstallIfNeeded(&o.GlobalOpts, runtime.GOOS, version.LastKnownEnvoy)
	require.NoError(t, e)
	require.Equal(t, fmt.Sprintln(version.LastKnownEnvoy, "is already downloaded"), out.String())

	newStat, e := os.Stat(envoyPath)
	require.NoError(t, e)

	// didn't overwrite
	require.Equal(t, envoyStat, newStat)
}

func TestVerifyEnvoy(t *testing.T) {
	tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
	defer removeTempDir()

	envoyPath := filepath.Join(tempDir, "versions", version.LastKnownEnvoy)
	require.NoError(t, os.MkdirAll(filepath.Join(envoyPath, "bin"), 0755))
	t.Run("envoy binary doesn't exist", func(t *testing.T) {
		EnvoyPath, e := verifyEnvoy(envoyPath)
		require.Empty(t, EnvoyPath)
		require.Contains(t, e.Error(), "no such file or directory")
	})

	expectedEnvoyPath := filepath.Join(envoyPath, "bin", "envoy")
	require.NoError(t, os.WriteFile(expectedEnvoyPath, []byte{}, 0600))
	t.Run("envoy binary not executable", func(t *testing.T) {
		EnvoyPath, e := verifyEnvoy(envoyPath)
		require.Empty(t, EnvoyPath)
		require.EqualError(t, e, fmt.Sprintf(`envoy binary not executable at %q`, expectedEnvoyPath))
	})

	require.NoError(t, os.Chmod(expectedEnvoyPath, 0750))
	t.Run("envoy binary ok", func(t *testing.T) {
		EnvoyPath, e := verifyEnvoy(envoyPath)
		require.Equal(t, expectedEnvoyPath, EnvoyPath)
		require.Nil(t, e)
	})
}

type installTest struct {
	globals.GlobalOpts
	tempDir, tarballURL string
}

func setupInstallTest(t *testing.T) (*installTest, func()) {
	var tearDown []func()

	tempDir, removeTempDir := morerequire.RequireNewTempDir(t)
	tearDown = append(tearDown, removeTempDir)

	versionsServer := test.RequireEnvoyVersionsTestServer(t, version.LastKnownEnvoy)
	tearDown = append(tearDown, versionsServer.Close)

	return &installTest{
			tempDir:    tempDir,
			tarballURL: test.TarballURL(versionsServer.URL, runtime.GOOS, runtime.GOARCH, version.LastKnownEnvoy),
			GlobalOpts: globals.GlobalOpts{
				HomeDir:          tempDir,
				EnvoyVersionsURL: versionsServer.URL + "/envoy-versions.json",
				UserAgent:        globals.DefaultUserAgent,
				Out:              new(bytes.Buffer),
				RunOpts: globals.RunOpts{
					EnvoyPath: filepath.Join(tempDir, "versions", version.LastKnownEnvoy, "bin", "envoy"),
				},
			},
		}, func() {
			for i := len(tearDown) - 1; i >= 0; i-- {
				tearDown[i]()
			}
		}
}
