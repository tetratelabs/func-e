// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

// Package opts holds shared configuration types for func-e options.
// This is internal and not intended for direct use by external packages.
package api

import (
	"io"
	"net/http"
)

// ArgsIgnoreRest is Envoy's CLI separator: args after this are not parsed.
const ArgsIgnoreRest = "--"

// HTTPTransport creates the HTTP client transport used during a run.
type HTTPTransport func() http.RoundTripper

// RunOpts holds the configuration set by RunOptions.
type RunOpts struct {
	ConfigHome       string
	DataHome         string
	StateHome        string
	RuntimeDir       string
	RunID            string // Optional: custom run identifier for StateDir and RuntimeDir paths
	EnvoyVersion     string
	EnvoyVersionsURL string
	Out              io.Writer
	EnvoyOut         io.Writer
	EnvoyErr         io.Writer
	HTTPTransport    http.RoundTripper
	EnvoyPath        string      // Internal: path to the Envoy binary (for tests).
	StartupHook      StartupHook // Experimental: custom startup hook
}
