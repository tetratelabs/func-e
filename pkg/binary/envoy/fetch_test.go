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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mholt/archiver/v3"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/manifest"
)

func TestRuntime_Fetch(t *testing.T) {
	defaultDarwinKey := &manifest.Key{Flavor: "standard", Version: "1.15.3", Platform: "darwin"}
	tests := []struct {
		name             string
		key              *manifest.Key
		tarballStructure string
		tarExtension     string
		envoyLocation    string
		libLocation      string
		alreadyLocal     bool
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
			envoyLocation:    "builds/standard/1.15.3/darwin/bin/envoy",
			libLocation:      "builds/standard/1.15.3/darwin/lib/somelib",
			wantServerCalled: true,
		},
		{
			name:             "Downloads and untars envoy and runtime libs to store/key/bin and store/key/lib",
			key:              defaultDarwinKey,
			tarballStructure: "envoy",
			tarExtension:     ".tar.xz",
			responseStatus:   http.StatusOK,
			envoyLocation:    "builds/standard/1.15.3/darwin/bin/envoy",
			libLocation:      "builds/standard/1.15.3/darwin/lib/somelib",
			wantServerCalled: true,
		},
		{
			name:             "Does nothing if it already has a local copy",
			key:              defaultDarwinKey,
			envoyLocation:    "builds/standard/1.15.3/darwin/bin/envoy",
			libLocation:      "builds/standard/1.15.3/darwin/lib/somelib",
			alreadyLocal:     true,
			wantServerCalled: false,
		},
		{
			name:             "errors if it can't find an envoy binary in tarball",
			key:              defaultDarwinKey,
			tarballStructure: "noenvoy",
			tarExtension:     ".tar.gz",
			responseStatus:   http.StatusOK,
			wantErr:          true,
			wantServerCalled: true,
		},
		{
			name:             "errors if it gets !200 from download",
			key:              defaultDarwinKey,
			tarballStructure: "noenvoy",
			tarExtension:     ".tar.gz",
			responseStatus:   http.StatusTeapot,
			wantErr:          true,
			wantServerCalled: true,
		},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			tmpDir, _ := ioutil.TempDir("", "getenvoy-test-")
			defer os.RemoveAll(tmpDir)
			envoyLocation := filepath.Join(tmpDir, tc.envoyLocation)
			libLocation := filepath.Join(tmpDir, tc.libLocation)
			mock, gotCalled := mockServer(tc.responseStatus, tc.tarballStructure, tc.tarExtension, tmpDir)
			if tc.alreadyLocal {
				createLocalFile(envoyLocation)
				createLocalFile(libLocation)
			}

			r := &Runtime{fetcher: fetcher{tmpDir}}
			err := r.Fetch(tc.key, mock.URL+"/"+tc.tarballStructure+tc.tarExtension)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
				for _, location := range []string{libLocation, envoyLocation} {
					f, _ := os.Open(location)
					bytes, _ := ioutil.ReadAll(f)
					require.Contains(t, string(bytes), "some c++")
				}
			}
			require.Equal(t, tc.wantServerCalled, *gotCalled, "mismatch of expectations for calling of remote server")
		})
	}
}

func createLocalFile(location string) {
	dir, _ := filepath.Split(location)
	os.MkdirAll(dir, 0750)
	f, _ := os.Create(location)
	f.WriteString("some c++")
	f.Close()
}

func mockServer(responseStatusCode int, tarballStructure, tarExtension, tmpDir string) (*httptest.Server, *bool) {
	called := false
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(responseStatusCode)
		if responseStatusCode == http.StatusOK {
			tarball := filepath.Join(tmpDir, tarballStructure+tarExtension)
			archiver.Archive([]string{filepath.Join("testdata", tarballStructure)}, tarball)
			bytes, _ := ioutil.ReadFile(tarball)
			w.Write(bytes)
		}
	})), &called
}
