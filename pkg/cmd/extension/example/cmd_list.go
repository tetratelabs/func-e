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

package example

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/globals"
	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
)

// NewListCmd returns a command that prints a list of existing example setups.
func NewListCmd(o *globals.GlobalOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List existing example setups.",
		Long: `
List existing example setups.`,
		Example: `
  # List existing example setups.
  getenvoy extension examples list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// find workspace
			workspace, err := workspaces.GetWorkspaceAt(o.ExtensionDir)
			if err != nil {
				return err
			}
			// list examples
			examples, err := workspace.ListExamples()
			if err != nil {
				return err
			}
			// handle empty list
			if len(examples) == 0 {
				fmt.Fprintf(cmd.ErrOrStderr(), `Extension has no example setups.

Use "getenvoy extension examples add --help" for more information on how to add one.
`)
				return nil
			}
			// print a list
			table := tabwriter.NewWriter(cmd.OutOrStdout(), 1, 0, 3, ' ', 0)
			defer table.Flush() //nolint
			fmt.Fprintf(table, "EXAMPLE\n")
			for _, example := range examples {
				fmt.Fprintf(table, "%s\n", example)
			}
			return nil
		},
	}
	return cmd
}
