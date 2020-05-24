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
	"github.com/spf13/cobra"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
	"github.com/tetratelabs/getenvoy/pkg/version"
)

// NewRoot create a new root command and sets the cliVersion to the passed variable
// TODO: Add version support on the command
func NewRoot() *cobra.Command {
	rootCmd.AddCommand(NewRunCmd())
	rootCmd.AddCommand(NewListCmd())
	rootCmd.AddCommand(NewFetchCmd())
	rootCmd.AddCommand(NewDocCmd())

	rootCmd.PersistentFlags().StringVar(&manifestURL, "manifest", manifest.DefaultURL, "sets the manifest URL")
	rootCmd.PersistentFlags().MarkHidden("manifest") // nolint
	return rootCmd
}

var (
	rootCmd = &cobra.Command{
		Use:               "getenvoy",
		DisableAutoGenTag: true, // removes autogenerate on ___ from produced docs
		Short:             "Fetch, deploy and debug Envoy",
		Long: `Manage full lifecycle of Envoy including fetching binaries,
bootstrap generation and automated collection of access logs, Envoy state and machine state.`,
		Version: version.Version,
	}

	manifestURL string
)
