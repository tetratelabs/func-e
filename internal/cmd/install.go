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

// NewInstallCmd create a command responsible for downloading and extracting Envoy
func NewInstallCmd(o *globals.GlobalOpts) *cli.Command {
	return &cli.Command{
		Name:      "install",
		Usage:     "Download and install a <version> of Envoy",
		ArgsUsage: "<version>",
		Description: fmt.Sprintf(`The '<version>' is from the "versions" command.
The Envoy <version> will be installed into $GETENVOY_HOME/versions/<version>

Example:
$ getenvoy install %s`, version.LastKnownEnvoy),
		Before: validateVersionArg,
		Action: func(c *cli.Context) error {
			_, err := envoy.InstallIfNeeded(o, globals.CurrentPlatform, c.Args().First())
			return err
		},
	}
}
