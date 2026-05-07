// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/tetratelabs/func-e/internal/envoy"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/runtime"
	"github.com/tetratelabs/func-e/internal/version"
)

// NewUseCmd create a command responsible for downloading and extracting Envoy
func NewUseCmd(o *globals.GlobalOpts) *cli.Command {
	versionsDir := "$FUNC_E_DATA_HOME/envoy-versions/"
	currentVersionWorkingDirFile := envoy.CurrentVersionWorkingDirFile
	currentVersionConfigFile := envoy.CurrentVersionConfigFile

	var v version.Version
	return &cli.Command{
		Name:      "use",
		Usage:     `Sets the current [version] used by the "run" command`,
		ArgsUsage: "[version]",
		HideHelp:  true,
		Description: fmt.Sprintf(`The '[version]' is from the "versions -a" command.
The Envoy [version] installs on-demand into `+versionsDir+`[version]
if needed. You may also exclude the patch component of the [version]
to use the latest patch version or to download the binary if it is
not already downloaded.

This updates %s or %s with [version],
depending on which is present.

Example:
$ func-e use %s
$ func-e use %s`, currentVersionWorkingDirFile, currentVersionConfigFile, version.LastKnownEnvoy, version.LastKnownEnvoyMinor),
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			var err error
			if v, err = version.NewVersion("[version] argument", c.Args().First()); err != nil {
				return ctx, NewValidationError(err.Error())
			}
			return ctx, nil
		},
		Action: func(ctx context.Context, _ *cli.Command) (err error) {
			// Create base XDG directories before any file operations
			if err := o.Mkdirs(); err != nil {
				return err
			}
			// The argument could be a MinorVersion (ex. 1.19) or a PatchVersion (ex. 1.19.3)
			// We need to download and install a patch version
			if o.EnvoyVersion, err = runtime.EnsurePatchVersion(ctx, o, v); err != nil {
				return err
			}
			if _, err = envoy.InstallIfNeeded(ctx, o); err != nil {
				return err
			}
			// Persist the input precision. This allows those specifying a MinorVersion to always get the latest patch.
			return envoy.WriteCurrentVersion(v, o.ConfigHome, o.EnvoyVersionFile())
		},
	}
}
