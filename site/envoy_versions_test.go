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

package site

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/version"
)

const envoyVersionsPath = "./envoy_versions.json"

// TestEnvoyVersionsJson ensures tarball URLs in the Envoy versions JSON appear correct.
// These are not fetched to avoid causing excess load.
func TestEnvoyVersionsJson(t *testing.T) {
	data, err := os.ReadFile(envoyVersionsPath)
	require.NoError(t, err)

	evs := version.EnvoyVersions{}
	err = json.Unmarshal(data, &evs)
	require.NoErrorf(t, err, "error parsing json from %s", envoyVersionsPath)

	type testCase struct{ version, platform, tarballURL string }

	var tests []testCase
	for v, ev := range evs.Versions {
		for p, tb := range ev.Tarballs {
			tests = append(tests, testCase{v, p, tb})
		}
	}

	require.NotEmptyf(t, evs.LatestVersion, "latest version isn't in %s", envoyVersionsPath)
	require.Containsf(t, evs.Versions, evs.LatestVersion, "latest version isn't in the version list of %s", envoyVersionsPath)
	require.Equalf(t, evs.LatestVersion, version.Envoy, "version.Envoy doesn't match latest version in %s", envoyVersionsPath)

	for _, tc := range tests {
		name := fmt.Sprintf("%s-%s", tc.version, tc.platform)
		tarballURL := tc.tarballURL
		t.Run(name, func(t *testing.T) {
			require.Regexpf(t, "https://.*.tar.(gz|xz)", tarballURL, "expected an https tar.gz or xz %s", tarballURL)
			res, err := http.Head(tarballURL)
			defer res.Body.Close() //nolint

			require.NoErrorf(t, err, "error reading %s", tarballURL)
			require.Equalf(t, 200, res.StatusCode, "unexpected HTTP status reading %s", tarballURL)
			require.Greaterf(t, res.ContentLength, int64(5<<20), "expected at least 5MB size %s", tarballURL)
		})
	}
}
