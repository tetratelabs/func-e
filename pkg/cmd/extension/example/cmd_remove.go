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

	"github.com/spf13/cobra"

	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
	"github.com/tetratelabs/getenvoy/pkg/globals"
	uiutil "github.com/tetratelabs/getenvoy/pkg/util/ui"
)

// NewRemoveCmd returns a command that removes an existing example setup.
func NewRemoveCmd(o *globals.GlobalOpts) *cobra.Command {
	name := ""
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove example setup.",
		Long: `
Remove example setup.`,
		Example: `
  # Remove example setup by name.
  getenvoy extension examples remove --name advanced`,
		Args: func(cmd *cobra.Command, args []string) error {
			return model.ValidateExampleName(name)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// find workspace
			workspace, err := workspaces.GetWorkspaceAt(o.ExtensionDir)
			if err != nil {
				return err
			}
			// try to find existing example
			exists, err := workspace.HasExample(name)
			if err != nil {
				return err
			}
			// handle the case where example doesn't exists
			if !exists {
				fmt.Fprintf(cmd.ErrOrStderr(), `There is no example setup named %q.

Use "getenvoy extension examples list" to list existing example setups.
`, name)
				return nil
			}
			// remove the example
			progressSink := NewRemoveExampleFeedback(uiutil.NewStyleFuncs(o.NoColors), cmd.ErrOrStderr())
			return workspace.RemoveExample(name, model.ProgressSink{ProgressSink: progressSink})
		},
	}
	cmd.PersistentFlags().StringVar(&name, "name", name, `Example name, e.g. "default", "advanced", "grpc-web", etc`)
	return cmd
}
