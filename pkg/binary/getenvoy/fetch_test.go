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

package getenvoy

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
		wantErr          bool
	}{
		{
			name:             "Downloads and untars envoy to local/key",
			key:              defaultDarwinKey,
			tarballStructure: "golden",
			envoyLocation:    "standard/1.11.0/darwin/envoy",
		},
		{
			name:             "Handles directories called Envoy",
			key:              defaultDarwinKey,
			tarballStructure: "envoydirectory",
			envoyLocation:    "standard/1.11.0/darwin/envoy",
		},
		{
			name:             "errors if it can't find an envoy binary in tarball",
			key:              defaultDarwinKey,
			tarballStructure: "noenvoy",
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			tmpDir, _ := ioutil.TempDir("", "getenvoy-test-")
			defer os.RemoveAll(tmpDir)
			mock := mockServer(http.StatusOK, tc.tarballStructure, tmpDir)

			r := &Runtime{local: tmpDir}
			err := r.Fetch(tc.key, mock.URL)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				f, _ := os.Open(filepath.Join(tmpDir, tc.envoyLocation))
				bytes, _ := ioutil.ReadAll(f)
				assert.Contains(t, string(bytes), "some complied c++")
			}
		})
	}
}

func mockServer(responseStatusCode int, tarballStructure string, tmpDir string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(responseStatusCode)
		if responseStatusCode == http.StatusOK {
			tarball := filepath.Join(tmpDir, tarballStructure+".tar.gz")
			archiver.Archive([]string{filepath.Join("testdata", tarballStructure)}, tarball)
			bytes, _ := ioutil.ReadFile(tarball)
			w.Write(bytes)
		}
	}))
}
