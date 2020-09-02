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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	examples "github.com/tetratelabs/getenvoy/pkg/extension/workspace/example"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
)

// NewAddCmd returns a command that generates a new example setup.
func NewAddCmd() *cobra.Command {
	name := examples.Default
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Scaffold a new example setup.",
		Long: `
Scaffold a new example setup.`,
		Example: `
  # Scaffold the default example setup (named "default").
  getenvoy extension examples add

  # Scaffold an example setup with a given name.
  getenvoy extension examples add --name advanced`,
		Args: func(cmd *cobra.Command, args []string) error {
			return model.ValidateExampleName(name)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// find workspace
			workspace, err := workspaces.GetCurrentWorkspace()
			if err != nil {
				return err
			}
			// try to find existing example
			exists, err := workspace.HasExample(name)
			if err != nil {
				return err
			}
			// handle the case where example already exists
			if exists {
				return errors.Errorf("example setup %q already exists", name)
			}
			// generate new example
			scaffoldOpts := &examples.ScaffoldOpts{
				Workspace:    workspace,
				Name:         name,
				ProgressSink: NewAddExampleFeedback(cmd),
			}
			return examples.Scaffold(scaffoldOpts)
		},
	}
	cmd.PersistentFlags().StringVar(&name, "name", name, `Example name, e.g. "default", advanced", "grpc-web", etc`)
	return cmd
}
