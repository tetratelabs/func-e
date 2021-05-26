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
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/version"
)

const envoyVersionsPath = "./envoy_versions.json"

// GitHubRelease includes a subset of fields we use from https://docs.github.com/en/rest/reference/repos#releases
type gitHubRelease struct {
	// Name ex "v1.15.4"
	Name string `json:"name"`
	// PublishedAt ex "2021-05-11T19:11:09Z"
	PublishedAt string `json:"published_at"`
	// Draft should always be false or it isn't a stable release, yet
	Draft bool `json:"draft"`
	// PreRelease should always be false or it isn't a stable release, yet
	PreRelease bool `json:"prerelease"`
}

// TestEnvoyVersionsJson ensures tarball URLs in the Envoy versions JSON appear correct.
// These are not fetched to avoid causing excess load.
func TestEnvoyVersionsJson(t *testing.T) {
	releaseDates, err := getEnvoyReleaseDates()
	require.NoError(t, err)

	data, err := os.ReadFile(envoyVersionsPath)
	require.NoError(t, err)

	evs := version.EnvoyVersions{}
	err = json.Unmarshal(data, &evs)
	require.NoErrorf(t, err, "error parsing json from %s", envoyVersionsPath)
	require.Greaterf(t, len(evs.Versions), 2, "expected more than two versions")

	require.NotEmptyf(t, evs.LatestVersion, "latest version isn't in %s", envoyVersionsPath)
	require.Containsf(t, evs.Versions, evs.LatestVersion, "latest version isn't in the version list of %s", envoyVersionsPath)
	require.Equalf(t, evs.LatestVersion, version.LastKnownEnvoy, "version.LastKnownEnvoy doesn't match latest version in %s", envoyVersionsPath)

	// Ensure there's an option besides the latest version
	require.GreaterOrEqualf(t, len(evs.Versions), 2, "expected more than two versions")

	type testCase struct{ version, platform, tarballURL string }

	var tests []testCase
	for v, ev := range evs.Versions {
		require.NotEmptyf(t, releaseDates[v], "version %s is not a published envoyproxy/proxy release", v)
		require.Equalf(t, releaseDates[v], ev.ReleaseDate, "releaseDate for %s doesn't match envoyproxy/proxy", v)
		require.GreaterOrEqualf(t, len(ev.Tarballs), 2, "expected at least two platforms for version %s", v)

		for p, tb := range ev.Tarballs {
			tests = append(tests, testCase{v, p, tb})
		}
	}

	for _, tc := range tests {
		name := fmt.Sprintf("%s-%s", tc.version, tc.platform)
		tarballURL := tc.tarballURL
		t.Run(name, func(t *testing.T) {
			require.Regexpf(t, "https://.*.tar.(gz|xz)", tarballURL, "expected an https tar.gz or xz %s", tarballURL)
			res, err := http.Head(tarballURL)
			require.NoErrorf(t, err, "error from HEAD %s", tarballURL)
			defer res.Body.Close() //nolint

			require.NoErrorf(t, err, "error reading %s", tarballURL)
			require.Equalf(t, 200, res.StatusCode, "unexpected HTTP status reading %s", tarballURL)
			require.Greaterf(t, res.ContentLength, int64(5<<20), "expected at least 5MB size %s", tarballURL)
		})
	}
}

// getEnvoyReleases returns release metadata we can use to verify ours
func getEnvoyReleaseDates() (map[string]string, error) {
	url := "https://api.github.com/repos/envoyproxy/envoy/releases"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received %v status code from %v", resp.StatusCode, url)
	}
	body, err := io.ReadAll(resp.Body) // fully read the response
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", url, err)
	}

	var releases []GitHubRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("error unmarshalling GitHub Releases: %w", err)
	}

	m := map[string]string{}
	for _, r := range releases { //nolint:gocritic
		if r.Draft || r.PreRelease {
			continue
		}
		// clean inputs "v1.15.4" -> "2021-05-11T19:11:09Z" into "1.15.4" -> "2021-05-11"
		m[r.Name[1:]] = r.PublishedAt[0:10]
	}
	return m, nil
}
