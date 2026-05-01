// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"

	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/runtime"
)

type cmdRun struct {
	Args []string `arg:"" optional:"" help:"Arguments passed through to Envoy"`
}

func (c *cmdRun) Run(ctx context.Context, o *globals.GlobalOpts) error {
	if err := runtime.EnsureEnvoyVersion(ctx, o); err != nil {
		return NewValidationError(err.Error())
	}
	return runtime.Run(ctx, o, c.Args)
}
