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

	"github.com/urfave/cli/v2"

	"github.com/tetratelabs/getenvoy/internal/envoy"
	"github.com/tetratelabs/getenvoy/internal/envoy/debug"
	"github.com/tetratelabs/getenvoy/internal/globals"
)

// NewRunCmd create a command responsible for starting an Envoy process
func NewRunCmd(o *globals.GlobalOpts) *cli.Command {
	var envoyVersion string
	cmd := &cli.Command{
		Name:            "run",
		Usage:           "Run Envoy with the given [arguments...], collecting process state on termination",
		ArgsUsage:       "[arguments...]",
		SkipFlagParsing: true,
		Description: fmt.Sprintf(`The version of Envoy run is chosen from %s
The '[arguments...]' are interpreted by Envoy.

Envoy uses $GETENVOY_HOME/runs/$epochtime as the working directory.
Upon termination, this is archived as $GETENVOY_HOME/runs/$epochtime.tar.gz.

Example:
$ getenvoy run -c ./bootstrap.yaml`, envoy.VersionUsageList()),
		Before: func(context *cli.Context) error {
			if err := os.MkdirAll(o.HomeDir, 0750); err != nil {
				return NewValidationError(err.Error())
			}

			if o.EnvoyVersion == "" { // not overridden for tests
				if err := setHomeEnvoyVersion(o); err != nil {
					return err
				}
				v, _, err := envoy.CurrentVersion(o.HomeDir)
				if err != nil {
					return NewValidationError(err.Error())
				}
				o.EnvoyVersion = v
			}
			envoyVersion = o.EnvoyVersion
			return nil
		},
		Action: func(c *cli.Context) error {
			if err := initializeRunOpts(o, globals.CurrentPlatform, envoyVersion); err != nil {
				return err
			}
			r := envoy.NewRuntime(&o.RunOpts)

			r.Out = c.App.Writer
			r.Err = c.App.ErrWriter

			for _, err := range debug.EnableAll(r) {
				fmt.Fprintln(r.Out, "failed to enable debug option:", err) //nolint
			}

			return r.Run(c.Context, c.Args().Slice())
		},
	}
	return cmd
}

// initializeRunOpts allows us to default values when not overridden for tests.
// The version parameter correlates with the globals.GlobalOpts EnvoyPath which is installed if needed.
// Notably, this creates and sets a globals.GlobalOpts WorkingDirectory for Envoy, and any files that precede it.
func initializeRunOpts(o *globals.GlobalOpts, platform, version string) error {
	runOpts := &o.RunOpts
	if o.EnvoyPath == "" { // not overridden for tests
		envoyPath, err := envoy.InstallIfNeeded(o, platform, version)
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
			return NewValidationError("unable to create working directory %q, so we cannot run envoy", workingDir)
		}
		runOpts.WorkingDir = workingDir
	}
	return nil
}

// setHomeEnvoyVersion makes sure the $GETENVOY_HOME/version exists.
func setHomeEnvoyVersion(o *globals.GlobalOpts) error {
	v, homeVersionFile, err := envoy.GetHomeVersion(o.HomeDir)
	if err != nil {
		return NewValidationError(err.Error())
	} else if v != "" { // home version is already valid
		return nil
	}

	// First time install: look up the latest version, which may be newer than version.LastKnownEnvoy!
	fmt.Fprintln(o.Out, "looking up latest version") //nolint
	m, err := envoy.GetEnvoyVersions(o.EnvoyVersionsURL, o.UserAgent)
	if err != nil {
		return NewValidationError(`couldn't read latest version from %s: %s`, o.EnvoyVersionsURL, err)
	}
	// Persist it for the next invocation
	return os.WriteFile(homeVersionFile, []byte(m.LatestVersion), 0600)
}
