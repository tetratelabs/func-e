// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	"github.com/tetratelabs/func-e/api"
	internalapi "github.com/tetratelabs/func-e/internal/api"
)

// RunMiddleware wraps an api.RunFunc to intercept and modify its behavior.
//
// The middleware can:
//   - Modify context, args, or options before calling next
//   - Add StartupHooks via hook.WithStartupHook
//   - Handle errors from next
//   - Perform pre/post processing
//
// See package documentation for usage constraints.
type RunMiddleware func(next api.RunFunc) api.RunFunc

// WithRunMiddleware returns a context that will cause run.Run to use the
// provided middleware to wrap the default RunFunc.
//
// Only the most recently set middleware will be used. If multiple callers
// set middleware, only the last one wins.
//
// This should only be called from CLI entrypoints. See package docs for details.
func WithRunMiddleware(ctx context.Context, middleware RunMiddleware) context.Context {
	if middleware == nil {
		return ctx
	}
	// Store as unnamed function type to enable type assertion in internal/run
	var mw func(api.RunFunc) api.RunFunc = middleware
	return context.WithValue(ctx, internalapi.RunMiddlewareKey{}, mw)
}
