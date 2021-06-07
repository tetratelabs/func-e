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

	"github.com/urfave/cli/v2"

	"github.com/tetratelabs/getenvoy/internal/envoy"
	"github.com/tetratelabs/getenvoy/internal/globals"
	"github.com/tetratelabs/getenvoy/internal/version"
)

// NewUseCmd create a command responsible for downloading and extracting Envoy
func NewUseCmd(o *globals.GlobalOpts) *cli.Command {
	return &cli.Command{
		Name:      "use",
		Usage:     `Sets the current [version] used by the "run" command, installing as necessary`,
		ArgsUsage: "[version]",
		Description: fmt.Sprintf(`The '[version]' is from the "versions -a" command.
The Envoy [version] installs on-demand into $GETENVOY_HOME/versions/[version]
if needed.

This updates %s or %s with [version],
depending on which is present.

Example:
$ getenvoy use %s`, envoy.CurrentVersionWorkingDirFile, envoy.CurrentVersionHomeDirFile, version.LastKnownEnvoy),
		Before: validateVersionArg,
		Action: func(c *cli.Context) error {
			v := c.Args().First()
			if _, err := envoy.InstallIfNeeded(c.Context, o, globals.CurrentPlatform, v); err != nil {
				return err
			}
			return envoy.WriteCurrentVersion(v, o.HomeDir)
		},
	}
}

func validateVersionArg(c *cli.Context) error {
	if c.NArg() == 0 {
		return NewValidationError("missing [version] argument")
	}
	v := c.Args().First()
	if matched := globals.EnvoyVersionPattern.MatchString(v); !matched {
		return NewValidationError("invalid [version] argument: %q should look like %q", v, version.LastKnownEnvoy)
	}
	return nil
}
