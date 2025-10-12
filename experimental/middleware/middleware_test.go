// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"

	func_e "github.com/tetratelabs/func-e"
	"github.com/tetratelabs/func-e/api"
	"github.com/tetratelabs/func-e/experimental/middleware"
	internalmiddleware "github.com/tetratelabs/func-e/internal/api"
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
				val := actual.Value(internalmiddleware.RunMiddlewareKey{})
				mw, ok := val.(func(api.RunFunc) api.RunFunc)
				require.NotNil(t, mw)
				require.True(t, ok)
			} else {
				require.Equal(t, testCtx, actual)
			}
		})
	}
}

func TestWithRunMiddleware_E2E(t *testing.T) {
	var stderr bytes.Buffer

	testMiddleware := func(next api.RunFunc) api.RunFunc {
		return func(ctx context.Context, args []string, options ...api.RunOption) error {
			// Override options to prove we override them
			options = append(options,
				api.Out(io.Discard),
				api.EnvoyOut(io.Discard),
				api.EnvoyErr(&stderr),
			)
			return next(ctx, args, options...)
		}
	}

	// Inject middleware via context
	ctx := middleware.WithRunMiddleware(t.Context(), testMiddleware)

	// Run with invalid Envoy config
	err := func_e.Run(ctx, []string{"--config-yaml", "foo"})

	// Expect it to have crashed due to the invalid args
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode())

	// Envoy should have written stderr to what we configured.
	require.NotEmpty(t, stderr)
}
