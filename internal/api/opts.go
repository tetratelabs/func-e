// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

// Package opts holds shared configuration types for func-e options.
// This is internal and not intended for direct use by external packages.
package api

import (
	"io"
	"net/http"
)

// HTTPTransportFunc creates the HTTP client transport used during a run.
type HTTPTransportFunc func() http.RoundTripper

// RunOpts holds the configuration set by RunOptions.
type RunOpts struct {
	ConfigHome        string
	DataHome          string
	StateHome         string
	RuntimeDir        string
	RunID             string // Optional: custom run identifier for StateDir and RuntimeDir paths
	EnvoyVersion      string
	EnvoyVersionsURL  string
	Out               io.Writer
	EnvoyOut          io.Writer
	EnvoyErr          io.Writer
	HTTPTransportFunc HTTPTransportFunc
	EnvoyPath         string      // Internal: path to the Envoy binary (for tests).
	StartupHook       StartupHook // Experimental: custom startup hook
}
