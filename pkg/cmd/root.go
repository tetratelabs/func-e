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
	"os"
	"strconv"

	"github.com/tetratelabs/log"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	"github.com/tetratelabs/getenvoy/pkg/cmd/extension"
	"github.com/tetratelabs/getenvoy/pkg/common"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
	"github.com/tetratelabs/getenvoy/pkg/version"

	osutil "github.com/tetratelabs/getenvoy/pkg/util/os"
)

// globalOpts represents options that affect all getenvoy sub-commands.
//
// Options will have their values set according to the following rules:
//  1) to a value of the command line argument, e.g. `--home-dir`
//  2) otherwise, to a non-empty value of the environment variable, e.g. `GETENVOY_HOME`
//  3) otherwise, to the default value, e.g. `${HOME}/.getenvoy`
type globalOpts struct {
	HomeDir     string
	ManifestURL string
}

func newRootOpts() *globalOpts {
	return &globalOpts{
		HomeDir:     common.DefaultHomeDir(),
		ManifestURL: manifest.GetURL(),
	}
}

// NewRoot create a new root command and sets the version to the passed variable
// TODO: Add version support on the command
func NewRoot() *cobra.Command {
	rootOpts := newRootOpts()
	logOpts := log.DefaultOptions()
	configureLogging := enableLoggingConfig()

	rootCmd := &cobra.Command{
		Use:               "getenvoy",
		DisableAutoGenTag: true, // removes autogenerate on ___ from produced docs
		Short:             "Fetch, deploy and debug Envoy",
		Long: `Manage full lifecycle of Envoy including fetching binaries,
bootstrap generation and automated collection of access logs, Envoy state and machine state.`,
		Version: version.Build.Version,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if rootOpts.HomeDir == "" {
				return errors.New("GetEnvoy home directory cannot be empty")
			}
			common.HomeDir = rootOpts.HomeDir

			if rootOpts.ManifestURL == "" {
				return errors.New("GetEnvoy manifest URL cannot be empty")
			}
			if err := manifest.SetURL(rootOpts.ManifestURL); err != nil {
				return err
			}

			if configureLogging {
				return log.Configure(logOpts)
			}
			return nil
		},
	}

	rootCmd.AddCommand(NewRunCmd())
	rootCmd.AddCommand(NewListCmd())
	rootCmd.AddCommand(NewFetchCmd())
	rootCmd.AddCommand(NewDocCmd())
	rootCmd.AddCommand(extension.NewCmd())

	if configureLogging {
		logOpts.AttachFlags(rootCmd)
	}
	rootCmd.PersistentFlags().StringVar(&rootOpts.HomeDir, "home-dir", osutil.Getenv("GETENVOY_HOME", rootOpts.HomeDir),
		"GetEnvoy home directory (location of downloaded artifacts, caches, etc)")
	rootCmd.PersistentFlags().StringVar(&rootOpts.ManifestURL, "manifest", osutil.Getenv("GETENVOY_MANIFEST_URL", rootOpts.ManifestURL),
		"GetEnvoy manifest URL (source of information about available Envoy builds)")
	rootCmd.PersistentFlags().MarkHidden("manifest") // nolint
	return rootCmd
}

// enableLoggingConfig checks whether logging should be configurable.
//
// At the moment, logging configuration is disabled by default to avoid abundance of options.
//
// TODO(yskopets): consider introducing simplified configuration options.
func enableLoggingConfig() bool {
	if enable, err := strconv.ParseBool(os.Getenv("EXPERIMENTAL_GETENVOY_LOGGING_CONFIG")); err == nil {
		return enable
	}
	return false
}
