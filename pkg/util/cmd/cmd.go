// Copyright 2020 Tetrate
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
	"strings"

	"github.com/spf13/cobra"
)

// Execute executes a given command and formats errors consistently.
func Execute(rootCmd *cobra.Command) error {
	rootCmd = rootCmd.Root()
	cmd, err := func() (*cobra.Command, error) {
		silenceErrors := rootCmd.SilenceErrors
		silenceUsage := rootCmd.SilenceUsage
		defer func() {
			rootCmd.SilenceErrors = silenceErrors
			rootCmd.SilenceUsage = silenceUsage
		}()
		rootCmd.SilenceErrors = true
		rootCmd.SilenceUsage = true
		return rootCmd.ExecuteC()
	}()
	if err == nil {
		return nil
	}
	if cmd.CalledAs() == "" {
		if cmd == rootCmd {
			// since errors were silenced, we need to print the error message
			printError(cmd, err)
			cmd.Printf("Run '%v --help' for usage.\n", cmd.CommandPath())
		}
		return err
	}
	if !cmd.SilenceErrors {
		printError(cmd, err)
	}
	if !cmd.SilenceUsage {
		cmd.Println(cmd.UsageString())
	}
	return err
}

// printError ensures that an error message is always followed by an empty line.
func printError(cmd *cobra.Command, err error) {
	message := err.Error()
	cmd.Println("Error:", message)
	if !strings.HasSuffix(message, "\n") {
		cmd.Println()
	}
}
