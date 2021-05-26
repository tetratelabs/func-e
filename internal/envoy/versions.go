// Copyright 2021 Tetrate
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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sort"
	"text/tabwriter"

	"github.com/tetratelabs/getenvoy/internal/version"
)

// GetEnvoyVersions returns a version map from a remote URL. eg globals.DefaultEnvoyVersionsURL.
func GetEnvoyVersions(envoyVersionsURL string) (version.EnvoyVersions, error) {
	result := version.EnvoyVersions{}
	// #nosec => This is by design, users can call out to wherever they like!
	resp, err := httpGet(envoyVersionsURL)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close() //nolint

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("received %v status code from %v", resp.StatusCode, envoyVersionsURL)
	}
	body, err := io.ReadAll(resp.Body) // fully read the response
	if err != nil {
		return result, fmt.Errorf("error reading %s: %w", envoyVersionsURL, err)
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return result, fmt.Errorf("error unmarshalling Envoy versions: %w", err)
	}
	return result, nil
}

// PrintVersions retrieves the Envoy versions from the passed location and writes it to the passed writer
func PrintVersions(vs version.EnvoyVersions, p string, w io.Writer) error {
	// Build a list of Envoy versions with release date for this platform
	type versionReleaseDate struct{ version, releaseDate string }

	var rows []versionReleaseDate
	for v, t := range vs.Versions {
		if _, ok := t.Tarballs[p]; ok {
			rows = append(rows, versionReleaseDate{v, t.ReleaseDate})
		}
	}

	if len(rows) == 0 {
		return nil
	}

	// Sort so that new release dates appear first and on conflict choosing the higher version
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].releaseDate == rows[j].releaseDate {
			return rows[i].version > rows[j].version
		}
		return rows[i].releaseDate > rows[j].releaseDate
	})

	tw := tabwriter.NewWriter(w, 1, 8, 1, '\t', 0)
	if _, err := fmt.Fprintln(tw, "VERSION\tRELEASE_DATE"); err != nil {
		return err
	}

	for _, vr := range rows { //nolint:gocritic
		if _, err := fmt.Fprintf(tw, "%s\t%s\n", vr.version, vr.releaseDate); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// CurrentPlatform is the platform of the current process. This is used as a key in EnvoyVersion.Tarballs.
func CurrentPlatform() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}
