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
	"strconv"
	"strings"

	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/version"
)

// NewFuncEVersions creates a new Envoy versions fetcher.
func NewFuncEVersions(envoyVersionsURL string, p version.Platform, v version.Version) version.FuncEVersions {
	feV := &funcEVersions{envoyVersionsURL: envoyVersionsURL, platform: p, version: v}
	feV.getFunc = feV.Get
	return feV
}

type funcEVersions struct {
	envoyVersionsURL string
	platform         version.Platform
	version          version.Version

	// getFunc allows to override the release versions getter implementation.
	getFunc func(ctx context.Context) (version.ReleaseVersions, error)
}

var _ version.FuncEVersions = (*funcEVersions)(nil)

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
func (f *funcEVersions) FindLatestPatch(ctx context.Context, minorVersion version.Version) (version.Version, error) {
	var latestPatch int
	var latestVersion version.Version

	releases, err := f.getFunc(ctx)
	if err != nil {
		return "", err
	}
	for v := range releases.Versions {
		// The "." suffix is required to avoild false-matching, e.g. 1.1 to 1.18.
		if !strings.HasPrefix(string(v), string(minorVersion)+".") {
			continue
		}

		var matched [][]string
		if matched = globals.EnvoyMinorVersionPattern.FindAllStringSubmatch(string(v), -1); matched == nil {
			continue
		}
		for _, sub := range matched {
			// A matched patch component should look like ".4".
			if !strings.HasPrefix(sub[1], ".") {
				continue
			}
			var p int
			if p, err = strconv.Atoi(sub[1][1:]); err != nil {
				continue
			}
			if p >= latestPatch {
				latestPatch = p
				latestVersion = v
			}
		}
	}

	if latestVersion == "" {
		return "", fmt.Errorf("couldn't find latest version for %q", minorVersion)
	}
	return latestVersion, nil
}
