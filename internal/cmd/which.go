// Copyright 2021 Tetrate
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
)

// NewWhichCmd create a command responsible for downloading printing the path to the Envoy binary
func NewWhichCmd(o *globals.GlobalOpts) *cli.Command {
	return &cli.Command{
		Name:        "which",
		Usage:       `Prints the path to the Envoy binary used by the "run" command`,
		Description: `The binary is downloaded as necessary. The version is controllable by the "use" command`,
		Before: func(c *cli.Context) error {
			// no logging on version query/download. This is deferred until we know we are executing "which"
			o.Quiet = true
			return ensureEnvoyVersion(c, o)
		},
		Action: func(c *cli.Context) error {
			ev, err := envoy.InstallIfNeeded(c.Context, o, o.EnvoyVersion)
			if err != nil {
				return err
			}
			_, err = moreos.Fprintf(o.Out, "%s\n", ev)
			return err
		},
		CustomHelpTemplate: moreos.Sprintf(cli.CommandHelpTemplate),
	}
}
