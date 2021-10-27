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

// Package version declares types for each string to keep strict coupling to the JSON schema
package version

import (
	_ "embed" // We embed the Envoy version so that we can cache it in CI
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

//go:embed last_known_envoy.txt
var lastKnownEnvoy string

// LastKnownEnvoy is the last known Envoy PatchVersion, used to ensure help statements aren't out-of-date.
// This is derived from https://archive.tetratelabs.io/envoy/envoy-versions.json and validated in `make check`.
//
// This is different from the "latestVersion" because this is built into the binary. For example, after the binary is
// built, a more recent "latestVersion" can be used, even if the help statements only know about the one from compile
// time.
var LastKnownEnvoy = NewPatchVersion(lastKnownEnvoy)

var (
	// LastKnownEnvoyMinor is a convenience constant
	LastKnownEnvoyMinor = LastKnownEnvoy.ToMinor()
	versionPattern      = regexp.MustCompile(`^[1-9][0-9]*\.[0-9]+(\.[0-9]+)?(` + debugSuffix + `)?$`)
)

// debugSuffix is used to implement PatchVersion.ToMinor
const debugSuffix = "_debug"

// Version is a union type that allows commands to operate regardless of whether the input is a MinorVersion or a
// PatchVersion.
type Version interface {
	// String allows access to the underlying representation. Ex. "1.18", "1.18_debug", "1.19.3_debug"
	String() string
	// ToMinor returns a variant used to look up the latest patch.
	ToMinor() MinorVersion
}

// MinorVersion is desired release from https://github.com/envoyproxy/envoy/blob/main/RELEASES.md, as a minor version.
// String will return a placeholder for the latest PatchVersion. Ex "1.18" or "1.20_debug"
type MinorVersion string

// NewMinorVersion returns a MinorVersion for a valid input like "1.19" or empty if invalid.
func NewMinorVersion(input string) MinorVersion {
	if matched := versionPattern.FindStringSubmatch(input); len(matched) == 3 && matched[0] != "" && matched[1] == "" {
		return MinorVersion(input)
	}
	return ""
}

// PatchVersion is a release version from https://github.com/envoyproxy/envoy/releases, without a 'v' prefix.
// This is the same form as "Version" in release-versions-schema.json. Ex "1.18.3" or "1.20.1_debug"
type PatchVersion string

// NewPatchVersion returns a PatchVersion for a valid input like "1.19.1" or empty if invalid.
func NewPatchVersion(input string) PatchVersion {
	if matched := versionPattern.FindStringSubmatch(input); len(matched) == 3 && matched[0] != "" && matched[1] != "" {
		return PatchVersion(input)
	}
	return ""
}

// NewVersion returns a valid input or an error
func NewVersion(tag, input string) (Version, error) {
	if input == "" {
		return nil, fmt.Errorf("missing %s", tag)
	}
	if pv := NewPatchVersion(input); pv != "" {
		return pv, nil
	}
	if mv := NewMinorVersion(input); mv != "" {
		return mv, nil
	}
	return nil, fmt.Errorf("invalid %s: %q should look like %q or %q", tag, input, LastKnownEnvoy, LastKnownEnvoy.ToMinor())
}

// String satisfies Version.String
func (v MinorVersion) String() string {
	return string(v)
}

// ToMinor satisfies Version.ToMinor
func (v MinorVersion) ToMinor() MinorVersion {
	return v
}

// String satisfies Version.String
func (v PatchVersion) String() string {
	return string(v)
}

// ToMinor satisfies Version.ToMinor
func (v PatchVersion) ToMinor() MinorVersion {
	splitDebug := strings.Split(string(v), debugSuffix)
	splitVersion := strings.Split(splitDebug[0], ".")
	latestPatchFormat := fmt.Sprintf("%s.%s", splitVersion[0], splitVersion[1])

	if len(splitDebug) == 2 {
		latestPatchFormat += debugSuffix
	}

	return MinorVersion(latestPatchFormat)
}

// ParsePatch attempts to parse a patch number from the Version.String.
// This will always succeed when created via NewVersion or NewPatchVersion
func (v PatchVersion) ParsePatch() int {
	var matched []string
	if matched = versionPattern.FindStringSubmatch(v.String()); matched == nil {
		return 0 // impossible if created via NewVersion or NewPatchVersion
	}
	i, _ := strconv.Atoi(matched[1][1:]) // matched[1] will look like .1 or .10
	return i
}

// ReleaseVersions primarily maps Version to TarballURL and tracks the LatestVersion
type ReleaseVersions struct {
	// LatestVersion is the latest stable Version
	LatestVersion PatchVersion `json:"latestVersion"`
	// Versions maps a Version to its Release
	Versions map[PatchVersion]Release `json:"versions"`
	// SHA256Sums maps a Tarball to its SHA256Sum
	SHA256Sums map[Tarball]SHA256Sum `json:"sha256sums"`
}

// Platform encodes 'runtime.GOOS/runtime.GOARCH'. Ex "darwin/amd64"
type Platform string

// Tarball is the name of the tar.gz or tar.xz archive. Ex. "envoy-v1.18.3-linux-amd64.tar.xz"
// Minimum contents are "${dist}/bin/envoy[.exe]" Ex. "envoy-v1.18.3-linux-amd64/bin/envoy"
type Tarball string

// TarballURL is the HTTPS URL to the Tarball. SHA256Sums must include its base name.
type TarballURL string

// SHA256Sum is a SHA-256 lower-hex hash. Ex. "1274f55b3022bc1331aed41089f189094e00729981fe132ce00aac6272ea0770"
type SHA256Sum string

// ReleaseDate is the publish date of the release Version. Ex. "2021-05-11"
type ReleaseDate string

// Release primarily maps available Tarballs for a Version
type Release struct {
	ReleaseDate ReleaseDate

	// Tarballs are the Tarballs available by Platform
	Tarballs map[Platform]TarballURL `json:"tarballs,omitempty"`
}
