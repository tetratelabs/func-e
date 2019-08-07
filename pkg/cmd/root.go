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
)

// NewRoot create a new root command and sets the cliVersion to the passed variable
// TODO: Add version support on the command
func NewRoot() *cobra.Command {
	runCmd := NewRunCmd()
	rootCmd.AddCommand(runCmd)

	listCmd := NewListCmd()
	rootCmd.AddCommand(listCmd)

	fetchCmd := NewFetchCmd()
	rootCmd.AddCommand(fetchCmd)

	rootCmd.PersistentFlags().StringVarP(&manifestURL, "url", "u",
		manifest.DefaultURL, "sets the manifest URL")
	return rootCmd
}

var (
	rootCmd = &cobra.Command{
		Use:   "getenvoy",
		Short: "getenvoy",
		Long:  "getenvoy",
	}

	manifestURL string
)
