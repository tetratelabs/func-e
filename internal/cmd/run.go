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
			if err := initializeRunOpts(c.Context, o); err != nil {
				return err
			}
			r := envoy.NewRuntime(&o.RunOpts)

			stdoutLog, err := os.OpenFile(filepath.Join(r.GetRunDir(), "stdout.log"), os.O_CREATE|os.O_WRONLY, 0o600)
			if err != nil {
				return fmt.Errorf("couldn't create stdout log file: %w", err)
			}
			defer stdoutLog.Close() //nolint
			r.OutFile = stdoutLog
			r.Out = io.MultiWriter(c.App.Writer, stdoutLog)

			stderrLog, err := os.OpenFile(filepath.Join(r.GetRunDir(), "stderr.log"), os.O_CREATE|os.O_WRONLY, 0o600)
			if err != nil {
				return fmt.Errorf("couldn't create stderr log file: %w", err)
			}
			defer stderrLog.Close() //nolint
			r.ErrFile = stderrLog
			r.Err = io.MultiWriter(c.App.ErrWriter, stderrLog)

			for _, enableShutdownHook := range shutdown.EnableHooks {
				if err := enableShutdownHook(r); err != nil {
					_, _ = moreos.Fprintf(r.Out, "failed to enable shutdown hook: %s\n", err)
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
func initializeRunOpts(ctx context.Context, o *globals.GlobalOpts) error {
	runOpts := &o.RunOpts
	if o.EnvoyPath == "" { // not overridden for tests
		envoyPath, err := envoy.InstallIfNeeded(ctx, o)
		if err != nil {
			return err
		}
		o.EnvoyPath = envoyPath
	}
	if runOpts.RunDir == "" { // not overridden for tests
		runID := strconv.FormatInt(time.Now().UnixNano(), 10)
		runDir := filepath.Join(filepath.Join(o.HomeDir, "runs"), runID)

		// Eagerly create the run dir, so that errors raise early
		if err := os.MkdirAll(runDir, 0o750); err != nil {
			return NewValidationError("unable to create working directory %q, so we cannot run envoy", runDir)
		}
		runOpts.RunDir = runDir
	}
	return nil
}

func ensureEnvoyVersion(c *cli.Context, o *globals.GlobalOpts) error {
	if o.EnvoyVersion == "" { // not overridden for tests
		if err := setEnvoyVersion(c.Context, o); err != nil {
			return NewValidationError(err.Error())
		}
	}
	return nil
}

// setEnvoyVersion makes sure the $FUNC_E_HOME/version exists.
func setEnvoyVersion(ctx context.Context, o *globals.GlobalOpts) (err error) {
	var v version.Version
	if v, _, err = envoy.CurrentVersion(o.HomeDir); err != nil {
		return err
	} else if v != nil { // We found an existing version, but it might be in MinorVersion format!
		o.EnvoyVersion, err = ensurePatchVersion(ctx, o, v)
		return err
	}

	// First time install: look up the latest version, which may be newer than version.LastKnownEnvoy!
	o.Logf("looking up the latest Envoy version\n")
	var evs *version.ReleaseVersions
	if evs, err = o.GetEnvoyVersions(ctx); err != nil {
		return fmt.Errorf("couldn't lookup the latest Envoy version from %s: %w", o.EnvoyVersionsURL, err)
	}
	o.EnvoyVersion = version.FindLatestVersion(versionsForPlatform(evs.Versions, o.Platform))
	if o.EnvoyVersion == "" {
		return fmt.Errorf("%s does not contain an Envoy release for platform %s", o.EnvoyVersionsURL, o.Platform)
	}
	// Persist it as a minor version, so that each invocation checks for the latest patch.
	return envoy.WriteCurrentVersion(o.EnvoyVersion.ToMinor(), o.HomeDir)
}

// ensurePatchVersion ensures we either have a valid version.PatchVersion or an error
// If remote lookup of the latest patch fails, this logs and falls back to the last installed one
// NOTE: Warnings and errors include the platform because a release isn't available at the same time for all platforms.
func ensurePatchVersion(ctx context.Context, o *globals.GlobalOpts, v version.Version) (version.PatchVersion, error) {
	if mv, ok := v.(version.MinorVersion); ok {
		o.Logf("looking up the latest patch for Envoy version %s\n", mv)
		evs, err := o.GetEnvoyVersions(ctx)
		var patchVersions []version.PatchVersion
		if err == nil {
			patchVersions = versionsForPlatform(evs.Versions, o.Platform)
			if pv := version.FindLatestPatchVersion(patchVersions, mv); pv != "" {
				return pv, nil
			}
			err = fmt.Errorf("%s does not contain an Envoy release for version %s on platform %s", o.EnvoyVersionsURL, mv, o.Platform)
		}

		// Attempt the last installed version instead of raising an error. There may not be one!
		if rows, e := getInstalledVersions(o.HomeDir); e == nil {
			for _, r := range rows { //nolint:gocritic
				patchVersions = append(patchVersions, r.version)
			}
			if pv := version.FindLatestPatchVersion(patchVersions, mv); pv != "" {
				o.Logf("couldn't look up an Envoy release for version %s on platform %s: using last installed version\n", mv, o.Platform)
				return pv, nil
			}
		}
		return "", err
	} // version.Version is a union type, so the only other option is a patch!
	return v.(version.PatchVersion), nil
}

func versionsForPlatform(vs map[version.PatchVersion]version.Release, p version.Platform) []version.PatchVersion {
	var patchVersions []version.PatchVersion
	for k, v := range vs {
		if _, ok := v.Tarballs[p]; ok {
			patchVersions = append(patchVersions, k)
		}
	}
	return patchVersions
}
