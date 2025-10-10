// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	"github.com/tetratelabs/func-e/api"
	internalhook "github.com/tetratelabs/func-e/internal/middleware"
	"github.com/tetratelabs/func-e/internal/opts"
)

// StartupHook runs just after Envoy logs "starting main dispatch loop".
//
// This provides access to two non-deterministic runtime values:
//  1. The run directory (where stdout, stderr, and pid file are written)
//  2. The admin address (which may be ephemeral)
//
// Startup hooks are considered mandatory and will stop the run with error if
// they fail. If your hook is optional, handle errors internally.
//
// Startup hooks run on the goroutine that consumes Envoy's STDERR. Keep them
// short or run long operations in a separate goroutine.
//
// To use a StartupHook, pass it via hook.WithStartupHook as a RunOption.
type StartupHook = internalhook.StartupHook

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
	return context.WithValue(ctx, internalhook.Key{}, mw)
}

// WithStartupHook returns a RunOption that sets a startup hook.
//
// This is an experimental API that should only be used by CLI entrypoints.
// See package documentation for usage constraints.
//
// If provided, this hook will REPLACE the default config dump hook.
// If you want to preserve default behavior, do not use this option.
func WithStartupHook(hook StartupHook) api.RunOption {
	return func(o *opts.RunOpts) {
		o.StartupHook = hook
	}
}
