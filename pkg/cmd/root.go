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
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tetratelabs/getenvoy/pkg/cmd/extension"
	"github.com/tetratelabs/getenvoy/pkg/cmd/run"
	"github.com/tetratelabs/getenvoy/pkg/globals"
	"github.com/tetratelabs/getenvoy/pkg/util/exec"
	"github.com/tetratelabs/getenvoy/pkg/version"
)

// NewRoot create a new root command. The globals.GlobalOpts parameter allows tests to scope overrides, which avoids
// having to define a flag for everything needed in tests.
func NewRoot(globalOpts *globals.GlobalOpts) *cobra.Command {
	// we set a default flags values eagerly so that handleFlagOverrides can validate if the user errored by clearing it
	// TODO: determine if it is worth the code and bother if the user or a tool they use accidentally set to empty!!
	homeDirFlag, manifestURLFlag := initializeGlobalOpts(globalOpts)

	rootCmd := &cobra.Command{
		Use:               "getenvoy",
		SilenceErrors:     true, // We can't adjust the error message on Ctrl+C, so we redo error logging in Execute
		SilenceUsage:      true, // We decided to return short usage form on error in Execute
		DisableAutoGenTag: true, // removes autogenerate on ___ from produced docs
		Short:             "Fetch, deploy and debug Envoy",
		Long: `Manage full lifecycle of Envoy including fetching binaries,
bootstrap generation and automated collection of access logs, Envoy state and machine state.`,
		Version: version.Build.Version, // TODO: Add version support on the command
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return handleFlagOverrides(globalOpts, homeDirFlag, manifestURLFlag)
		},
	}

	rootCmd.AddCommand(run.NewRunCmd(globalOpts))
	rootCmd.AddCommand(NewListCmd(globalOpts))
	rootCmd.AddCommand(NewFetchCmd(globalOpts))
	rootCmd.AddCommand(NewDocCmd())
	rootCmd.AddCommand(extension.NewCmd(globalOpts))

	rootCmd.PersistentFlags().StringVar(&homeDirFlag, "home-dir", homeDirFlag,
		"GetEnvoy home directory (location of downloaded artifacts, caches, etc)")

	rootCmd.PersistentFlags().StringVar(&manifestURLFlag, "manifest", manifestURLFlag,
		"GetEnvoy manifest URL (source of information about available Envoy builds)")
	rootCmd.PersistentFlags().MarkHidden("manifest") // nolint
	return rootCmd
}

func initializeGlobalOpts(globalOpts *globals.GlobalOpts) (string, string) {
	homeDirFlag := os.Getenv("GETENVOY_HOME")
	if homeDirFlag == "" && globalOpts.HomeDir == "" { // don't lookup homedir when overridden for tests
		homeDirFlag = globals.DefaultHomeDir()
	}

	manifestURLFlag := os.Getenv("GETENVOY_MANIFEST_URL")
	if manifestURLFlag == "" {
		manifestURLFlag = globals.DefaultManifestURL
	}
	return homeDirFlag, manifestURLFlag
}

func handleFlagOverrides(o *globals.GlobalOpts, homeDirFlag, manifestURLFlag string) error {
	if o.HomeDir == "" { // not overridden for tests
		if homeDirFlag == "" {
			return errors.New("GetEnvoy home directory cannot be empty")
		}
		abs, err := filepath.Abs(homeDirFlag)
		if err != nil {
			return err
		}
		o.HomeDir = abs
	}

	if o.ManifestURL == "" { // not overridden for tests
		if manifestURLFlag == "" {
			return errors.New("GetEnvoy manifest URL cannot be empty")
		}
		otherURL, err := url.Parse(manifestURLFlag)
		if err != nil || otherURL.Host == "" || otherURL.Scheme == "" {
			return fmt.Errorf("%q is not a valid manifest URL", manifestURLFlag)
		}
		o.ManifestURL = manifestURLFlag
	}
	return nil
}

// Execute prints short usage (similar to args failing to parse) on any error.
// This requires ExecuteC, which doesn't support context.Context, or looking up the subcommand with Find.
func Execute(cmd *cobra.Command) error {
	actualCmd, err := cmd.ExecuteC()
	if actualCmd != nil && err != nil { // both are always true on error
		var serr exec.ShutdownError
		if errors.As(err, &serr) { // in case of ShutdownError, we want to avoid any wrapper messages
			cmd.PrintErrln("NOTE:", serr.Error())
		} else {
			cmd.PrintErrln("Error:", err.Error())
			// actualCmd ensures command path includes the subcommand (ex "extension run")
			cmd.PrintErrf("\nRun '%v --help' for usage.\n", actualCmd.CommandPath())
			return err
		}
	}
	return nil
}
