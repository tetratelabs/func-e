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

	"github.com/tetratelabs/getenvoy/internal/errors"
	"github.com/tetratelabs/getenvoy/internal/globals"
	"github.com/tetratelabs/getenvoy/internal/version"
)

// NewApp create a new root command. The globals.GlobalOpts parameter allows tests to scope overrides, which avoids
// having to define a flag for everything needed in tests.
func NewApp(o *globals.GlobalOpts) *cli.App {
	var homeDir, manifestURL string

	app := cli.NewApp()
	app.Name = "getenvoy"
	app.HelpName = "getenvoy"
	app.HideHelpCommand = true
	app.Usage = `Download and run Envoy`
	app.Version = version.GetEnvoy
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "home-dir",
			Usage:       "GetEnvoy home directory (location of downloaded artifacts, caches, etc)",
			DefaultText: globals.DefaultHomeDir,
			Destination: &homeDir,
			EnvVars:     []string{"GETENVOY_HOME"},
		},
		&cli.StringFlag{
			Name:        "manifest",
			Usage:       "GetEnvoy manifest URL (list of available Envoy versions)",
			Hidden:      true,
			DefaultText: globals.DefaultManifestURL,
			Destination: &manifestURL,
			EnvVars:     []string{"GETENVOY_MANIFEST_URL"},
		}}
	app.Before = func(c *cli.Context) error {
		if err := setHomeDir(o, homeDir); err != nil {
			return err
		}
		return setManifestURL(o, manifestURL)
	}

	app.Commands = []*cli.Command{
		NewRunCmd(o),
		NewVersionsCmd(o),
		NewInstallCmd(o),
		NewDocCmd(),
	}
	return app
}

func setManifestURL(o *globals.GlobalOpts, manifestURL string) error {
	if o.ManifestURL != "" { // overridden for tests
		return nil
	}
	if manifestURL == "" {
		o.ManifestURL = globals.DefaultManifestURL
	} else {
		otherURL, err := url.Parse(manifestURL)
		if err != nil || otherURL.Host == "" || otherURL.Scheme == "" {
			return errors.NewValidationError("%q is not a valid manifest URL", manifestURL)
		}
		o.ManifestURL = manifestURL
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
			return errors.NewValidationError("unable to determine home directory. Set GETENVOY_HOME instead: %v", err)
		}
		o.HomeDir = filepath.Join(u.HomeDir, ".getenvoy")
	} else {
		abs, err := filepath.Abs(homeDir)
		if err != nil {
			return errors.NewValidationError(err.Error())
		}
		o.HomeDir = abs
	}
	return nil
}

func validateVersionArg(c *cli.Context) error {
	if c.NArg() == 0 {
		return errors.NewValidationError("missing version parameter")
	}
	return nil
}
