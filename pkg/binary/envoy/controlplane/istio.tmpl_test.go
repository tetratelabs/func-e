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

package controlplane

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/mholt/archiver"
	"gotest.tools/assert"
)

func Test_VersionedIstioTemplate(t *testing.T) {
	t.Run(fmt.Sprintf("checking Istio bootstrap matches version %s", IstioVersion), func(t *testing.T) {
		got := retrieveIstioBootstrap(t)
		assert.Equal(t, istioBootStrapTemplate, string(got))
	})
}

func retrieveIstioBootstrap(t *testing.T) []byte {
	// Retrieve Istio tarball os doesn't matter we only car about the bootstrap JSON
	url := fmt.Sprintf("https://github.com/istio/istio/releases/download/%v/istio-%v-linux.tar.gz", IstioVersion, IstioVersion)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close() //nolint
	if resp.StatusCode != http.StatusOK {
		t.Errorf("received %v status code", resp.StatusCode)
	}
	dst := os.TempDir()
	defer os.RemoveAll(dst)
	tarball := filepath.Join(dst, "istio.tar.gz")
	f, err := os.OpenFile(tarball, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close() //nolint
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	// Walk the tarball until we find the bootstrap
	var bytes []byte
	if walkErr := archiver.Walk(tarball, func(f archiver.File) error {
		if f.Name() == "envoy_bootstrap_v2.json" {
			bytes, err = ioutil.ReadAll(f)
			if err != nil {
				return err
			}
		}
		return nil
	}); walkErr != nil {
		t.Fatal(err)
	}
	if bytes == nil {
		t.Fatal("no boostrap found")
	}
	return bytes
}
