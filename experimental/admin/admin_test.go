// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package admin_test

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	func_e "github.com/tetratelabs/func-e"
	"github.com/tetratelabs/func-e/api"
	"github.com/tetratelabs/func-e/experimental/admin"
	"github.com/tetratelabs/func-e/internal/test"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestWithStartupHook(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var actualAdminPort int
	var actualRunID string

	// Inject startup hook that captures the adminPort and runID
	startupHook := func(_ context.Context, adminClient admin.AdminClient, runID string) error {
		actualAdminPort = adminClient.Port()
		actualRunID = runID
		// Cancel immediately to stop Envoy and complete test quickly
		cancel()
		return nil
	}

	versionsServer := test.RequireEnvoyVersionsTestServer(t, version.LastKnownEnvoy)
	defer versionsServer.Close()

	opts := []api.RunOption{
		api.EnvoyOut(io.Discard),
		api.EnvoyErr(io.Discard),
		api.HomeDir(t.TempDir()),
		api.EnvoyVersionsURL(versionsServer.URL + "/envoy-versions.json"),
		admin.WithStartupHook(startupHook),
	}

	// Run with minimal Envoy config
	err := func_e.Run(ctx, []string{
		"--config-yaml",
		"admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}",
	}, opts...)

	// Expect nil error since Run returns nil on context cancellation (documented behavior)
	require.NoError(t, err)

	// Should get a real admin port, not the ephemeral input (0)
	require.NotZero(t, actualAdminPort)

	// Should get a non-empty runID
	require.NotEmpty(t, actualRunID, "runID should be provided to startup hook")
}
