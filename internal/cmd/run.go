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

// NewRunCmd create a command responsible for starting an Envoy process
func NewRunCmd(o *globals.GlobalOpts) *cli.Command {
	stopOnFirstArg := 0
	cmd := &cli.Command{
		Name:         "run",
		Usage:        "Run Envoy with the given [arguments...] until interrupted",
		ArgsUsage:    "[arguments...]",
		StopOnNthArg: &stopOnFirstArg,
		HideHelp:     true,
		Description: `To run Envoy, execute ` + "`func-e run -c your_envoy_config.yaml`" + `.

The first version in the below is run, controllable by the "use" command:
` + fmt.Sprintf("```\n%s\n```", envoy.VersionUsageList()) + `
The version to use is downloaded and installed, if necessary.

Envoy interprets the '[arguments...]' and runs in the current working
directory (aka $PWD) until func-e is interrupted (ex Ctrl+C, Ctrl+Break).

Envoy's console output writes to "stdout.log" and "stderr.log" in the run directory
(` + fmt.Sprintf("`%s`", globals.DefaultStateHome) + `/envoy-logs/{runID}).`,
		Before: func(ctx context.Context, _ *cli.Command) (context.Context, error) {
			if o.EnvoyPath != "" { // custom binary, skip version resolution
				return ctx, nil
			}
			if err := runtime.EnsureEnvoyVersion(ctx, o); err != nil {
				return ctx, NewValidationError(err.Error())
			}
			return ctx, nil
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			o.EnvoyOut = c.Root().Writer
			o.EnvoyErr = c.Root().ErrWriter
			return runtime.Run(ctx, o, c.Args().Slice())
		},
	}
	return cmd
}
