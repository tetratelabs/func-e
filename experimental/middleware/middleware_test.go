// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	func_e "github.com/tetratelabs/func-e"
	"github.com/tetratelabs/func-e/internal/version"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/api"
	"github.com/tetratelabs/func-e/experimental/middleware"
	internalmiddleware "github.com/tetratelabs/func-e/internal/middleware"
)

type arbitrary struct{}

// testCtx is an arbitrary, non-default context. Non-nil also prevents linter errors.
var testCtx = context.WithValue(context.Background(), arbitrary{}, "arbitrary")

func TestWithRunMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		input    middleware.RunMiddleware
		expected bool
	}{
		{
			name:     "returns input when middleware nil",
			input:    nil,
			expected: false,
		},
		{
			name: "decorates with middleware",
			input: func(next api.RunFunc) api.RunFunc {
				return next
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := middleware.WithRunMiddleware(testCtx, tt.input)
			if tt.expected {
				val := actual.Value(internalmiddleware.MiddlewareKey{})
				mw, ok := val.(func(api.RunFunc) api.RunFunc)
				require.NotNil(t, mw)
				require.True(t, ok)
			} else {
				require.Equal(t, testCtx, actual)
			}
		})
	}
}

func TestWithStartupHook(t *testing.T) {
	// Test that middleware.WithStartupHook returns a valid RunOption
	customHook := func(ctx context.Context, runDir, adminAddress string) error {
		return nil
	}

	actual := middleware.WithStartupHook(customHook)
	require.NotNil(t, actual)
}

func TestMiddleware_E2E(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	// Setup: known temp dir
	expectedHomeDir := t.TempDir()
	var actualRunDir string
	var actualAdminAddress string

	// Define middleware that:
	// 1. Overrides Out/EnvoyOut/EnvoyErr to io.Discard
	// 2. Sets HomeDir to known temp dir
	// 3. Injects StartupHook to capture runDir
	testMiddleware := func(next api.RunFunc) api.RunFunc {
		return func(ctx context.Context, args []string, options ...api.RunOption) error {
			// Override options
			options = append(options,
				api.EnvoyVersion(version.LastKnownEnvoy.String()),
				api.Out(io.Discard),
				api.EnvoyOut(io.Discard),
				api.EnvoyErr(io.Discard),
				api.HomeDir(expectedHomeDir),
			)

			// Inject startup hook that captures runDir and adminAddress
			startupHook := func(ctx context.Context, runDir, adminAddress string) error {
				actualRunDir = runDir
				actualAdminAddress = adminAddress
				// Cancel immediately to stop Envoy and complete test quickly
				cancel()
				return nil
			}
			options = append(options, middleware.WithStartupHook(startupHook))

			return next(ctx, args, options...)
		}
	}

	// Inject middleware via context
	ctx = middleware.WithRunMiddleware(ctx, testMiddleware)

	// Run with minimal Envoy config
	err := func_e.Run(ctx, []string{
		"--config-yaml",
		"admin: {address: {socket_address: {address: '127.0.0.1', port_value: 0}}}",
	})

	// Expect nil error since Run returns nil on context cancellation (documented behavior)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(expectedHomeDir, "runs"), filepath.Dir(actualRunDir))

	// Should get a real admin address, not the ephemeral input
	require.True(t, strings.HasPrefix(actualAdminAddress, "127.0.0.1:"))
	require.NotEqual(t, "127.0.0.1:0", actualAdminAddress)
}
