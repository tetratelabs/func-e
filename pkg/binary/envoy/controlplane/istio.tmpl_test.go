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
	"github.com/stretchr/testify/require"
)

// TestIstioBootStrapTemplateEqualsReleaseJson ensures istioBootStrapTemplate is exactly the same as what would have been
// downloaded from the istio release for version IstioVersion.
func TestIstioBootStrapTemplateEqualsReleaseJson(t *testing.T) {
	// Retrieve Istio tarball os doesn't matter we only car about the bootstrap JSON
	url := fmt.Sprintf("https://github.com/istio/istio/releases/download/%s/istio-%s-linux.tar.gz", IstioVersion, IstioVersion)
	resp, err := http.Get(url)
	require.NoError(t, err, "error getting tarball for istio version %s", IstioVersion)

	defer resp.Body.Close() //nolint
	require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected HTTP status from %s", url)

	dst, err := ioutil.TempDir("", "")
	require.NoError(t, err, `ioutil.TempDir("", "") erred`)
	defer os.RemoveAll(dst)

	tarball := filepath.Join(dst, "istio.tar.gz")
	f, err := os.OpenFile(tarball, os.O_CREATE|os.O_WRONLY, 0600)
	require.NoError(t, err, "Couldn't open file %s", tarball)
	defer f.Close() //nolint

	_, err = io.Copy(f, resp.Body)
	require.NoError(t, err, "Couldn't download %s into %s", url, tarball)

	// Walk the tarball until we find the bootstrap
	bootstrapName := "envoy_bootstrap_v2.json"
	var bytes []byte
	err = archiver.Walk(tarball, func(f archiver.File) error {
		if f.Name() == bootstrapName {
			bytes, err = ioutil.ReadAll(f)
			require.NoError(t, err, "error reading %s in %s", bootstrapName, url)
		}
		return nil
	})

	require.NotNil(t, bytes, "couldn't find %s in %s", bootstrapName, url)
	require.Equal(t, istioBootStrapTemplate, string(bytes), "istioBootStrapTemplate isn't the same as the istio %s distribution", IstioVersion)
}
