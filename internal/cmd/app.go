// Copyright 2019 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"net/url"
	"os/user"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/tetratelabs/getenvoy/internal/globals"
	"github.com/tetratelabs/getenvoy/internal/version"
)

// NewApp create a new root command. The globals.GlobalOpts parameter allows tests to scope overrides, which avoids
// having to define a flag for everything needed in tests.
func NewApp(o *globals.GlobalOpts) *cli.App {
	var homeDir, envoyVersionsURL, userAgent string

	app := cli.NewApp()
	app.Name = "getenvoy"
	app.HelpName = "getenvoy"
	app.Usage = `Install and run Envoy`
	app.Version = version.GetEnvoy
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "home-dir",
			Usage:       "GetEnvoy home directory (location of installed versions and run archives)",
			DefaultText: globals.DefaultHomeDir,
			Destination: &homeDir,
			EnvVars:     []string{"GETENVOY_HOME"},
		},
		&cli.StringFlag{
			Name:        "envoy-versions-url",
			Usage:       "URL of Envoy versions JSON",
			DefaultText: globals.DefaultEnvoyVersionsURL,
			Destination: &envoyVersionsURL,
			EnvVars:     []string{"ENVOY_VERSIONS_URL"},
		},
		&cli.StringFlag{
			Name:        "internal-user-agent", // Hidden, but settable for release e2e tests
			Hidden:      true,                  // Hidden, but settable for release e2e tests
			Value:       globals.DefaultUserAgent,
			Destination: &userAgent,
		}}
	app.Before = func(c *cli.Context) error {
		o.UserAgent = userAgent
		if err := setHomeDir(o, homeDir); err != nil {
			return err
		}
		return setEnvoyVersionsURL(o, envoyVersionsURL)
	}

	app.HideHelp = true
	app.Commands = []*cli.Command{
		helpCommand,
		NewRunCmd(o),
		NewVersionsCmd(o),
		NewInstallCmd(o),
		NewInstalledCmd(o),
		NewDocCmd(),
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
			return cli.ShowCommandHelp(c, args.First())
		}
		return cli.ShowAppHelp(c)
	},
}

func setEnvoyVersionsURL(o *globals.GlobalOpts, versionsURL string) error {
	if o.EnvoyVersionsURL != "" { // overridden for tests
		return nil
	}
	if versionsURL == "" {
		o.EnvoyVersionsURL = globals.DefaultEnvoyVersionsURL
	} else {
		otherURL, err := url.Parse(versionsURL)
		if err != nil || otherURL.Host == "" || otherURL.Scheme == "" {
			return NewValidationError("%q is not a valid Envoy versions URL", versionsURL)
		}
		o.EnvoyVersionsURL = versionsURL
	}
	return nil
}

func setHomeDir(o *globals.GlobalOpts, homeDir string) error {
	if o.HomeDir != "" { // overridden for tests
		return nil
	}
	if homeDir == "" {
		u, err := user.Current()
		if err != nil || u.HomeDir == "" {
			return NewValidationError("unable to determine home directory. Set GETENVOY_HOME instead: %v", err)
		}
		o.HomeDir = filepath.Join(u.HomeDir, ".getenvoy")
	} else {
		abs, err := filepath.Abs(homeDir)
		if err != nil {
			return NewValidationError(err.Error())
		}
		o.HomeDir = abs
	}
	return nil
}
