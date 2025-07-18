// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

// Package version declares types for each string to keep strict coupling to the JSON schema
package version

import (
	"context"
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
	versionPattern      = regexp.MustCompile(`^([1-9][0-9]*\.[0-9]+)(\.[0-9]+)?(` + debugSuffix + `)?$`)
)

// debugSuffix is used to implement PatchVersion.ToMinor
const debugSuffix = "_debug"

// Version is a union type that allows commands to operate regardless of whether the input is a MinorVersion or a
// PatchVersion.
type Version interface {
	// String allows access to the underlying representation. Ex. "1.18", "1.18_debug", "1.19.3_debug"
	String() string
	// ToMinor returns a variant used to look up the latest Patch.
	ToMinor() MinorVersion
}

// MinorVersion is desired release from https://github.com/envoyproxy/envoy/blob/main/RELEASES.md, as a minor version.
// String will return a placeholder for the latest PatchVersion. Ex "1.18" or "1.20_debug"
type MinorVersion string

// NewMinorVersion returns a MinorVersion for a valid input like "1.19" or empty if invalid.
func NewMinorVersion(input string) MinorVersion {
	if matched := versionPattern.FindStringSubmatch(input); len(matched) == 4 && matched[0] != "" && matched[2] == "" {
		return MinorVersion(input)
	}
	return ""
}

// PatchVersion is a release version from https://github.com/envoyproxy/envoy/releases, without a 'v' prefix.
// This is the same form as "Version" in release-versions-schema.json. Ex "1.18.3" or "1.20.1_debug"
type PatchVersion string

// NewPatchVersion returns a PatchVersion for a valid input like "1.19.1" or empty if invalid.
func NewPatchVersion(input string) PatchVersion {
	if matched := versionPattern.FindStringSubmatch(input); len(matched) == 4 && matched[0] != "" && matched[2] != "" {
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
	matched := versionPattern.FindStringSubmatch(v.String())
	if matched == nil {
		return "" // impossible if created via NewVersion or NewPatchVersion
	}
	return MinorVersion(matched[1] + matched[3]) // ex. "1.18" + ""  or "1.18" + "_debug"
}

// Patch attempts to parse a Patch number from the Version.String.
// This will always succeed when created via NewVersion or NewPatchVersion
func (v PatchVersion) Patch() int {
	matched := versionPattern.FindStringSubmatch(v.String())
	if matched == nil {
		return 0 // impossible if created via NewVersion or NewPatchVersion
	}
	i, _ := strconv.Atoi(matched[2][1:]) // matched[2] will look like .1 or .10
	return i
}

// FindLatestPatchVersion finds the latest Patch version for the given minor version or empty if not found.
func FindLatestPatchVersion(patchVersions []PatchVersion, minorVersion MinorVersion) PatchVersion {
	var latestVersion PatchVersion
	var latestPatch int
	for _, v := range patchVersions {
		if v.ToMinor() != minorVersion {
			continue
		}

		if p := v.Patch(); p >= latestPatch {
			latestPatch = p
			latestVersion = v
		}
	}
	return latestVersion
}

// FindLatestVersion finds the latest non-debug version or empty if the input is.
func FindLatestVersion(patchVersions []PatchVersion) PatchVersion {
	var latestVersion PatchVersion
	for _, v := range patchVersions {
		if strings.HasSuffix(v.String(), debugSuffix) {
			continue
		}

		// Until Envoy 2.0.0, minor versions are double-digit and can be lexicographically compared.
		// Patches have to be numerically compared, as they can be single or double-digit (albeit unlikely).
		if m := v.ToMinor(); m > latestVersion.ToMinor() ||
			m == latestVersion.ToMinor() && v.Patch() > latestVersion.Patch() {
			latestVersion = v
		}
	}
	return latestVersion
}

// GetReleaseVersions returns a version map from a remote URL. e.g. from globals.DefaultEnvoyVersionsURL
type GetReleaseVersions func(ctx context.Context) (*ReleaseVersions, error)

// ReleaseVersions primarily maps Version to TarballURL and tracks the LatestVersion
type ReleaseVersions struct {
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
