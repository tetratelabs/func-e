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
	"github.com/urfave/cli/v2"

	"github.com/tetratelabs/func-e/internal/envoy"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/version"
)

// NewUseCmd create a command responsible for downloading and extracting Envoy
func NewUseCmd(o *globals.GlobalOpts) *cli.Command {
	versionsDir := moreos.ReplacePathSeparator("$FUNC_E_HOME/versions/")
	currentVersionWorkingDirFile := moreos.ReplacePathSeparator(envoy.CurrentVersionWorkingDirFile)
	currentVersionHomeDirFile := moreos.ReplacePathSeparator(envoy.CurrentVersionHomeDirFile)

	var v version.Version
	return &cli.Command{
		Name:      "use",
		Usage:     `Sets the current [version] used by the "run" command`,
		ArgsUsage: "[version]",
		Description: moreos.Sprintf(`The '[version]' is from the "versions -a" command.
The Envoy [version] installs on-demand into `+versionsDir+`[version]
if needed. You may also exclude the patch component of the [version]
to use the latest patch version or to download the binary if it is
not already downloaded.

This updates %s or %s with [version],
depending on which is present.

Example:
$ func-e use %s
$ func-e use %s`, currentVersionWorkingDirFile, currentVersionHomeDirFile, version.LastKnownEnvoy, version.LastKnownEnvoyMinor),
		Before: func(c *cli.Context) (err error) {
			if v, err = version.NewVersion("[version] argument", c.Args().First()); err != nil {
				err = NewValidationError(err.Error())
			}
			return
		},
		Action: func(c *cli.Context) (err error) {
			// The argument could be a MinorVersion (ex. 1.19) or a PatchVersion (ex. 1.19.3)
			// We need to download and install a patch version
			if o.EnvoyVersion, err = ensurePatchVersion(c.Context, o, v); err != nil {
				return err
			}
			if _, err = envoy.InstallIfNeeded(c.Context, o); err != nil {
				return err
			}
			// Persist the input precision. This allows those specifying a MinorVersion to always get the latest patch.
			return envoy.WriteCurrentVersion(v, o.HomeDir)
		},
		CustomHelpTemplate: cli.CommandHelpTemplate,
	}
}
