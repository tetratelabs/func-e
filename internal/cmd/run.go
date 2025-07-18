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

// NewRunCmd create a command responsible for starting an Envoy process
func NewRunCmd(o *globals.GlobalOpts) *cli.Command {
	runDirectoryExpression := "$FUNC_E_HOME/runs/$epochtime"
	cmd := &cli.Command{
		Name:            "run",
		Usage:           "Run Envoy with the given [arguments...] until interrupted",
		ArgsUsage:       "[arguments...]",
		SkipFlagParsing: true,
		Description: `To run Envoy, execute ` + "`func-e run -c your_envoy_config.yaml`" + `.

The first version in the below is run, controllable by the "use" command:
` + fmt.Sprintf("```\n%s\n```", envoy.VersionUsageList()) + `
The version to use is downloaded and installed, if necessary.

Envoy interprets the '[arguments...]' and runs in the current working
directory (aka $PWD) until func-e is interrupted (ex Ctrl+C, Ctrl+Break).

Envoy's process ID and console output write to "envoy.pid", stdout.log" and
"stderr.log" in the run directory (` + fmt.Sprintf("`%s`", runDirectoryExpression) + `).
When interrupted, shutdown hooks write files including network and process
state. On exit, these archive into ` + fmt.Sprintf("`%s.tar.gz`", runDirectoryExpression),
		Before: func(c *cli.Context) error {
			if err := api.EnsureEnvoyVersion(c.Context, o); err != nil {
				return NewValidationError(err.Error())
			}
			return nil
		},
		Action: func(c *cli.Context) error {
			o.EnvoyOut = c.App.Writer
			o.EnvoyErr = c.App.ErrWriter
			return api.Run(c.Context, o, c.Args().Slice())
		},
		CustomHelpTemplate: cli.CommandHelpTemplate,
	}
	return cmd
}
