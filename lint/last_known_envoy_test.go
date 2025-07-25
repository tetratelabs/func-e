// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package lint

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/envoy"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/version"
)

const lastKnownEnvoyFile = "../internal/version/last_known_envoy.txt"

// TestLastKnownEnvoyAvailableOnAllPlatforms ensures that an inconsistent Envoy release doesn't end up being suggested,
// or used in unit tests. This passes only when all platforms are available. This is most frequently inconsistent due to
// Homebrew (macOS) being a version behind latest Linux.
//
// This issues a remote call to the versions server, so shouldn't be a normal unit test (as they must pass offline).
// This is invoked via `make lint`.
func TestLastKnownEnvoyAvailableOnAllPlatforms(t *testing.T) {
	getEnvoyVersions := envoy.NewGetVersions(globals.DefaultEnvoyVersionsURL, globals.DefaultPlatform, "dev")
	evs, err := getEnvoyVersions(context.Background())
	require.NoError(t, err)

	var patchVersions []version.PatchVersion
	for v, r := range evs.Versions {
		if supportsAllPlatforms(r.Tarballs) {
			patchVersions = append(patchVersions, v)
		}
	}

	lastKnownEnvoy := version.FindLatestVersion(patchVersions)
	actual, err := os.ReadFile(lastKnownEnvoyFile)
	require.NoError(t, err)
	require.Equal(t, lastKnownEnvoy.String(), string(actual))
}

// allPlatforms are the platforms that Envoy is available on, which may differ than func-e.
// func-e's platforms are defined in the Makefile and are slightly wider due to the --platform flag.
var allPlatforms = []version.Platform{
	"linux/amd64",
	"linux/arm64",
	// We don't support darwin/amd64 anymore as the brew is no longer a reliable source for Envoy binaries.
	// For example, the latest Envoy version there was 1.33.x, and we switched to self-building for macOS
	// at the archive-envoy repo where we only build for darwin/arm64 due to technical difficulties. That
	// doesn't necessarily mean that we will never support darwin/amd64, but it is not a priority at the moment.
	"darwin/arm64",
}

func supportsAllPlatforms(r map[version.Platform]version.TarballURL) bool {
	for _, p := range allPlatforms {
		if _, ok := r[p]; !ok {
			return false
		}
	}
	return true
}
