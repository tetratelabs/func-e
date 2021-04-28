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
	"fmt"

	"github.com/containerd/containerd/reference/docker"
	"github.com/spf13/cobra"

	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/globals"
	"github.com/tetratelabs/getenvoy/pkg/cmd/extension/common"
	"github.com/tetratelabs/getenvoy/pkg/extension/wasmimage"
	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/example/runtime"
)

// cmdOpts represents configuration options of the `push` command.
type cmdOpts struct {
	// toolchain to use to build the *.wasm file.
	toolchain common.ToolchainOpts
	// extension to use to specify the built *.wasm file.
	extension runtime.ExtensionOpts

	allowInsecure bool
	plainHTTP     bool
}

func (opts *cmdOpts) GetToolchainName() string {
	return opts.toolchain.Name
}

func (opts *cmdOpts) ApplyTo(interface{}) {}

func (opts *cmdOpts) Validate() error {
	return opts.toolchain.Validate()
}

// NewCmd returns a command that pushes the built extension.
func NewCmd(o *globals.GlobalOpts) *cobra.Command {
	opts := &cmdOpts{
		toolchain:     common.NewToolchainOpts(o),
		extension:     runtime.ExtensionOpts{},
		allowInsecure: false,
		plainHTTP:     false,
	}

	cmd := &cobra.Command{
		Use:   "push <image-reference>",
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
			imagePath := opts.extension.WasmFile
			if imagePath == "" {
				ws, err := workspaces.GetWorkspaceAt(o.ExtensionDir)
				if err != nil {
					return err
				}
				tc, err := common.LoadToolchain(ws, opts)
				if err != nil {
					return err
				}
				imagePath = tc.GetBuildOutputWasmFile()
			}
			imageRef := args[0]
			ref, err := docker.ParseNormalizedNamed(imageRef)
			if err != nil {
				return fmt.Errorf("invalid image-reference: %w", err)
			}
			ref = docker.TagNameOnly(ref)
			if tagged, ok := ref.(docker.Tagged); ok {
				cmd.Printf("Using default tag: %s\n", tagged.Tag())
			}
			pusher, err := wasmimage.NewPusher(opts.allowInsecure, opts.plainHTTP)
			if err != nil {
				return fmt.Errorf("failed to push the wasm image: %w", err)
			}
			manifest, size, err := pusher.Push(imagePath, ref.String())
			if err != nil {
				return fmt.Errorf("failed to push the wasm image: %w", err)
			}

			cmd.Printf("Pushed %s\n", ref)
			cmd.Printf("digest: %s size: %d\n", manifest.Digest, size)

			return nil
		},
	}
	cmd.PersistentFlags().StringVar(&opts.toolchain.Name, "toolchain", opts.toolchain.Name,
		`Name of the toolchain to use, e.g. "default" toolchain that is backed by a Docker build container`)
	cmd.PersistentFlags().BoolVar(&opts.allowInsecure, "allow-insecure", opts.allowInsecure, `allow insecure TLS communication with registry`)
	cmd.PersistentFlags().BoolVar(&opts.plainHTTP, "use-http", opts.plainHTTP, `Use HTTP for communication with registry`)
	cmd.PersistentFlags().StringVar(&opts.extension.WasmFile, "extension-file", opts.extension.WasmFile,
		`Use a pre-built *.wasm file`)
	return cmd
}
