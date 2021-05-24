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
	"fmt"
	"io"
	"sort"

	"github.com/tetratelabs/getenvoy/internal/manifest"
)

// PrintVersions retrieves the manifest from the passed location and writes it to the passed writer
func PrintVersions(m *manifest.Manifest, goos string, w io.Writer) {
	p := manifest.BuildPlatform(goos)
	// print nothing if the only released "flavor" or the platform doesn't exist
	if m.Flavors["standard"] == nil || p == "" {
		return
	}

	// Build a list of versions for this platform
	var versions []string
	for _, v := range m.Flavors["standard"].Versions {
		for _, b := range v.Builds {
			if p == b.Platform {
				versions = append(versions, v.Name)
			}
		}
	}

	// Sort lexicographically descending, so that new versions appear first
	sort.Slice(versions, func(i, j int) bool {
		return versions[i] > versions[j]
	})

	// Print the versions
	for _, v := range versions {
		fmt.Fprintln(w, v) //nolint
	}
}
