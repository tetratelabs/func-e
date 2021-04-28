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
	"github.com/spf13/cobra"

	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/globals"
)

// NewCmd returns a command that manages example setups.
func NewCmd(o *globals.GlobalOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "examples",
		Short: "Manage example setups.",
		Long: `
Manage example setups that demo the extension in action.`,
	}
	cmd.AddCommand(NewListCmd(o))
	cmd.AddCommand(NewAddCmd(o))
	cmd.AddCommand(NewRemoveCmd(o))
	return cmd
}
