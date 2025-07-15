// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/test"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestNewGetVersions(t *testing.T) {
	versionsServer := test.RequireEnvoyVersionsTestServer(t, version.LastKnownEnvoy)
	gv := NewGetVersions(versionsServer.URL+"/envoy-versions.json", globals.DefaultPlatform, "dev")

	evs, err := gv(context.Background())
	require.NoError(t, err)
	require.Contains(t, evs.Versions, version.LastKnownEnvoy)
}
