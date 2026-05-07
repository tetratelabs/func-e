// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/urfave/cli/v3"

	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/runtime"
	"github.com/tetratelabs/func-e/internal/version"
)

// NewApp create a new root command. The globals.GlobalOpts parameter allows tests to scope overrides, which avoids
// having to define a flag for everything needed in tests.
func NewApp(o *globals.GlobalOpts) *cli.Command {
	if o.HTTPClient == nil {
		o.HTTPClient = http.DefaultClient
	}

	var envoyVersionsURL, homeDir, configHome, dataHome, stateHome, runtimeDir, platform, runID string
	lastKnownEnvoyPath := fmt.Sprintf("`$FUNC_E_DATA_HOME/envoy-versions/%s`", version.LastKnownEnvoy)

	app := &cli.Command{
		Name:  "func-e",
		Usage: `Install and run Envoy`,
		// Keep lines at 77 to address leading indent of 3 in help statements
		UsageText: `To run Envoy, execute ` + "`func-e run -c your_envoy_config.yaml`" + `. This
downloads and installs the latest version of Envoy for you.

To list versions of Envoy you can use, execute ` + "`func-e versions -a`" + `. To
choose one, invoke ` + fmt.Sprintf("`func-e use %s`", version.LastKnownEnvoy) + `. This installs into
` + lastKnownEnvoyPath + `, if not already present. You may
also use minor version, such as ` + fmt.Sprintf("`func-e use %s`", version.LastKnownEnvoyMinor) + `.

You may want to override ` + "`$ENVOY_VERSIONS_URL`" + ` to supply custom builds or
otherwise control the source of Envoy binaries. When overriding, validate
your JSON first: ` + globals.DefaultEnvoyVersionsSchemaURL + `

Directory structure:
  ` + "`$FUNC_E_CONFIG_HOME`" + ` stores configuration files
    (default: ` + globals.DefaultConfigHome + `)
  ` + "`$FUNC_E_DATA_HOME`" + ` stores Envoy binaries
    (default: ` + globals.DefaultDataHome + `)
  ` + "`$FUNC_E_STATE_HOME`" + ` stores logs
    (default: ` + globals.DefaultStateHome + `)
  ` + "`$FUNC_E_RUNTIME_DIR`" + ` stores temporary files
    (default: ` + globals.DefaultRuntimeDir + `)

Advanced:
` + "`FUNC_E_PLATFORM`" + ` overrides the host OS and architecture of Envoy binaries.
This is used when emulating another platform, e.g. x86 on Apple Silicon M1.
Note: Changing the OS value can cause problems as Envoy has dependencies,
such as glibc. This value must be constant within a ` + "`$FUNC_E_DATA_HOME`" + `.`,
		Version: o.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "home-dir",
				Usage:       "func-e home directory",
				Destination: &homeDir,
				Local:       true,
				Sources:     cli.EnvVars("FUNC_E_HOME"),
			},
			&cli.StringFlag{
				Name:        "config-home",
				Usage:       "directory for configuration files",
				DefaultText: globals.DefaultConfigHome,
				Destination: &configHome,
				Local:       true,
				Sources:     cli.EnvVars("FUNC_E_CONFIG_HOME"),
			},
			&cli.StringFlag{
				Name:        "data-home",
				Usage:       "directory for Envoy binaries",
				DefaultText: globals.DefaultDataHome,
				Destination: &dataHome,
				Local:       true,
				Sources:     cli.EnvVars("FUNC_E_DATA_HOME"),
			},
			&cli.StringFlag{
				Name:        "state-home",
				Usage:       "directory for logs (used by run command)",
				DefaultText: globals.DefaultStateHome,
				Destination: &stateHome,
				Local:       true,
				Sources:     cli.EnvVars("FUNC_E_STATE_HOME"),
			},
			&cli.StringFlag{
				Name:        "runtime-dir",
				Usage:       "directory for temporary files (used by run command)",
				DefaultText: globals.DefaultRuntimeDir,
				Destination: &runtimeDir,
				Local:       true,
				Sources:     cli.EnvVars("FUNC_E_RUNTIME_DIR"),
			},
			&cli.StringFlag{
				Name:        "run-id",
				Usage:       "custom run identifier for logs/runtime directories (used by run command)",
				DefaultText: "auto-generated timestamp",
				Destination: &runID,
				Local:       true,
				Sources:     cli.EnvVars("FUNC_E_RUN_ID"),
			},
			&cli.StringFlag{
				Name:        "envoy-versions-url",
				Usage:       "URL of Envoy versions JSON",
				DefaultText: globals.DefaultEnvoyVersionsURL,
				Destination: &envoyVersionsURL,
				Local:       true,
				Sources:     cli.EnvVars("ENVOY_VERSIONS_URL"),
			},
			&cli.StringFlag{
				Name:        "platform",
				Usage:       "the host OS and architecture of Envoy binaries. Ex. darwin/arm64",
				DefaultText: "$GOOS/$GOARCH",
				Destination: &platform,
				Local:       true,
				Sources:     cli.EnvVars("FUNC_E_PLATFORM"),
			},
		},
		Before: func(ctx context.Context, _ *cli.Command) (context.Context, error) {
			if err := runtime.InitializeGlobalOpts(o, envoyVersionsURL, homeDir, configHome, dataHome, stateHome, runtimeDir, platform, runID); err != nil {
				return ctx, NewValidationError(err.Error())
			}
			return ctx, nil
		},
		Commands: []*cli.Command{
			helpCommand,
			NewRunCmd(o),
			NewVersionsCmd(o),
			NewUseCmd(o),
			NewWhichCmd(o),
		},
	}
	return app
}

// helpCommand allows us to hide the global flags which cleans up help and markdown
var helpCommand = &cli.Command{
	Name:      "help",
	Usage:     "Shows how to use a [command]",
	ArgsUsage: "[command]",
	Action: func(ctx context.Context, c *cli.Command) error {
		args := c.Args()
		if args.Present() {
			return cli.ShowCommandHelp(ctx, c.Root(), args.First())
		}
		return cli.ShowRootCommandHelp(c.Root())
	},
}
