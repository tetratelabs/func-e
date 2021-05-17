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

	internalreference "github.com/tetratelabs/getenvoy/internal/reference"
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
`, internalreference.Latest),
		Args: validateReferenceArg,
		RunE: func(c *cobra.Command, args []string) error {
			if err := initializeRunOpts(o, args[0]); err != nil {
				return err
			}
			r := envoy.NewRuntime(&o.RunOpts)
			r.Out = c.OutOrStderr()
			r.Err = c.ErrOrStderr()

			for _, err := range debug.EnableAll(r) {
				fmt.Fprintln(r.Out, "failed to enable debug option:", err) //nolint
			}

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
	return nil
}
