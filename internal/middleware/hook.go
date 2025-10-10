// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
)

// MiddlewareKey is a context.Context Value key. Its associated value should be a RunMiddleware.
type MiddlewareKey struct{}

// StartupHook runs just after Envoy logs "starting main dispatch loop".
//
// This provides access to two non-deterministic runtime values:
//  1. The run directory (where stdout, stderr and pid file are written)
//  2. The admin address (which may be ephemeral)
type StartupHook func(ctx context.Context, runDir, adminAddress string) error
