// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/tetratelabs/func-e/internal/envoy"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/version"
)

// EnsurePatchVersion ensures we either have a valid version.PatchVersion or an error
// If remote lookup of the latest patch fails, this logs and falls back to the last installed one
// NOTE: Warnings and errors include the platform because a release isn't available at the same time for all platforms.
func EnsurePatchVersion(ctx context.Context, o *globals.GlobalOpts, v version.Version) (version.PatchVersion, error) {
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
		if rows, e := getInstalledVersions(o.EnvoyVersionsDir()); e == nil {
			for _, r := range rows {
				patchVersions = append(patchVersions, r.version)
			}
			if pv := version.FindLatestPatchVersion(patchVersions, mv); pv != "" {
				o.Logf("couldn't look up an Envoy release for version %s on platform %s: using last installed version\n", mv, o.Platform)
				return pv, nil
			}
		}
		return "", err
	} // version.Version is a union type, so the only other option is a patch!
	vv, ok := v.(version.PatchVersion)
	if !ok {
		panic(fmt.Sprintf("unexpected version type %T", v))
	}
	return vv, nil
}

// Run runs Envoy with the given arguments.
// Returns nil when Envoy exits cleanly, including when interrupted by signals (SIGINT/SIGTERM).
// This matches Envoy's behavior of returning exit code 0 on graceful shutdown.
func Run(ctx context.Context, o *globals.GlobalOpts, args []string) error {
	if err := initializeRunOpts(ctx, o); err != nil {
		return err
	}

	stateDir := o.RunDir
	r := envoy.NewRuntime(&o.RunOpts, o.Logf)

	stdoutLog, err := os.OpenFile(filepath.Join(stateDir, "stdout.log"), os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("couldn't create stdout log file: %w", err)
	}
	defer stdoutLog.Close() //nolint
	r.OutFile = stdoutLog
	r.Out = io.MultiWriter(o.EnvoyOut, stdoutLog)

	stderrLog, err := os.OpenFile(filepath.Join(stateDir, "stderr.log"), os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("couldn't create stderr log file: %w", err)
	}
	defer stderrLog.Close() //nolint
	r.ErrFile = stderrLog
	r.Err = io.MultiWriter(o.EnvoyErr, stderrLog)

	return r.Run(ctx, args)
}

// setEnvoyVersion makes sure the version file exists.
func setEnvoyVersion(ctx context.Context, o *globals.GlobalOpts) (err error) {
	var v version.Version
	if v, _, err = envoy.CurrentVersion(o.DataHome, o.EnvoyVersionFile(), o.EnvoyVersionFileSource()); err != nil {
		return err
	} else if v != nil { // We found an existing version, but it might be in MinorVersion format!
		o.EnvoyVersion, err = EnsurePatchVersion(ctx, o, v)
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
	return envoy.WriteCurrentVersion(o.EnvoyVersion.ToMinor(), o.DataHome, o.EnvoyVersionFile())
}

// initializeRunOpts initializes the api options
func initializeRunOpts(ctx context.Context, o *globals.GlobalOpts) error {
	runOpts := &o.RunOpts

	// Set up directories using pre-generated runID
	if runOpts.RunDir == "" { // not overridden for tests
		runOpts.RunDir = o.EnvoyRunDir(o.RunID)
	}
	if runOpts.TempDir == "" { // not overridden for tests
		runOpts.TempDir = o.EnvoyRuntimeDir(o.RunID)
	}
	if runOpts.RunID == "" { // not overridden for tests
		runOpts.RunID = o.RunID
	}

	// Create all XDG directories now that runID is finalized
	if err := o.Mkdirs(); err != nil {
		return err
	}

	if o.EnvoyPath == "" { // not overridden for tests
		envoyPath, err := envoy.InstallIfNeeded(ctx, o)
		if err != nil {
			return err
		}
		o.EnvoyPath = envoyPath
	}

	return nil
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

type versionReleaseDate struct {
	version     version.PatchVersion
	releaseDate version.ReleaseDate
}

func getInstalledVersions(versionsDir string) ([]versionReleaseDate, error) {
	var rows []versionReleaseDate
	files, err := os.ReadDir(versionsDir)
	if os.IsNotExist(err) {
		return rows, nil
	} else if err != nil {
		return nil, err
	}

	for _, f := range files {
		pv := version.NewPatchVersion(f.Name())
		if i, err := f.Info(); f.IsDir() && pv != "" && err == nil {
			rows = append(rows, versionReleaseDate{
				pv,
				version.ReleaseDate(i.ModTime().Format("2006-01-02")),
			})
		}
	}
	return rows, nil
}
