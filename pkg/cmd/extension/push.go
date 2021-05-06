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

package extension

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/containerd/containerd/reference/docker"
	"github.com/spf13/cobra"

	wasmimage2 "github.com/tetratelabs/getenvoy/pkg/wasm"
)

// cmdOpts represents configuration options of the `push` command.
type cmdOpts struct {
	wasmFile string
	imageRef docker.Named

	allowInsecure, plainHTTP bool
}

func validateWasmFile(wasmFile string, opts *cmdOpts) error {
	if wasmFile == "" {
		return fmt.Errorf("WASM file empty")
	}
	info, err := os.Stat(wasmFile)
	if err != nil {
		return fmt.Errorf("invalid WASM file %q: %w", wasmFile, err)
	}
	if info.IsDir() {
		return fmt.Errorf("WASM file argument was a directory %q", wasmFile)
	}
	// Use absolute path because Envoy doesn't run in this directory and may be configured to hot-reload the VM.
	absoluteWasmFile, err := filepath.Abs(wasmFile)
	if err != nil {
		return err
	}
	opts.wasmFile = absoluteWasmFile
	return nil
}

func validateImageReference(imageRef string, opts *cmdOpts) error {
	if imageRef == "" {
		return fmt.Errorf("image reference empty")
	}
	ref, err := docker.ParseNormalizedNamed(imageRef)
	if err != nil {
		return fmt.Errorf("invalid image reference: %w", err)
	}
	opts.imageRef = ref
	return nil
}

// NewPushCmd returns a command that pushes the given extension.wasm file
func NewPushCmd() *cobra.Command {
	opts := &cmdOpts{}

	cmd := &cobra.Command{
		Use:   "push /path/to/extension.wasm <image-reference>",
		Short: "Push the WASM extension to the OCI-compliant registry.",
		Long: `
Push the WASM extension to the OCI-compliant registry. This command requires to login the target container registry with docker CLI`,
		Example: `
  # Push WASM extension to the local docker registry.
  getenvoy extension push /path/to/extension.wasm localhost:5000/test/image-name:tag`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.New("expected 2 arguments")
			}
			if err := validateWasmFile(args[0], opts); err != nil {
				return err
			}
			if err := validateImageReference(args[1], opts); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := docker.TagNameOnly(opts.imageRef)
			if tagged, ok := ref.(docker.Tagged); ok {
				cmd.Printf("Using default tag: %s\n", tagged.Tag())
			}
			pusher, err := wasmimage2.NewPusher(opts.allowInsecure, opts.plainHTTP)
			if err != nil {
				return fmt.Errorf("failed to push the wasm image: %w", err)
			}
			manifest, size, err := pusher.Push(opts.wasmFile, ref.String())
			if err != nil {
				return fmt.Errorf("failed to push the wasm image: %w", err)
			}

			cmd.Printf("Pushed %s\n", ref)
			cmd.Printf("digest: %s size: %d\n", manifest.Digest, size)

			return nil
		},
	}

	cmd.PersistentFlags().BoolVar(&opts.allowInsecure, "allow-insecure", opts.allowInsecure, `allow insecure TLS communication with registry`)
	cmd.PersistentFlags().BoolVar(&opts.plainHTTP, "use-http", opts.plainHTTP, `Use HTTP for communication with registry`)
	return cmd
}
