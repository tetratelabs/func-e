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

package manifest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tetratelabs/getenvoy/internal/transport"
)

// GetManifest returns a manifest from a remote URL. eg global.manifestURL.
func GetManifest(manifestURL string) (*Manifest, error) {
	// #nosec => This is by design, users can call out to wherever they like!
	resp, err := transport.Get(manifestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received %v status code from %v", resp.StatusCode, manifestURL)
	}
	body, err := io.ReadAll(resp.Body) // fully read the response
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", manifestURL, err)
	}

	result := Manifest{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshalling manifest: %w", err)
	}
	return &result, nil
}

// BuildPlatform returns the Build.Platform for the goos (runtime.GOOS value) or "" if not found.
func BuildPlatform(goos string) string {
	switch goos {
	case "darwin":
		return "DARWIN"
	case "linux":
		return "LINUX_GLIBC"
	default:
		return "" // no other platform was ever released
	}
}

// Manifest is derived from /manifest.proto
type Manifest struct {
	ManifestVersion string `json:"manifestVersion,omitempty"`
	// Key is the flavor name
	Flavors map[string]*Flavor `json:"flavors,omitempty"`
}

// Flavor is a type of release, typically always "standard".
type Flavor struct {
	// Name is almost always "standard"
	Name string `json:"name,omitempty"`
	// Key is the version's name
	Versions map[string]*Version `json:"versions,omitempty"`
}

// Version is an aggregation of Builds.
type Version struct {
	// Name is the Envoy version
	// Examples: 1.10.0, 1.11.0, nightly
	Name string `json:"name,omitempty"`
	// Key is the build's platform
	Builds map[string]*Build `json:"builds,omitempty"`
}

// Build associates a platform with a download URL.
type Build struct {
	// Platform ex "DARWIN", "WINDOWS", "LINUX_GLIBC"
	Platform            string `json:"platform,omitempty"`
	DownloadLocationURL string `json:"downloadLocationUrl,omitempty"`
}
