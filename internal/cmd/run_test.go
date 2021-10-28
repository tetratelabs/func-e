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

package cmd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/globals"

	"github.com/tetratelabs/func-e/internal/version"
)

func TestEnsurePatchVersion(t *testing.T) {
	versions := map[version.PatchVersion]version.Release{
		version.PatchVersion("1.18.3"):       {},
		version.PatchVersion("1.18.14"):      {},
		version.PatchVersion("1.18.4"):       {},
		version.PatchVersion("1.18.4_debug"): {},
	}

	o := &globals.GlobalOpts{
		GetEnvoyVersions: func(context.Context) (*version.ReleaseVersions, error) {
			return &version.ReleaseVersions{Versions: versions}, nil
		},
		HomeDir: t.TempDir(),
	}
	actual, err := ensurePatchVersion(context.Background(), o, version.MinorVersion("1.18"))
	require.NoError(t, err)
	require.Equal(t, version.PatchVersion("1.18.14"), actual)
}

func TestEnsurePatchVersion_NotFound(t *testing.T) {
	versions := map[version.PatchVersion]version.Release{
		version.PatchVersion("1.20.0"):    {},
		version.PatchVersion("1.1_debug"): {},
	}

	o := &globals.GlobalOpts{
		GetEnvoyVersions: func(context.Context) (*version.ReleaseVersions, error) {
			return &version.ReleaseVersions{Versions: versions}, nil
		},
		HomeDir: t.TempDir(),
	}
	_, err := ensurePatchVersion(context.Background(), o, version.MinorVersion("1.18"))
	require.EqualError(t, err, "couldn't find the latest patch for version 1.18")
}

func TestEnsurePatchVersion_NoOpWhenAlreadyAPatchVersion(t *testing.T) {
	expected := version.PatchVersion("1.19.1")
	actual, err := ensurePatchVersion(context.Background(), &globals.GlobalOpts{}, expected)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestEnsurePatchVersion_FallbackOnLookupFailure(t *testing.T) {
	o := &globals.GlobalOpts{
		GetEnvoyVersions: func(context.Context) (*version.ReleaseVersions, error) {
			return nil, errors.New("ice cream")
		},
		HomeDir: t.TempDir(),
	}

	lastKnownEnvoyDir := filepath.Join(o.HomeDir, "versions", "1.18.14")
	require.NoError(t, os.MkdirAll(lastKnownEnvoyDir, 0700))

	// Ensure that when we ask for a minor, the latest version is returned from the filesystem
	actual, err := ensurePatchVersion(context.Background(), o, version.MinorVersion("1.18"))
	require.NoError(t, err)
	require.Equal(t, version.PatchVersion("1.18.14"), actual)
}

func TestEnsurePatchVersion_RaisesErrorWhenNothingInstalled(t *testing.T) {
	o := &globals.GlobalOpts{
		GetEnvoyVersions: func(context.Context) (*version.ReleaseVersions, error) {
			return nil, errors.New("ice cream")
		},
		HomeDir: t.TempDir(),
	}

	// Since we have nothing local to fall back to, we should raise the remote error
	_, err := ensurePatchVersion(context.Background(), o, version.LastKnownEnvoyMinor)
	require.EqualError(t, err, "ice cream")
}
