// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tetratelabs/func-e/internal/version"
)

// NewGetVersions creates a new Envoy versions fetcher.
// TODO: validate the data before returning it!
func NewGetVersions(envoyVersionsURL string, p version.Platform, v string) version.GetReleaseVersions {
	return func(ctx context.Context) (*version.ReleaseVersions, error) {
		// #nosec => This is by design, users can call out to wherever they like!
		resp, err := httpGet(ctx, envoyVersionsURL, p, v)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close() //nolint

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("received %v status code from %v", resp.StatusCode, envoyVersionsURL)
		}
		body, err := io.ReadAll(resp.Body) // fully read the response
		if err != nil {
			return nil, fmt.Errorf("error reading %s: %w", envoyVersionsURL, err)
		}

		result := version.ReleaseVersions{}
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("error unmarshalling Envoy versions: %w", err)
		}
		return &result, nil
	}
}
