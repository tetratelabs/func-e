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

	"github.com/tetratelabs/getenvoy/internal/reference"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/globals"
)

// NewFetchCmd create a command responsible for retrieving Envoy binaries
func NewFetchCmd(o *globals.GlobalOpts) *cli.Command {
	return &cli.Command{
		Name:      "fetch",
		Usage:     "Downloads a version of Envoy. Available builds can be retrieved using `getenvoy list`.",
		ArgsUsage: "<reference>",
		Description: fmt.Sprintf(`# Fetch using a partial manifest reference to retrieve a build suitable for your operating system.
getenvoy fetch %[1]s

# Fetch using a full manifest reference to retrieve a specific build for Linux.
getenvoy fetch %[1]s/linux-glibc`, reference.Latest),
		Before: validateReferenceArg,
		Action: func(c *cli.Context) error {
			_, err := envoy.FetchIfNeeded(o, c.Args().First())
			return err
		},
	}
}
