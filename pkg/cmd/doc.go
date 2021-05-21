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
	"strings"

	"github.com/urfave/cli/v2"
)

// NewDocCmd returns command that generates documentation
func NewDocCmd() *cli.Command {
	cmd := &cli.Command{
		Name:   "doc",
		Usage:  "Generates Markdown documentation for the CLI.",
		Hidden: true,
		Action: func(c *cli.Context) error {
			m, err := c.App.ToMarkdown()
			m = strings.ReplaceAll(m, "% getenvoy 8\n\n", "") // remove man header until urfave/cli#1275
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(c.App.Writer, m)
			return err
		},
	}
	return cmd
}
