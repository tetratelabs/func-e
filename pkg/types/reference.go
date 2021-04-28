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

package types

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

// PlatformFromEnum takes a Reference.Platform like "LINUX_GLIBC" and returns a string variant like "linux-glibc"
func PlatformFromEnum(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "_", "-")
	return s
}

var (
	referenceFormat = regexp.MustCompile(`^([\w\d-\._]+):([\w\d-\._]+)/?([\w\d-\._]+)?$`)
)

// ParseReference parses a given text as a Reference.
func ParseReference(text string) (*Reference, error) {
	matches := referenceFormat.FindStringSubmatch(text)
	if len(matches) != 4 {
		return nil, fmt.Errorf("%q is not a valid GetEnvoy reference. Expected format: <flavor>:<version>[/<platform>]", text)
	}
	return &Reference{strings.ToLower(matches[1]), strings.ToLower(matches[2]), PlatformFromEnum(matches[3])}, nil
}

func (r *Reference) String() string {
	if r.Platform == "" {
		return fmt.Sprintf("%s:%s", r.Flavor, r.Version)
	}
	return fmt.Sprintf("%s:%s/%s", r.Flavor, r.Version, r.Platform)
}
