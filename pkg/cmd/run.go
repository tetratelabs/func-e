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
	"strings"

	"github.com/spf13/cobra"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/controlplane"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/debug"

	cmdutil "github.com/tetratelabs/getenvoy/pkg/util/cmd"
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
		Use:   "run <reference|filepath> [flags] [-- <envoy-args>]",
		Short: "Runs an instance of Envoy.",
		Long: `
Manages full lifecycle of Envoy including bootstrap generation and automated collection of access logs,
Envoy state and machine state into the ` + "`~/.getenvoy/debug`" + ` directory.`,
		Example: `# Run using a manifest reference.
getenvoy run standard:1.11.1 -- --config-path ./bootstrap.yaml

# Run as a gateway using an Istio controlplane bootstrap.
getenvoy run standard:1.11.1 --mode loadbalancer --bootstrap istio --controlplaneAddress istio-pilot.istio-system:15010

# Run using a filepath.
getenvoy run ./envoy -- --config-path ./bootstrap.yaml

# List available Envoy flags.
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
				func(r *envoy.Runtime) {
					r.Config = cfg
					r.IO = cmdutil.StreamsOf(cmd)
				},
				debug.EnableEnvoyAdminDataCollection,
				debug.EnableEnvoyLogCollection,
				debug.EnableNodeCollection,
				debug.EnableOpenFilesDataCollection,
				controlplaneFunc(),
			)
			if err != nil {
				return err
			}

			return runtime.FetchAndRun(args[0], args[1:])
		},
	}
	cmd.Flags().StringVarP(&bootstrap, "bootstrap", "b", "",
		fmt.Sprintf("(experimental) controlplane bootstrap to generate and use <%v>", strings.Join(supported, "|")))
	cmd.Flags().StringVar(&controlplaneAddress, "controlplaneAddress", "",
		"(experimental) location of Envoy's dynamic configuration server <host|ip:port> (requires bootstrap flag)")
	cmd.Flags().StringVar(&accessLogServerAddress, "accessLogServerAddress", "",
		"(experimental) location of Envoy's access log server <host|ip:port> (requires bootstrap flag)")
	cmd.Flags().StringVar(&mode, "mode", "",
		fmt.Sprintf("(experimental) mode to run Envoy in <%v> (requires bootstrap flag)", strings.Join(envoy.SupportedModes, "|")))
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
