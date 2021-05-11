// Copyright 2020 Tetrate
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
	"fmt"
	"regexp"
	"strings"
)

// Reference identifies an Envoy release provided by getenvoy.io.
type Reference struct {
	Flavor  string
	Version string
	// Platform is an lower-hyphen representation of the platform. Ex. "darwin" or "linux-glibc"
	Platform string
}

// platformFromEnum takes a Reference.Platform like "LINUX_GLIBC" and returns a string variant like "linux-glibc"
func platformFromEnum(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "_", "-")
	return s
}

var (
	referenceFormat = regexp.MustCompile(`^([^:]+\:)?([\w\d-\._]+)/?([\w\d-\._]+)?$`)
)

// ParseReference parses a given text as a Reference.
func ParseReference(text string) (*Reference, error) {
	matches := referenceFormat.FindStringSubmatch(text)
	if len(matches) != 4 {
		return nil, fmt.Errorf("%q is not a valid GetEnvoy reference. Expected format: [<flavor>:]<version>[/<platform>]", text)
	}
	flavor := strings.ToLower(matches[1])
	if flavor == "" {
		flavor = "standard"
	} else {
		flavor = flavor[:len(flavor)-1]
	}
	// Even though versions are usually semver, we currently allow anything in the version. Ex "latest" "1.19.2-alpha"
	version := strings.ToLower(matches[2])
	platform := platformFromEnum(matches[3])
	// If platform is empty, fill it in.
	if platform == "" {
		platform = currentPlatform()
	}
	return &Reference{flavor, version, platform}, nil
}

// String returns a plain version optionally platform qualified. Ex "1.17.2" or "1.17.2/darwin"
// Reference.Flavor is prefixed with a colon when non-standard. Ex "experimental:1.18.2"
func (r *Reference) String() string {
	if r.Flavor == "standard" { // don't clutter output in normal case
		return fmt.Sprintf("%s/%s", r.Version, r.Platform)
	}
	return fmt.Sprintf("%s:%s/%s", r.Flavor, r.Version, r.Platform)
}
