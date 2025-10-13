// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"io"
	"testing"

	internalapi "github.com/tetratelabs/func-e/internal/api"
)

type FuncEFactory interface {
	New(ctx context.Context, t *testing.T, stdout, stderr io.Writer) (FuncE, error)
}

// FuncE abstracts func-e, so that the same tests can run for library calls and a compiled func-e binary.
type FuncE interface {
	// Run starts func-e with the given arguments and block until completion
	// The implementation should stop func-e if the context is canceled.
	// The returned error might be a process exit or context cancellation.
	Run(ctx context.Context, args []string) error

	// EnvoyPid is non-zero when the process launched.
	EnvoyPid() int

	// Interrupt signals the running func-e process to terminate gracefully
	Interrupt(context.Context) error

	// OnStart is called to get the Envoy admin client and run directory.
	// This method polls until Envoy's admin API is ready before returning.
	OnStart(ctx context.Context) (internalapi.AdminClient, error)
}
