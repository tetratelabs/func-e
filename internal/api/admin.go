// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"io"
	"net/http"
	"time"
)

// AdminClient provides methods to interact with Envoy's admin API.
//
// This is typically created via envoy.NewAdminClient, which polls for the
// admin port and PID from the run directory.
type AdminClient interface {
	// Port returns the Envoy admin API port.
	Port() int

	// Pid returns the Envoy process ID.
	Pid() int32

	// RunDir returns the directory where files like stdout.log are written to.
	RunDir() string

	// Get returns the content at the given path or an error if != 200
	Get(ctx context.Context, path string) ([]byte, error)

	// IsReady returns nil if Envoy is ready to accept requests, or an error otherwise.
	//
	// This is a convenience form of Get on the /ready endpoint.
	IsReady(ctx context.Context) error

	// AwaitReady polls IsReady until it succeeds or the context is cancelled.
	// On timeout/cancellation, returns the last error from IsReady if available,
	// otherwise returns the context error.
	AwaitReady(ctx context.Context, tickDuration time.Duration) error

	// NewListenerRequest creates an HTTP request against a named listener.
	// Similar to http.NewRequestWithContext, but targets the specified listener (e.g., "main").
	// The path parameter should include the path, query, and fragment (e.g., "/path?query#fragment").
	NewListenerRequest(ctx context.Context, name, method, path string, body io.Reader) (*http.Request, error)
}

// StartupHook runs once the Envoy admin server is ready.
//
// Note: Startup hooks are considered mandatory and will stop the run with
// error if failed. If your hook is optional, rescue panics and log your own
// errors.
type StartupHook func(ctx context.Context, adminClient AdminClient) error
