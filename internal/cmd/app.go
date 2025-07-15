// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/tetratelabs/func-e/internal/api"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/version"
)

// NewApp create a new root command. The globals.GlobalOpts parameter allows tests to scope overrides, which avoids
// having to define a flag for everything needed in tests.
func NewApp(o *globals.GlobalOpts) *cli.App {
	var envoyVersionsURL, homeDir, platform string
	lastKnownEnvoyPath := moreos.ReplacePathSeparator(fmt.Sprintf("`$FUNC_E_HOME/versions/%s`", version.LastKnownEnvoy))

	app := cli.NewApp()
	app.Name = "func-e"
	app.HelpName = "func-e"
	app.Usage = `Install and run Envoy`
	// Keep lines at 77 to address leading indent of 3 in help statements
	app.UsageText = `To run Envoy, execute ` + "`func-e run -c your_envoy_config.yaml`" + `. This
downloads and installs the latest version of Envoy for you.

To list versions of Envoy you can use, execute ` + "`func-e versions -a`" + `. To
choose one, invoke ` + fmt.Sprintf("`func-e use %s`", version.LastKnownEnvoy) + `. This installs into
` + lastKnownEnvoyPath + `, if not already present. You may also use
minor version, such as ` + fmt.Sprintf("`func-e use %s`", version.LastKnownEnvoyMinor) + `.

You may want to override ` + "`$ENVOY_VERSIONS_URL`" + ` to supply custom builds or
otherwise control the source of Envoy binaries. When overriding, validate
your JSON first: ` + globals.DefaultEnvoyVersionsSchemaURL + `

Advanced:
` + "`FUNC_E_PLATFORM`" + ` overrides the host OS and architecture of Envoy binaries.
This is used when emulating another platform, e.g. x86 on Apple Silicon M1.
Note: Changing the OS value can cause problems as Envoy has dependencies,
such as glibc. This value must be constant within a ` + "`$FUNC_E_HOME`" + `.`
	app.Version = o.Version
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "home-dir",
			Usage:       "func-e home directory (location of installed versions and run archives)",
			DefaultText: globals.DefaultHomeDir,
			Destination: &homeDir,
			EnvVars:     []string{"FUNC_E_HOME"},
		},
		&cli.StringFlag{
			Name:        "envoy-versions-url",
			Usage:       "URL of Envoy versions JSON",
			DefaultText: globals.DefaultEnvoyVersionsURL,
			Destination: &envoyVersionsURL,
			EnvVars:     []string{"ENVOY_VERSIONS_URL"},
		},
		&cli.StringFlag{
			Name:        "platform",
			Usage:       "the host OS and architecture of Envoy binaries. Ex. darwin/arm64",
			DefaultText: "$GOOS/$GOARCH",
			Destination: &platform,
			EnvVars:     []string{"FUNC_E_PLATFORM"},
		},
	}
	app.Before = func(c *cli.Context) error {
		if err := api.InitializeGlobalOpts(o, envoyVersionsURL, homeDir, platform); err != nil {
			return NewValidationError(err.Error())
		}
		return nil
	}

	app.CustomAppHelpTemplate = cli.AppHelpTemplate
	cli.VersionPrinter = printVersion
	app.Commands = []*cli.Command{
		helpCommand,
		NewRunCmd(o),
		NewVersionsCmd(o),
		NewUseCmd(o),
		NewWhichCmd(o),
	}
	return app
}

// helpCommand allows us to hide the global flags which cleans up help and markdown
var helpCommand = &cli.Command{
	Name:      "help",
	Usage:     "Shows how to use a [command]",
	ArgsUsage: "[command]",
	Action: func(c *cli.Context) error {
		args := c.Args()
		if args.Present() {
			for _, cmd := range c.App.Commands {
				if cmd.Name == args.First() {
					cli.HelpPrinter(c.App.Writer, cmd.CustomHelpTemplate, cmd)
					return nil
				}
			}
			return fmt.Errorf("unknown command: %q", args.First())
		}
		return cli.ShowAppHelp(c)
	},
}

func printVersion(c *cli.Context) {
	moreos.Fprintf(c.App.Writer, "%v version %v\n", c.App.Name, c.App.Version)
}
