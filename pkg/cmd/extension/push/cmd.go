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

package push

import (
	"errors"
	"github.com/spf13/cobra"
	"github.com/tetratelabs/getenvoy/pkg/cmd/extension/common"
	"github.com/tetratelabs/getenvoy/pkg/extension/wasmimage"
	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/example/runtime"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain"

	//cmdutil "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

// cmdOpts represents configuration options of the `push` command.
type cmdOpts struct {
	// Toolchain to use to build the *.wasm file.
	Toolchain common.ToolchainOpts
	// Extension to use to specify the built *.wasm file.
	Extension runtime.ExtensionOpts
	// Pusher to use to specify options for Pusher
	Pusher wasmimage.PusherOpts
}

func newCmdOpts() *cmdOpts {
	return &cmdOpts{
		Toolchain: common.ToolchainOpts{
			Name: toolchain.Default,
		},
		Extension: runtime.ExtensionOpts{},
		Pusher: wasmimage.NewPusherOpts(),
	}
}

func (opts *cmdOpts) GetToolchainName() string {
	return opts.Toolchain.Name
}

func (opts *cmdOpts) ApplyTo(interface{}) {}

func (opts *cmdOpts) Validate() error {
	if err := opts.Toolchain.Validate(); err != nil {
		return err
	}

	return nil
}

// NewCmd returns a command that pushes the built extension.
func NewCmd() *cobra.Command {
	opts := newCmdOpts()
    cmd := &cobra.Command{
        Use: "push <image-reference>",
        Short: "Push the built WASM extension to the OCI-compliant registry.",
        Long: `
Push the built WASM extension to the OCI-compliant registry. This command requires to login the target container registry with docker CLI`,
        Example: `
  # Push built WASM extension to the local docker registry.
  getenvoy extension push localhost:5000/test/image-name:tag`,
		Args: func(cmd *cobra.Command, args []string) error {
            if err := opts.Validate(); err != nil {
            	return err
			}

			if len(args) == 0 {
				return errors.New("missing image-reference parameter")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			imageRef := args[0]
			workspace, err := workspaces.GetCurrentWorkspace()
			if err != nil {
				return err
			}
			toolchain, err := common.LoadToolchain(workspace, opts)
			if err != nil {
				return err
			}
			var image *wasmimage.WasmImage
			if opts.Extension.WasmFile != "" {
				image, err = wasmimage.NewWasmImage(imageRef, opts.Extension.WasmFile)
			} else {
				image, err = toolchain.Package(imageRef)
			}
			pusher, err := wasmimage.NewPusher(false, false)
			_, err = pusher.Push(image)
			return err
		},
    }
	cmd.PersistentFlags().StringVar(&opts.Toolchain.Name, "toolchain", opts.Toolchain.Name,
		`Name of the toolchain to use, e.g. "default" toolchain that is backed by a Docker build container`)
	cmd.PersistentFlags().BoolVar(&opts.Pusher.AllowInsecure, "allow-insecure", opts.Pusher.AllowInsecure, `Allow insecure registry`)
    cmd.PersistentFlags().BoolVar(&opts.Pusher.UseHTTP, "use-http", opts.Pusher.UseHTTP, `Use HTTP for communication with registry`)
	cmd.PersistentFlags().StringVar(&opts.Extension.WasmFile, "extension-file", opts.Extension.WasmFile,
		`Use a pre-built *.wasm file`)
	cmd.PersistentFlags().StringVar(&opts.Extension.Config.Source, "extension-config-file", opts.Extension.Config.Source,
		`Use a custom extension config`)
    return cmd
}
