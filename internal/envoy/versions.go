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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tetratelabs/getenvoy/internal/version"
)

// GetEnvoyVersions returns a version map from a remote URL. eg globals.DefaultEnvoyVersionsURL.
func GetEnvoyVersions(ctx context.Context, envoyVersionsURL string, p version.Platform, v version.Version) (version.ReleaseVersions, error) {
	result := version.ReleaseVersions{}
	// #nosec => This is by design, users can call out to wherever they like!
	resp, err := httpGet(ctx, envoyVersionsURL, p, v)
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
