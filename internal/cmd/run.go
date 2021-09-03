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
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/tetratelabs/func-e/internal/envoy"
	"github.com/tetratelabs/func-e/internal/envoy/shutdown"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/version"
)

// NewRunCmd create a command responsible for starting an Envoy process
func NewRunCmd(o *globals.GlobalOpts) *cli.Command {
	runDirectoryExpression := moreos.ReplacePathSeparator("$FUNC_E_HOME/runs/$epochtime")
	cmd := &cli.Command{
		Name:            "run",
		Usage:           "Run Envoy with the given [arguments...] until interrupted",
		ArgsUsage:       "[arguments...]",
		SkipFlagParsing: true,
		Description: moreos.Sprintf(`To run Envoy, execute ` + "`func-e run -c your_envoy_config.yaml`" + `.

The first version in the below is run, controllable by the "use" command:
` + fmt.Sprintf("```\n%s\n```", envoy.VersionUsageList()) + `
The version to use is downloaded and installed, if necessary.

Envoy interprets the '[arguments...]' and runs in the current working
directory (aka $PWD) until func-e is interrupted (ex Ctrl+C, Ctrl+Break).

Envoy's process ID and console output write to "envoy.pid", stdout.log" and
"stderr.log" in the run directory (` + fmt.Sprintf("`%s`", runDirectoryExpression) + `).
When interrupted, shutdown hooks write files including network and process
state. On exit, these archive into ` + fmt.Sprintf("`%s.tar.gz`", runDirectoryExpression)),
		Before: func(c *cli.Context) error {
			return ensureEnvoyVersion(c, o)
		},
		Action: func(c *cli.Context) error {
			if err := initializeRunOpts(c.Context, o, o.EnvoyVersion); err != nil {
				return err
			}
			r := envoy.NewRuntime(&o.RunOpts)

			stdoutLog, err := os.OpenFile(filepath.Join(r.GetRunDir(), "stdout.log"), os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				return fmt.Errorf("couldn't create stdout log file: %w", err)
			}
			defer stdoutLog.Close() //nolint
			r.OutFile = stdoutLog
			r.Out = io.MultiWriter(c.App.Writer, stdoutLog)

			stderrLog, err := os.OpenFile(filepath.Join(r.GetRunDir(), "stderr.log"), os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				return fmt.Errorf("couldn't create stderr log file: %w", err)
			}
			defer stderrLog.Close() //nolint
			r.ErrFile = stderrLog
			r.Err = io.MultiWriter(c.App.ErrWriter, stderrLog)

			for _, enableShutdownHook := range shutdown.EnableHooks {
				if err := enableShutdownHook(r); err != nil {
					moreos.Fprintf(r.Out, "failed to enable shutdown hook: %s\n", err) //nolint
				}
			}

			return r.Run(c.Context, c.Args().Slice())
		},
		CustomHelpTemplate: moreos.Sprintf(cli.CommandHelpTemplate),
	}
	return cmd
}

// initializeRunOpts allows us to default values when not overridden for tests.
// The version parameter correlates with the globals.GlobalOpts EnvoyPath which is installed if needed.
// Notably, this creates and sets a globals.GlobalOpts WorkingDirectory for Envoy, and any files that precede it.
func initializeRunOpts(ctx context.Context, o *globals.GlobalOpts, v version.Version) error {
	runOpts := &o.RunOpts
	if o.EnvoyPath == "" { // not overridden for tests
		envoyPath, err := envoy.InstallIfNeeded(ctx, o, v)
		if err != nil {
			return err
		}
		o.EnvoyPath = envoyPath
	}
	if runOpts.RunDir == "" { // not overridden for tests
		runID := strconv.FormatInt(time.Now().UnixNano(), 10)
		runDir := filepath.Join(filepath.Join(o.HomeDir, "runs"), runID)

		// Eagerly create the run dir, so that errors raise early
		if err := os.MkdirAll(runDir, 0750); err != nil {
			return NewValidationError("unable to create working directory %q, so we cannot run envoy", runDir)
		}
		runOpts.RunDir = runDir
	}
	return nil
}

// setHomeEnvoyVersion makes sure the $FUNC_E_HOME/version exists.
func setHomeEnvoyVersion(ctx context.Context, o *globals.GlobalOpts) error {
	v, homeVersionFile, err := envoy.GetHomeVersion(o.HomeDir)
	if err != nil {
		return NewValidationError(err.Error())
	} else if v != "" { // home version is already valid
		return nil
	}

	// First time install: look up the latest version, which may be newer than version.LastKnownEnvoy!
	o.Logf("looking up the latest Envoy version\n") //nolint
	m, err := o.FuncEVersions.Get(ctx)
	if err != nil {
		return NewValidationError(`couldn't read latest version from %s: %s`, o.EnvoyVersionsURL, err)
	}
	// Persist it for the next invocation
	return os.WriteFile(homeVersionFile, []byte(m.LatestVersion), 0600)
}

func ensureEnvoyVersion(c *cli.Context, o *globals.GlobalOpts) error {
	if err := os.MkdirAll(o.HomeDir, 0750); err != nil {
		return NewValidationError(err.Error())
	}

	if o.EnvoyVersion == "" { // not overridden for tests
		if err := setHomeEnvoyVersion(c.Context, o); err != nil {
			return err
		}
		v, _, err := envoy.CurrentVersion(o.HomeDir)
		if err != nil {
			return NewValidationError(err.Error())
		}
		o.EnvoyVersion = v
		if matched := globals.EnvoyStrictMinorVersionPattern.MatchString(string(v)); matched {
			var err error
			var latest version.Version
			if latest, err = o.FuncEVersions.FindLatestPatch(c.Context, v); err != nil {
				if latest, err = getLatestInstalledPatch(o, v); err != nil {
					return err
				}
				o.Logf("couldn't check the latest patch for %q for platform %q, using the latest installed version %q\n", v, o.Platform, latest)
			}
			o.EnvoyVersion = latest
		}
	}
	return nil
}
