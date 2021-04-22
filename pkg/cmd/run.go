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
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	reference "github.com/tetratelabs/getenvoy/pkg"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/debug"
	"github.com/tetratelabs/getenvoy/pkg/flavors"
	_ "github.com/tetratelabs/getenvoy/pkg/flavors/postgres" //nolint
	"github.com/tetratelabs/getenvoy/pkg/manifest"
	cmdutil "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

var (
	templateArgs map[string]string
)

// NewRunCmd create a command responsible for starting an Envoy process
func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <reference> [flags] [-- <envoy-args>]",
		Short: "Runs an instance of Envoy.",
		Long: `
Manages full lifecycle of Envoy including bootstrap generation and automated collection of access logs,
Envoy state and machine state into the ` + "`~/.getenvoy/debug`" + ` directory.`,
		Example: fmt.Sprintf(`# Run using a manifest reference.
getenvoy run %[1]s -- --config-path ./bootstrap.yaml

# List available Envoy flags.
getenvoy run %[1]s -- --help

# Run with Postgres specific configuration bootstrapped
getenvoy run postgres:nightly --templateArg endpoints=127.0.0.1:5432,192.168.0.101:5432 --templateArg inport=5555
`, reference.Latest),
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := envoy.NewRuntime(envoy.RuntimeOption(
				func(r *envoy.Runtime) {
					r.IO = cmdutil.StreamsOf(cmd)
				}).
				AndAll(debug.EnableAll())...,
			)
			if err != nil {
				return err
			}

			key, manifestErr := manifest.NewKey(args[0])

			// Check if the templateArgs were passed to the cmd line.
			// If they were passed, config must be created based on
			// template.
			if manifestErr == nil && len(templateArgs) > 0 {
				cmdArg, err := processTemplateArgs(key.Flavor, templateArgs, runtime.(*envoy.Runtime))
				if err != nil {
					return err
				}
				args = append(args, cmdArg)
			}

			return runtime.FetchAndRun(args[0], args[1:])
		},
	}
	cmd.Flags().StringToStringVar(&templateArgs, "templateArg", map[string]string{},
		"arguments passed to a config template for substitution")
	return cmd
}

// Function creates config file based on template args passed by a user.
// The return value is Envoy command line option which must be passed to Envoy.
func processTemplateArgs(flavor string, templateArgs map[string]string, runtime *envoy.Runtime) (string, error) {
	config, err := flavors.CreateConfig(flavor, templateArgs)
	if err != nil {
		return "", err
	}
	// Save config in getenvoy directory
	path, err := saveConfig(flavor, config, filepath.Join(runtime.RootDir, "configs"))
	if err != nil {
		return "", err
	}
	return "--config-path " + path, nil
}

// SaveConfig saves configuration string in getenvoy directory.
func saveConfig(name, config, configDir string) (string, error) {
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return "", fmt.Errorf("unable to create directory %q: %w", configDir, err)
	}
	filename := name + ".yaml"
	err := os.WriteFile(filepath.Join(configDir, filename), []byte(config), 0600)
	if err != nil {
		return "", fmt.Errorf("cannot save config file %s: %w", filepath.Join(configDir, filename), err)
	}
	return filepath.Join(configDir, filename), nil
}
