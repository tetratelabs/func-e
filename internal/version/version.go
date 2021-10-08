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
	"strings"
)

//go:embed last_known_envoy.txt
var lastKnownEnvoy string

// LastKnownEnvoy is the last known Envoy Version, used to ensure help statements aren't out-of-date.
// This is derived from /site/envoy-versions.json, but not used directly because go:embed requires a file read from the
// current directory tree.
//
// This is different than the "latestVersion" because this is built into the binary. For example, after the binary is
// built, a more recent "latestVersion" can be used, even if the help statements only know about the one from compile
// time.
var LastKnownEnvoy = Version(lastKnownEnvoy)

// LastKnownMinorVersionEnvoy is LastKnownEnvoy wihout the patch component.
var LastKnownMinorVersionEnvoy = Version(lastKnownEnvoy[:strings.LastIndex(lastKnownEnvoy, ".")])

// ReleaseVersions primarily maps Version to TarballURL and tracks the LatestVersion
type ReleaseVersions struct {
	// LatestVersion is the latest stable Version
	LatestVersion Version `json:"latestVersion"`
	// Versions maps a Version to its Release
	Versions map[Version]Release `json:"versions"`
	// SHA256Sums maps a Tarball to its SHA256Sum
	SHA256Sums map[Tarball]SHA256Sum `json:"sha256sums"`
}

// Version is a release version from https://github.com/envoyproxy/envoy/releases, without a 'v' prefix. Ex "1.18.3"
type Version string

// IsDebug shows if the version is a debug version
func (v *Version) IsDebug() bool {
	return strings.HasSuffix(string(v), "_debug")
}

// MinorPrefix expects the v is a valid version pattern
// extracts the v and returns the EnvoyStrictMinorVersionPattern without debug component
// withTrailingDot indicate whether the result would have a "." suffix or not
// The "." suffix is required to avoid false-matching, e.g. 1.1 to 1.18.
// e.g: 1.19.1_debug -> 1.19.
func (v *Version) MinorPrefix() string {
	withoutDebug := strings.Split(string(v), "_debug")[0]
	splitVersion := strings.Split(withoutDebug, ".")
	minorPrefix := fmt.Sprintf("%s.%s", splitVersion[0], splitVersion[1])

	if withTrailingDot {
		minorPrefix += "."
	}

	return minorPrefix
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
