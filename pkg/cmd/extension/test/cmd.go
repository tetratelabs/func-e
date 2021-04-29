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

package test // nolint:dupl

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tetratelabs/getenvoy/pkg/cmd/extension/common"
	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	builtinconfig "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/toolchain/builtin"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain/types"
	"github.com/tetratelabs/getenvoy/pkg/globals"
	cmdutil "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

// cmdOpts represents configuration options of the `test` command.
type cmdOpts struct {
	Toolchain common.ToolchainOpts
}

func (opts *cmdOpts) GetToolchainName() string {
	return opts.Toolchain.Name
}

// ApplyTo applies toolchain-related command options to a given toolchain config.
func (opts *cmdOpts) ApplyTo(config interface{}) {
	if c, ok := config.(*builtinconfig.ToolchainConfig); ok {
		opts.Toolchain.Builtin.ApplyTo(c.GetTestContainer())
	}
}

// NewCmd returns a command that unit tests the extension.
func NewCmd(o *globals.GlobalOpts) *cobra.Command {
	opts := &cmdOpts{Toolchain: common.NewToolchainOpts(o)}

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Unit test Envoy extension.",
		Long: `
Run unit tests on Envoy extension.`,
		Example: `
  # Run unit tests on Envoy extension using default toolchain (Docker build container)
  getenvoy extension test

  # Run unit tests on Envoy extension using Docker build container with extra options
  getenvoy extension test --toolchain-container-options '-e VAR=VALUE -v /host/path:/container/path'

  # Run unit tests on Envoy extension using Docker build container with SSH agent forwarding enabled (Docker for Mac)
  getenvoy extension test --toolchain-container-options ` +
			`'--mount type=bind,src=/run/host-services/ssh-auth.sock,target=/run/host-services/ssh-auth.sock ` +
			`-e SSH_AUTH_SOCK=/run/host-services/ssh-auth.sock'`,
		Args: func(*cobra.Command, []string) error {
			return opts.Toolchain.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := workspaces.GetWorkspaceAt(o.ExtensionDir)
			if err != nil {
				return err
			}
			toolchain, err := common.LoadToolchain(workspace, opts)
			if err != nil {
				return err
			}
			err = toolchain.Test(types.TestContext{
				IO: cmdutil.StreamsOf(cmd),
			})
			if err != nil {
				return fmt.Errorf("failed to unit test Envoy extension using %q toolchain: %w", opts.Toolchain.Name, err)
			}
			return nil
		},
	}
	common.AddToolchainFlags(cmd, &opts.Toolchain)
	return cmd
}
