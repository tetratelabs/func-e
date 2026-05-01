// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"io"
	"net/http"
)

// HTTPClientFunc creates HTTP clients used during a run.
type HTTPClientFunc func() *http.Client

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
	HTTPClientFunc   HTTPClientFunc
	EnvoyPath        string      // Internal: path to the Envoy binary (for tests).
	StartupHook      StartupHook // Experimental: custom startup hook
}
