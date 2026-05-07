// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/tetratelabs/func-e/internal/envoy"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/runtime"
)

// NewWhichCmd create a command responsible for downloading printing the path to the Envoy binary
func NewWhichCmd(o *globals.GlobalOpts) *cli.Command {
	return &cli.Command{
		Name:        "which",
		Usage:       `Prints the path to the Envoy binary used by the "run" command`,
		HideHelp:    true,
		Description: `The binary is downloaded as necessary. The version is controllable by the "use" command`,
		Before: func(ctx context.Context, _ *cli.Command) (context.Context, error) {
			// no logging on version query/download. This is deferred until we know we are executing "which"
			o.Quiet = true
			return ctx, runtime.EnsureEnvoyVersion(ctx, o)
		},
		Action: func(ctx context.Context, _ *cli.Command) error {
			// Create base XDG directories before any file operations
			if err := o.Mkdirs(); err != nil {
				return err
			}
			ev, err := envoy.InstallIfNeeded(ctx, o)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(o.Out, "%s\n", ev)
			return nil
		},
	}
}
