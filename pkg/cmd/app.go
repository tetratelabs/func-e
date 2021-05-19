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
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tetratelabs/getenvoy/internal/version"
	"github.com/tetratelabs/getenvoy/pkg/globals"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
)

// NewApp create a new root command. The globals.GlobalOpts parameter allows tests to scope overrides, which avoids
// having to define a flag for everything needed in tests.
func NewApp(o *globals.GlobalOpts) *cobra.Command {
	homeDir, manifestURL := initializeGlobalOpts(o)

	rootCmd := &cobra.Command{
		Use:               "getenvoy",
		SilenceErrors:     true, // We can't adjust the error message on Ctrl+C, so we redo error logging in Execute
		SilenceUsage:      true, // We decided to return short usage form on error in Execute
		DisableAutoGenTag: true, // removes autogenerate on ___ from produced docs
		Short:             "Fetch and run Envoy",
		Long:              `Manage Envoy lifecycle including fetching binaries and collection of process state.`,
		Version:           version.Current,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if err := setHomeDir(o, homeDir); err != nil {
				return err
			}
			return setManifestURL(o, manifestURL)
		},
	}

	rootCmd.AddCommand(NewRunCmd(o))
	rootCmd.AddCommand(NewListCmd(o))
	rootCmd.AddCommand(NewFetchCmd(o))
	rootCmd.AddCommand(NewDocCmd())
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true}) // only support -h

	rootCmd.PersistentFlags().StringVar(&homeDir, "home-dir", homeDir,
		"GetEnvoy home directory (location of downloaded artifacts, caches, etc)")

	rootCmd.PersistentFlags().StringVar(&manifestURL, "manifest", manifestURL,
		"GetEnvoy manifest URL (source of information about available Envoy builds)")
	rootCmd.PersistentFlags().MarkHidden("manifest") // nolint
	return rootCmd
}

func initializeGlobalOpts(o *globals.GlobalOpts) (homeDir, manifestURL string) {
	if o.HomeDir == "" { // don't lookup homedir when overridden for tests
		homeDir = os.Getenv("GETENVOY_HOME")
	}

	if o.ManifestURL == "" {
		manifestURL = os.Getenv("GETENVOY_MANIFEST_URL")
	}
	return
}

func setManifestURL(o *globals.GlobalOpts, manifestURL string) error {
	if o.ManifestURL == "" { // not overridden for tests
		if manifestURL == "" {
			manifestURL = globals.DefaultManifestURL
		}
		otherURL, err := url.Parse(manifestURL)
		if err != nil || otherURL.Host == "" || otherURL.Scheme == "" {
			return newValidationError("%q is not a valid manifest URL", manifestURL)
		}
		o.ManifestURL = manifestURL
	}
	return nil
}

func setHomeDir(o *globals.GlobalOpts, homeDir string) error {
	if o.HomeDir == "" { // not overridden for tests
		if homeDir == "" {
			u, err := user.Current()
			if err != nil || u.HomeDir == "" {
				return newValidationError("unable to determine home directory. Set GETENVOY_HOME instead: %v", err)
			}
			homeDir = filepath.Join(u.HomeDir, ".getenvoy")
		}
		abs, err := filepath.Abs(homeDir)
		if err != nil {
			return newValidationError(err.Error())
		}
		o.HomeDir = abs
	}
	return nil
}

func validateReferenceArg(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		return newValidationError("missing reference parameter")
	}
	if _, e := manifest.ParseReference(args[0]); e != nil {
		return newValidationError(e.Error())
	}
	return nil
}
