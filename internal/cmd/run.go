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
	"runtime"
	"strconv"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/tetratelabs/getenvoy/internal/binary/envoy"
	"github.com/tetratelabs/getenvoy/internal/binary/envoy/debug"
	"github.com/tetratelabs/getenvoy/internal/errors"
	"github.com/tetratelabs/getenvoy/internal/globals"
	latestversion "github.com/tetratelabs/getenvoy/internal/version"
)

// NewRunCmd create a command responsible for starting an Envoy process
func NewRunCmd(o *globals.GlobalOpts) *cli.Command {
	cmd := &cli.Command{
		Name:      "run",
		Usage:     "Run Envoy as <version> with <args> as arguments, collecting process state on termination",
		ArgsUsage: "<version> <args>",
		Description: fmt.Sprintf(`The '<version>' is from the "versions" command and installed if necessary.
The '<args>' are interpreted by Envoy.
The Envoy working directory is archived as $GETENVOY_HOME/runs/$epochtime.tar.gz upon termination.

Example:
$ getenvoy run %s --config-path ./bootstrap.yaml`, latestversion.Envoy),
		Before: validateVersionArg,
		Action: func(c *cli.Context) error {
			args := c.Args().Slice()
			if err := initializeRunOpts(o, runtime.GOOS, args[0]); err != nil {
				return err
			}
			r := envoy.NewRuntime(&o.RunOpts)

			r.Out = c.App.Writer
			r.Err = c.App.ErrWriter

			for _, err := range debug.EnableAll(r) {
				fmt.Fprintln(r.Out, "failed to enable debug option:", err) //nolint
			}

			return r.Run(c.Context, args[1:])
		},
	}
	return cmd
}

// initializeRunOpts allows us to default values when not overridden for tests.
// The version parameter correlates with the globals.GlobalOpts EnvoyPath which is installed if needed.
// Notably, this creates and sets a globals.GlobalOpts WorkingDirectory for Envoy, and any files that precede it.
func initializeRunOpts(o *globals.GlobalOpts, goos, version string) error {
	runOpts := &o.RunOpts
	if o.EnvoyPath == "" { // not overridden for tests
		envoyPath, err := envoy.InstallIfNeeded(o, goos, version)
		if err != nil {
			return err
		}
		o.EnvoyPath = envoyPath
	}
	if runOpts.WorkingDir == "" { // not overridden for tests
		// Historically, the directory run files wrote to was called DebugStore
		runID := strconv.FormatInt(time.Now().UnixNano(), 10)
		workingDir := filepath.Join(filepath.Join(o.HomeDir, "runs"), runID)

		// When the directory is implicitly generated, we should create it to avoid late errors.
		if err := os.MkdirAll(workingDir, 0750); err != nil {
			return errors.NewValidationError("unable to create working directory %q, so we cannot run envoy", workingDir)
		}
		runOpts.WorkingDir = workingDir
	}
	return nil
}
