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
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/tetratelabs/func-e/internal/envoy"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/test"
	"github.com/tetratelabs/func-e/internal/test/morerequire"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/version"
)

func TestEnsurePatchVersion_Remote(t *testing.T) {
	server, o := setupTestServer(t)
	defer server.Close()

	// Ensure that when we ask for a minor, the latest version is returned from the remote JSON
	v := version.LastKnownEnvoyMinor
	pv, err := ensurePatchVersion(context.Background(), o, v)
	require.NoError(t, err)
	require.Equal(t, version.LastKnownEnvoy, pv)
}

func TestEnsurePatchVersion_NoOpWhenAlreadyAPatchVersion(t *testing.T) {
	v := version.PatchVersion("1.19.1")
	pv, err := ensurePatchVersion(context.Background(), &globals.GlobalOpts{}, v)
	require.NoError(t, err)
	require.Equal(t, v, pv)
}

func TestEnsurePatchVersion_FallbackOnLookupFailure(t *testing.T) {
	server, o := setupTestServer(t)
	defer server.Close()

	minor := version.MinorVersion("1.12")
	installedPatch := version.PatchVersion("1.12.1")

	lastKnownEnvoyDir := filepath.Join(o.HomeDir, "versions", installedPatch.String())
	require.NoError(t, os.MkdirAll(lastKnownEnvoyDir, 0700))
	morerequire.RequireSetMtime(t, lastKnownEnvoyDir, "2020-12-31")

	// Stop the server to simulate an outage
	server.Close()

	// Ensure that when we ask for a minor, the latest version is returned from the filesystem
	pv, err := ensurePatchVersion(context.Background(), o, minor)
	require.NoError(t, err)
	require.Equal(t, installedPatch, pv)
}

func TestEnsurePatchVersion_RaisesErrorWhenNothingInstalled(t *testing.T) {
	server, o := setupTestServer(t)
	// Stop the server to simulate an outage
	server.Close()

	// Since we have nothing local to fall back to, we should raise the remote error
	_, err := ensurePatchVersion(context.Background(), o, version.LastKnownEnvoyMinor)
	require.Error(t, err)
}

func setupTestServer(t *testing.T) (*httptest.Server, *globals.GlobalOpts) {
	server := test.RequireEnvoyVersionsTestServer(t, version.LastKnownEnvoy)
	o := &globals.GlobalOpts{EnvoyVersionsURL: server.URL + "/envoy-versions.json", HomeDir: t.TempDir()}
	o.FuncEVersions = envoy.NewFuncEVersions(o.EnvoyVersionsURL, globals.DefaultPlatform, "dev")
	return server, o
}
