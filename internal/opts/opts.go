// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

// Package opts holds shared configuration types for func-e options.
// This is internal and not intended for direct use by external packages.
package opts

import "io"

// RunOpts holds the configuration set by RunOptions.
type RunOpts struct {
	HomeDir          string
	EnvoyVersion     string
	EnvoyVersionsURL string
	Out              io.Writer
	EnvoyOut         io.Writer
	EnvoyErr         io.Writer
	EnvoyPath        string // Internal: path to the Envoy binary (for tests).
}
