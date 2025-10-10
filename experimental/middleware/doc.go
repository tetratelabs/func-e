// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

// Package middleware provides experimental APIs for intercepting func-e's Run lifecycle.
//
// # Experimental
//
// This package is experimental and may change or be removed in future versions.
//
// # Usage Constraints
//
// This package should ONLY be used by CLI entry points (main.go level).
//
// DO NOT use this package in library code because:
//
//  1. Experimental packages in dependencies cause rev lock
//  2. CLI entry points would overwrite library hooks.
//
// # Example Usage
//
// This is intended for CLI entry points that need to intercept enforce runtime
// aspects such as the home directory or Envoy version without requiring
// library dependencies to expose api.RunOption directly.
//
//	func main() {
//	    ctx := context.Background()
//
//	    // Define a middleware that re-uses the CLI home dir for Envoy runs.
//	    homeDirMiddleware := func(next api.RunFunc) api.RunFunc {
//	        return func(ctx context.Context, args []string, options ...api.RunOption) error {
//	            options = append(options, api.HomeDir(cliHome))
//	            return next(ctx, args, options...)
//	        }
//	    }
//
//	    // Set the middleware in context, so that it will be used downstream.
//	    ctx = middleware.WithRunMiddleware(ctx, homeDirMiddleware)
//
//	    // Use that context when calling a library function that uses func-e.
//	    if err := library.Main(ctx, os.Args[1:]); err != nil {
//	        log.Fatal(err)
//	    }
//	}
package middleware
