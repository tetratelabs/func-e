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
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/urfave/cli/v2"

	"github.com/tetratelabs/getenvoy/internal/globals"
)

// NewInstalledCmd returns command that lists installed Envoy versions
func NewInstalledCmd(o *globals.GlobalOpts) *cli.Command {
	return &cli.Command{
		Name:  "installed",
		Usage: "List installed Envoy versions",
		Action: func(c *cli.Context) error {
			files, err := os.ReadDir(filepath.Join(o.HomeDir, "versions"))
			if os.IsNotExist(err) {
				fmt.Fprintln(o.Out, "No envoy versions installed, yet")
				return nil
			} else if err != nil {
				return err
			}

			var rows []string
			for _, f := range files {
				if f.IsDir() {
					rows = append(rows, f.Name())
				}
			}

			// Sort so that new versions appear first
			sort.Slice(rows, func(i, j int) bool {
				return rows[i] > rows[j]
			})

			out := "VERSION\n"
			for _, vr := range rows {
				out += vr + "\n"
			}
			fmt.Fprint(o.Out, out)
			return nil
		},
	}
}
