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
	"time"

	"github.com/tetratelabs/getenvoy/internal/version"
)

// GetEnvoyVersions returns a version map from a remote URL. eg globals.DefaultEnvoyVersionsURL.
func GetEnvoyVersions(envoyVersionsURL, userAgent string) (version.EnvoyVersions, error) {
	result := version.EnvoyVersions{}
	// #nosec => This is by design, users can call out to wherever they like!
	resp, err := httpGet(envoyVersionsURL, userAgent)
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

// AddVersions adds Envoy versions containing a release date for this platform
func AddVersions(out map[string]string, update map[string]version.EnvoyVersion, p string) error {
	for k, v := range update {
		if _, ok := v.Tarballs[p]; ok && out[k] == "" {
			if _, err := time.Parse("2006-01-02", v.ReleaseDate); err != nil {
				return fmt.Errorf("invalid releaseDate of version %q for platform %q: %w", k, p, err)
			}
			out[k] = v.ReleaseDate
		}
	}
	return nil
}
