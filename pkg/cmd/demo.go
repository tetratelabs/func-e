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
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/debug"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/demo"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
)

// NewDemoCmd create a command responsible for starting an Envoy process
func NewDemoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "demo [-- <envoy-args>]",
		Short: "Runs an instance of Envoy with a demo front proxy bootstrap.",
		Long: `
Manages full lifecycle of Envoy including bootstrap generation and automated collection of access logs,
Envoy state and machine state into the ` + "`~/.getenvoy/debug`" + ` directory.`,
		Example: `# Run using standard Envoy.
getenvoy demo
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := envoy.NewRuntime(
				debug.EnableEnvoyAdminDataCollection,
				debug.EnableEnvoyLogCollection,
				debug.EnableNodeCollection,
				bootstrapFunc(args),
			)
			if err != nil {
				return err
			}
			// TODO: add standard:latest functionality
			key, _ := manifest.NewKey("standard:1.11.1")
			if !runtime.AlreadyDownloaded(key) {
				location, err := manifest.Locate(key, manifestURL)
				if err != nil {
					return err
				}
				if err := runtime.Fetch(key, location); err != nil {
					return err
				}
			}
			fmt.Println(demo.Instructions) // Intentionally println and not log to make it more readable
			return runtime.Run(key, args)
		},
	}
	return cmd
}

func bootstrapFunc(args []string) func(r *envoy.Runtime) {
	if !demo.HasBootstrapArg(args) {
		return demo.StaticBootstrap
	}
	return func(r *envoy.Runtime) {}
}
