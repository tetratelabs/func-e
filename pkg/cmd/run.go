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
	"strings"

	"github.com/spf13/cobra"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/controlplane"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/debug"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
)

var (
	controlplaneAddress    string
	accessLogServerAddress string
	mode                   string
	bootstrap              string
)

// NewRunCmd create a command responsible for starting an Envoy process
func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [manifest-reference|filepath] -- <envoy-args>",
		Short: "Starts an Envoy process using the reference or path passed.",
		Long: `
Starts an Envoy process using the location passed. 
Location can be a manifest reference or path to an Envoy binary.`,
		Example: `# Run using a manifest reference. Reference format is <flavor>:<version>.
getenvoy run standard:1.11.1 -- --config-path ./bootstrap.yaml

# Run using a filepath
getenvoy run ./envoy -- --config-path ./bootstrap.yaml

# List available Envoy flags
getenvoy run standard:1.11.1 -- --help
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing binary parameter")
			}
			if err := validateMode(); err != nil {
				return err
			}
			if err := validateBootstrap(); err != nil {
				return err
			}
			return validateRequiresBootstrap()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := envoy.NewConfig(
				func(c *envoy.Config) {
					c.XDSAddress = controlplaneAddress
					c.Mode = envoy.ParseMode(mode)
					c.ALSAddresss = accessLogServerAddress
				},
			)

			runtime, err := envoy.NewRuntime(
				func(r *envoy.Runtime) { r.Config = cfg },
				debug.EnableEnvoyAdminDataCollection,
				debug.EnableEnvoyLogCollection,
				debug.EnableNodeCollection,
				controlplaneFunc(),
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
	cmd.Flags().StringVarP(&bootstrap, "bootstrap", "b", "",
		fmt.Sprintf("which controlplane's bootstrap to generate and use (%v) [experimental]", strings.Join(supported, "|")))
	cmd.Flags().StringVar(&controlplaneAddress, "controlplaneAddress", "",
		"location of Envoy's dynamic configuration server (<host|ip>:port) [requires bootstrap to be set]")
	cmd.Flags().StringVar(&accessLogServerAddress, "accessLogServerAddress", "",
		"location of Envoy's access log server(<host|ip>:port) [requires bootstrap to be set]")
	cmd.Flags().StringVarP(&mode, "mode", "m", "",
		fmt.Sprintf("mode to run Envoy in (%v) [requires bootstrap to be set]", strings.Join(envoy.SupportedModes, "|")))
	return cmd
}

var (
	istio = "istio"

	supported         = []string{istio}
	requiresBootstrap = []*string{&controlplaneAddress, &accessLogServerAddress, &mode}
)

func validateBootstrap() error {
	for _, bs := range append(supported, "") {
		if bs == bootstrap {
			return nil
		}
	}
	return fmt.Errorf("unsupported bootstrap %v, must be one of (%v)", bootstrap, strings.Join(supported, "|"))
}

func validateMode() error {
	for _, m := range append(envoy.SupportedModes, "") {
		if m == mode {
			return nil
		}
	}
	return fmt.Errorf("unsupported mode %v, must be one of (%v)", mode, strings.Join(envoy.SupportedModes, "|"))
}

func validateRequiresBootstrap() error {
	if bootstrap == "" {
		for i := range requiresBootstrap {
			if *requiresBootstrap[i] != "" {
				return fmt.Errorf("--%v requires --bootstrap to be set", *requiresBootstrap[i])
			}
		}
	}
	return nil
}

func controlplaneFunc() func(r *envoy.Runtime) {
	switch bootstrap {
	case istio:
		return controlplane.Istio
	default:
		// do nothing...
		return func(r *envoy.Runtime) {}
	}
}
