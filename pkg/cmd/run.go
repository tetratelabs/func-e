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
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	defaultreference "github.com/tetratelabs/getenvoy/pkg"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/debug"
	"github.com/tetratelabs/getenvoy/pkg/globals"
)

// NewRunCmd create a command responsible for starting an Envoy process
func NewRunCmd(o *globals.GlobalOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run reference [flags] [-- <envoy-args>]",
		Short: "Runs Envoy and collects process state on exit. Available builds can be retrieved using `getenvoy list`.",
		Example: fmt.Sprintf(`# Run using a manifest reference.
getenvoy run %[1]s -- --config-path ./bootstrap.yaml

# List available Envoy flags.
getenvoy run %[1]s -- --help
`, defaultreference.Latest),
		Args: validateReferenceArg,
		RunE: func(c *cobra.Command, args []string) error {
			if err := initializeRunOpts(o, args[0]); err != nil {
				return err
			}
			r := envoy.NewRuntime(&o.RunOpts)
			r.Out = c.OutOrStderr()
			r.Err = c.ErrOrStderr()

			debug.EnableAll(r)

			envoyArgs := args[1:]
			return r.Run(c.Context(), envoyArgs)
		},
	}
	return cmd
}

// initializeRunOpts allows us to default values when not overridden for tests.
// The reference parameter corresponds to the globals.GlobalOpts EnvoyPath which is fetched if needed.
// Notably, this creates and sets a globals.GlobalOpts WorkingDirectory for Envoy, and any files that precede it.
func initializeRunOpts(o *globals.GlobalOpts, reference string) error {
	runOpts := &o.RunOpts
	if o.EnvoyPath == "" { // not overridden for tests
		envoyPath, err := envoy.FetchIfNeeded(o, reference)
		if err != nil {
			return newValidationError(err.Error())
		}
		o.EnvoyPath = envoyPath
	}
	if runOpts.WorkingDir == "" { // not overridden for tests
		// Historically, the directory run files wrote to was called DebugStore
		runID := strconv.FormatInt(time.Now().UnixNano(), 10)
		workingDir := filepath.Join(filepath.Join(o.HomeDir, "debug"), runID)

		// When the directory is implicitly generated, we should create it to avoid late errors.
		if err := os.MkdirAll(workingDir, 0750); err != nil {
			return newValidationError("unable to create working directory %q, so we cannot run envoy", workingDir)
		}
		runOpts.WorkingDir = workingDir
	}
	if runOpts.Log == nil { // not overridden for tests
		runOpts.Log = log.New(os.Stdout, "run: ", log.LstdFlags)
	}
	if runOpts.DebugLog == nil { // not overridden for tests
		// All debug features are optional. If there is any unexpected failure, log as "debug" to stdout.
		runOpts.DebugLog = log.New(os.Stdout, "debug: ", log.LstdFlags)
	}
	return nil
}

// Run enables debug and runs Envoy with the IO from the cobra.Command
// This is exposed for re-use in "getenvoy extension run"
func Run(o *globals.GlobalOpts, cmd *cobra.Command, args []string) error {
	r := envoy.NewRuntime(&o.RunOpts)
	r.Out = cmd.OutOrStdout()
	r.Err = cmd.ErrOrStderr()

	debug.EnableAll(r)

	return r.Run(cmd.Context(), args)
}
