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
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/debug"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
)

// NewRunCmd create a command responsible for starting an Envoy process
func NewRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run [manifest-reference|filepath] -- <envoy-args>",
		Short: "Starts an Envoy process using the reference or path passed.",
		Long: `
Starts an Envoy process using the location passed. 
Location can be a manifest reference or path to an Envoy binary.`,
		Example: `# Run using a manifest reference. Reference format is <flavor>:<version>.
getenvoy run standard:1.10.1 -- --config-path ./bootstrap.yaml

# Run using a filepath .
getenvoy run ./envoy -- --config-path ./bootstrap.yaml

# List available Envoy flags
getenvoy run standard:1.10.1 -- --help
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing binary parameter")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := envoy.NewRuntime(
				debug.EnableEnvoyAdminDataCollection,
			)
			if err != nil {
				return err
			}
			key, manifestErr := manifest.NewKey(args[0])
			if manifestErr != nil {
				if _, err := os.Stat(args[0]); err != nil {
					return fmt.Errorf("%v isn't valid manifest reference or an existing filepath", args[0])
				}
				return runtime.RunPath(args[0], args[1:])
			}
			if !runtime.AlreadyDownloaded(key) {
				location, err := manifest.Locate(key, manifestURL)
				if err != nil {
					return err
				}
				if err := runtime.Fetch(key, location); err != nil {
					return err
				}
			}
			return runtime.Run(key, args[1:])
		},
	}
}
