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
	"strconv"
	"time"

	"github.com/spf13/cobra"

	defaultreference "github.com/tetratelabs/getenvoy/pkg"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/debug"
	"github.com/tetratelabs/getenvoy/pkg/flavors"
	_ "github.com/tetratelabs/getenvoy/pkg/flavors/postgres" //nolint
	"github.com/tetratelabs/getenvoy/pkg/globals"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
	ioutil "github.com/tetratelabs/getenvoy/pkg/util/io"
)

// NewRunCmd create a command responsible for starting an Envoy process
func NewRunCmd(o *globals.GlobalOpts) *cobra.Command {
	templateArgs := map[string]string{}

	cmd := &cobra.Command{
		Use:   "run reference [flags] [-- <envoy-args>]",
		Short: "Runs an instance of Envoy.",
		Long: `
Manages full lifecycle of Envoy including bootstrap generation and automated collection of access logs,
Envoy state and machine state into the ` + "`~/.getenvoy/debug`" + ` directory.`,
		Example: fmt.Sprintf(`# Run using a manifest reference.
getenvoy run %[1]s -- --config-path ./bootstrap.yaml

# Run using a filepath.
getenvoy run /usr/local/bin/envoy -- --config-path ./bootstrap.yaml

# List available Envoy flags.
getenvoy run %[1]s -- --help

# Run with Postgres specific configuration bootstrapped
getenvoy run postgres:nightly --templateArg endpoints=127.0.0.1:5432,192.168.0.101:5432 --templateArg inport=5555
`, defaultreference.Latest),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := InitializeRunOpts(o, args[0]); err != nil {
				return err
			}

			// Check if the templateArgs were passed to the cmd line. If they were passed, config must be created based
			// on template.
			// TODO: delete this template thing as it is really complicated and is less functional than shell scripts.
			if len(templateArgs) > 0 {
				key, err := manifest.NewKey(args[0])
				if err != nil {
					return fmt.Errorf("envoy version is not valid: %w", err)
				}
				configYaml, err := flavors.CreateConfig(key.Flavor, templateArgs)
				if err != nil {
					return err
				}
				args = append(args, `--config-yaml`, configYaml)
			}

			return Run(o, cmd, args[1:]) // consume the envoy path argument
		},
	}
	cmd.Flags().StringToStringVar(&templateArgs, "templateArg", map[string]string{},
		"arguments passed to a config template for substitution")
	return cmd
}

// InitializeRunOpts allows us to default values when not overridden for tests.
// The reference parameter corresponds to the globals.GlobalOpts EnvoyPath which is fetched if needed.
// Notably, this creates and sets a globals.GlobalOpts WorkingDirectory for Envoy, and any files that precede it.
func InitializeRunOpts(o *globals.GlobalOpts, reference string) error {
	runOpts := &o.RunOpts
	if o.EnvoyPath == "" { // not overridden for tests
		envoyPath, err := envoy.FetchIfNeeded(o, reference)
		if err != nil {
			return err
		}
		o.EnvoyPath = envoyPath
	}
	if runOpts.WorkingDir == "" { // not overridden for tests
		// Historically, the directory run files wrote to was called DebugStore
		runID := strconv.FormatInt(time.Now().UnixNano(), 10)
		workingDir := filepath.Join(filepath.Join(o.HomeDir, "debug"), runID)

		// When the directory is implicitly generated, we should create it to avoid late errors.
		if err := os.MkdirAll(workingDir, 0750); err != nil {
			return fmt.Errorf("unable to create working directory %q, so we cannot run envoy: %w", workingDir, err)
		}
		runOpts.WorkingDir = workingDir
	}
	return nil
}

// Run enables debug and runs Envoy with the IO from the cobra.Command
// This is exposed for re-use in "getenvoy extension run"
func Run(o *globals.GlobalOpts, cmd *cobra.Command, args []string) error {
	r := envoy.NewRuntime(&o.RunOpts)
	r.IO = ioutil.StdStreams{
		In:  cmd.InOrStdin(),
		Out: cmd.OutOrStdout(),
		Err: cmd.ErrOrStderr(),
	}
	debug.EnableAll(r)
	return r.Run(cmd.Context(), args)
}
