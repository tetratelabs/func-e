// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package admin_test

import (
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	func_e "github.com/tetratelabs/func-e"
	"github.com/tetratelabs/func-e/api"
	"github.com/tetratelabs/func-e/experimental/admin"
)

func TestWithStartupHook(t *testing.T) {
	// Test that middleware.WithStartupHook returns a valid RunOption
	customHook := func(ctx context.Context, adminClient admin.AdminClient) error {
		return nil
	}

	actual := admin.WithStartupHook(customHook)
	require.NotNil(t, actual)
}

func TestWithStartupHook_E2E(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var actualRunDir string
	var actualAdminPort int

	// Inject startup hook that captures runDir and adminPort
	startupHook := func(ctx context.Context, adminClient admin.AdminClient) error {
		actualRunDir = adminClient.RunDir()
		actualAdminPort = adminClient.Port()
		// Cancel immediately to stop Envoy and complete test quickly
		cancel()
		return nil
	}

	// Run with minimal Envoy config
	err := func_e.Run(ctx, []string{
		"--config-yaml",
		"admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}",
	}, api.Out(io.Discard), api.EnvoyOut(io.Discard), api.EnvoyErr(io.Discard), admin.WithStartupHook(startupHook))

	// Expect nil error since Run returns nil on context cancellation (documented behavior)
	require.NoError(t, err)

	// Run dir should have been caught
	require.FileExists(t, filepath.Join(actualRunDir, "stdout.log"))

	// Should get a real admin port, not the ephemeral input (0)
	require.NotZero(t, actualAdminPort)
}
