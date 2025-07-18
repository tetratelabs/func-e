// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/tetratelabs/func-e/internal/api"
	"github.com/tetratelabs/func-e/internal/envoy"
	"github.com/tetratelabs/func-e/internal/globals"
)

// NewWhichCmd create a command responsible for downloading printing the path to the Envoy binary
func NewWhichCmd(o *globals.GlobalOpts) *cli.Command {
	return &cli.Command{
		Name:        "which",
		Usage:       `Prints the path to the Envoy binary used by the "run" command`,
		Description: `The binary is downloaded as necessary. The version is controllable by the "use" command`,
		Before: func(c *cli.Context) error {
			// no logging on version query/download. This is deferred until we know we are executing "which"
			o.Quiet = true
			return api.EnsureEnvoyVersion(c.Context, o)
		},
		Action: func(c *cli.Context) error {
			ev, err := envoy.InstallIfNeeded(c.Context, o)
			if err != nil {
				return err
			}
			fmt.Fprintf(o.Out, "%s\n", ev) //nolint:errcheck
			return nil
		},
		CustomHelpTemplate: cli.CommandHelpTemplate,
	}
}
