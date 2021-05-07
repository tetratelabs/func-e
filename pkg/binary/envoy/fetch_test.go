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
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mholt/archiver/v3"
	"github.com/stretchr/testify/require"

	reference "github.com/tetratelabs/getenvoy/pkg"
	"github.com/tetratelabs/getenvoy/pkg/globals"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
	manifesttest "github.com/tetratelabs/getenvoy/pkg/test/manifest"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

func TestFetchIfNeeded(t *testing.T) {
	o := &globals.GlobalOpts{}
	homeDir, removeHomeDir := morerequire.RequireNewTempDir(t)
	defer removeHomeDir()
	o.HomeDir = homeDir

	// Hard code even if the unit test env isn't darwin. This avoids drifts like linux vs linux-glibc
	r := "standard:1/darwin"
	testManifest, err := manifesttest.NewSimpleManifest(r)
	require.NoError(t, err, `error creating test manifest`)

	manifestServer := manifesttest.RequireManifestTestServer(t, testManifest)
	defer manifestServer.Close()

	t.Run("error on incorrect URL", func(t *testing.T) {
		o.ManifestURL = manifestServer.URL + "/mannyfest.json"
		_, e := FetchIfNeeded(o, r)
		require.EqualError(t, e, "received 404 response code from "+o.ManifestURL)
	})

	o.ManifestURL = manifestServer.URL + "/manifest.json"

	expectedPath := filepath.Join(homeDir, "builds", "standard", "1", "darwin", "bin", "envoy")
	t.Run("downloads when doesn't exists", func(t *testing.T) {
		envoyPath, e := FetchIfNeeded(o, r)
		require.NoError(t, e)
		require.Equal(t, expectedPath, envoyPath)
		require.FileExists(t, envoyPath)
	})

	envoyStat, err := os.Stat(expectedPath)
	require.NoError(t, err)

	t.Run("doesn't error when already exists", func(t *testing.T) {
		envoyPath, e := FetchIfNeeded(o, r)
		require.NoError(t, e)

		newStat, e := os.Stat(envoyPath)
		require.NoError(t, e)

		// didn't overwrite
		require.Equal(t, envoyStat, newStat)
	})
}

func TestFetchIfNeededAlreadyExists(t *testing.T) {
	// Hard code even if the unit test env isn't darwin. This avoids drifts like linux vs linux-glibc
	r := "standard:1/darwin"
	testManifest, err := manifesttest.NewSimpleManifest(r)
	require.NoError(t, err, `error creating test manifest`)

	o := &globals.GlobalOpts{}
	homeDir, removeHomeDir := morerequire.RequireNewTempDir(t)
	defer removeHomeDir()
	o.HomeDir = homeDir

	manifestServer := manifesttest.RequireManifestTestServer(t, testManifest)
	defer manifestServer.Close()
	o.ManifestURL = manifestServer.URL + "/manifest.json"

	envoyPath, err := FetchIfNeeded(o, r)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(homeDir, "builds", "standard", "1", "darwin", "bin", "envoy"), envoyPath)
	require.FileExists(t, envoyPath)
}

// TODO: this whole test needs to be rewritten, possibly with the mock registry server
func TestFetchEnvoy(t *testing.T) {
	key, err := manifest.NewKey(reference.Latest)
	require.NoError(t, err)

	defaultDarwinKey := &manifest.Key{Flavor: key.Flavor, Version: key.Version, Platform: "darwin"}
	tests := []struct {
		name             string
		key              *manifest.Key
		tarballStructure string
		tarExtension     string
		responseStatus   int
		wantErr          bool
		wantServerCalled bool
	}{
		{
			name:             "Downloads and untars envoy and runtime libs to store/key/bin and store/key/lib",
			key:              defaultDarwinKey,
			tarballStructure: "envoy",
			tarExtension:     ".tar.gz",
			responseStatus:   http.StatusOK,
			wantServerCalled: true,
		},
		{
			name:             "Downloads and untars envoy and runtime libs to store/key/bin and store/key/lib",
			key:              defaultDarwinKey,
			tarballStructure: "envoy",
			tarExtension:     ".tar.xz",
			responseStatus:   http.StatusOK,
			wantServerCalled: true,
		},
		{
			name:             "errors if it can't find an envoy binary in tarball",
			key:              defaultDarwinKey,
			tarballStructure: "envoy/lib",
			tarExtension:     ".tar.gz",
			responseStatus:   http.StatusOK,
			wantErr:          true,
			wantServerCalled: true,
		},
		{
			name:             "errors if it gets !200 from download",
			key:              defaultDarwinKey,
			tarballStructure: "envoy/lib",
			tarExtension:     ".tar.gz",
			responseStatus:   http.StatusTeapot,
			wantErr:          true,
			wantServerCalled: true,
		},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			tempDir, cleanupTempDir := morerequire.RequireNewTempDir(t)
			defer cleanupTempDir()

			mock, gotCalled := mockServer(t, tc.responseStatus, tc.tarballStructure, tc.tarExtension, tempDir)

			err = fetchEnvoy(tempDir, mock.URL+"/"+tc.tarballStructure+tc.tarExtension)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
				for _, location := range []string{filepath.Join(tempDir, "lib/somelib"), filepath.Join(tempDir, "bin/envoy")} {
					f, _ := os.Open(location)
					bytes, _ := io.ReadAll(f)
					require.Contains(t, string(bytes), "some c++")
				}
			}
			require.Equal(t, tc.wantServerCalled, *gotCalled, "mismatch of expectations for calling of remote server")
		})
	}
}

func mockServer(t *testing.T, responseStatusCode int, tarballStructure, tarExtension, tmpDir string) (*httptest.Server, *bool) {
	called := false
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(responseStatusCode)
		if responseStatusCode == http.StatusOK {
			tarball := filepath.Join(tmpDir, tarballStructure+tarExtension)
			err := archiver.Archive([]string{filepath.Join("testdata", tarballStructure)}, tarball)
			require.NoError(t, err)
			bytes, err := os.ReadFile(tarball)
			require.NoError(t, err)
			_, err = w.Write(bytes)
			require.NoError(t, err)
		}
	})), &called
}
