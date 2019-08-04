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

package binary

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"os"

	"github.com/mholt/archiver"
	"github.com/stretchr/testify/assert"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
)

func TestRuntime_Fetch(t *testing.T) {
	defaultDarwinKey := &manifest.Key{Flavor: "standard", Version: "1.11.0", Platform: "darwin"}
	tests := []struct {
		name             string
		key              *manifest.Key
		tarballStructure string
		envoyLocation    string
		alreadyLocal     bool
		responseStatus   int
		wantErr          bool
		wantServerCalled bool
	}{
		{
			name:             "Downloads and untars envoy to local/key",
			key:              defaultDarwinKey,
			tarballStructure: "golden",
			responseStatus:   http.StatusOK,
			envoyLocation:    "builds/standard/1.11.0/darwin/envoy",
			wantServerCalled: true,
		},
		{
			name:             "Does nothing if it already has a local copy",
			key:              defaultDarwinKey,
			envoyLocation:    "builds/standard/1.11.0/darwin/envoy",
			alreadyLocal:     true,
			wantServerCalled: false,
		},
		{
			name:             "Handles directories called Envoy",
			key:              defaultDarwinKey,
			tarballStructure: "envoydirectory",
			responseStatus:   http.StatusOK,
			envoyLocation:    "builds/standard/1.11.0/darwin/envoy",
			wantServerCalled: true,
		},
		{
			name:             "errors if it can't find an envoy binary in tarball",
			key:              defaultDarwinKey,
			tarballStructure: "noenvoy",
			responseStatus:   http.StatusOK,
			wantErr:          true,
			wantServerCalled: true,
		},
		{
			name:             "errors if it gets !200 from download",
			key:              defaultDarwinKey,
			tarballStructure: "noenvoy",
			responseStatus:   http.StatusTeapot,
			wantErr:          true,
			wantServerCalled: true,
		},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			tmpDir, _ := ioutil.TempDir("", "getenvoy-test-")
			envoyLocation := filepath.Join(tmpDir, tc.envoyLocation)
			defer os.RemoveAll(tmpDir)
			mock, gotCalled := mockServer(tc.responseStatus, tc.tarballStructure, tmpDir)
			if tc.alreadyLocal {
				createLocalEnvoy(envoyLocation)
			}

			r := &Runtime{local: tmpDir}
			err := r.Fetch(tc.key, mock.URL)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				f, _ := os.Open(envoyLocation)
				bytes, _ := ioutil.ReadAll(f)
				assert.Contains(t, string(bytes), "some complied c++")
			}
			assert.Equal(t, tc.wantServerCalled, *gotCalled, "mismatch of expectations for calling of remote server")
		})
	}
}

func createLocalEnvoy(envoyLocation string) {
	dir, _ := filepath.Split(envoyLocation)
	os.MkdirAll(dir, 0750)
	f, _ := os.Create(envoyLocation)
	f.WriteString("some complied c++")
	f.Close()
}

func mockServer(responseStatusCode int, tarballStructure string, tmpDir string) (*httptest.Server, *bool) {
	called := false
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(responseStatusCode)
		if responseStatusCode == http.StatusOK {
			tarball := filepath.Join(tmpDir, tarballStructure+".tar.gz")
			archiver.Archive([]string{filepath.Join("testdata", tarballStructure)}, tarball)
			bytes, _ := ioutil.ReadFile(tarball)
			w.Write(bytes)
		}
	})), &called
}
