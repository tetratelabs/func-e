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

	"github.com/tetratelabs/func-e/internal/version"
)

// NewFuncEVersions creates a new Envoy versions fetcher.
func NewFuncEVersions(envoyVersionsURL string, p version.Platform, v string) version.FuncEVersions {
	feV := &funcEVersions{envoyVersionsURL: envoyVersionsURL, platform: p, version: v}
	feV.getFunc = feV.Get
	return feV
}

type funcEVersions struct {
	envoyVersionsURL string
	platform         version.Platform
	version          string

	// getFunc allows to override the release versions getter implementation.
	getFunc func(ctx context.Context) (version.ReleaseVersions, error)
}

// Get implements fetching the Envoy versions from the specified Envoy versions URL.
func (f *funcEVersions) Get(ctx context.Context) (version.ReleaseVersions, error) {
	result := version.ReleaseVersions{}
	// #nosec => This is by design, users can call out to wherever they like!
	resp, err := httpGet(ctx, f.envoyVersionsURL, f.platform, f.version)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close() //nolint

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("received %v status code from %v", resp.StatusCode, f.envoyVersionsURL)
	}
	body, err := io.ReadAll(resp.Body) // fully read the response
	if err != nil {
		return result, fmt.Errorf("error reading %s: %w", f.envoyVersionsURL, err)
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return result, fmt.Errorf("error unmarshalling Envoy versions: %w", err)
	}
	return result, nil
}

// FindLatestPatch implements finding the latest patch version for the given minor version or raises
// an error. The Envoy release versions fetching logic can be overridden by setting the getFunc with
// different implementation.
func (f *funcEVersions) FindLatestPatch(ctx context.Context, minorVersion version.MinorVersion) (version.PatchVersion, error) {
	releases, err := f.getFunc(ctx)
	if err != nil {
		return "", err
	}

	var latestVersion version.PatchVersion
	var latestPatch int
	for v := range releases.Versions {
		if v.ToMinor() != minorVersion {
			continue
		}

		if p := v.ParsePatch(); p >= latestPatch {
			latestPatch = p
			latestVersion = v
		}
	}

	if latestVersion == "" {
		return "", fmt.Errorf("couldn't find the latest patch for version %s", minorVersion)
	}
	return latestVersion, nil
}
