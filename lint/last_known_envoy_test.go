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

// requiredPlatforms are the platforms that Envoy is available on, which may differ than func-e.
// func-e's platforms are defined in the Makefile and are slightly wider due to the --platform flag.
var requiredPlatforms = []version.Platform{
	"linux/amd64",
	"linux/arm64",
	"darwin/amd64",
	// "darwin/arm64", TODO: https://github.com/envoyproxy/envoy/issues/1648
	"windows/amd64",
	// "windows/arm64", TODO: https://github.com/envoyproxy/envoy/issues/17572
}

func TestLastKnownEnvoyAvailableOnAllPlatforms(t *testing.T) {
	// We know that the canonical json never has a debug version, so that doesn't need to be handled.
	getEnvoyVersions := envoy.NewGetVersions(globals.DefaultEnvoyVersionsURL, globals.DefaultPlatform, "dev")
	evs, err := getEnvoyVersions(context.Background())
	require.NoError(t, err)

	lastKnownEnvoy := version.PatchVersion("1.10.0")
version:
	for v, r := range evs.Versions {
		// Ensure all platforms are available, or skip the version
		for _, p := range requiredPlatforms {
			if _, ok := r.Tarballs[p]; !ok {
				continue version
			}
		}

		// Until Envoy 2.0.0, minor versions are double-digit and can be lexicographically compared.
		// Patches have to be numerically compared, as they can be single or double-digit (albeit unlikely).
		if m := v.ToMinor(); m > lastKnownEnvoy.ToMinor() ||
			m == lastKnownEnvoy.ToMinor() && v.Patch() > lastKnownEnvoy.Patch() {
			lastKnownEnvoy = v
		}
	}

	actual, err := os.ReadFile(lastKnownEnvoyFile)
	require.NoError(t, err)
	require.Equal(t, lastKnownEnvoy.String(), string(actual))
}
