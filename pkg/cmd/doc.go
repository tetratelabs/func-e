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
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	outputDir string
	linkDir   string
)

// NewDocCmd returns command that generates documentation
func NewDocCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "doc",
		Short:  "Generates Markdown documentation for the CLI.",
		Hidden: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			getenvoy := cmd.Parent()
			return doc.GenMarkdownTreeCustom(getenvoy, outputDir, filePrepender, linkHandler)
		},
	}
	cmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "", "directory to create generated docs")
	cmd.PersistentFlags().StringVarP(&linkDir, "link", "l", "", "directory to prepend to filename in links")
	return cmd
}

func filePrepender(filename string) string {
	name := filepath.Base(filename)
	base := strings.TrimSuffix(name, path.Ext(name))
	split := strings.Split(base, "_")
	parent := "root"
	if len(split) > 1 {
		parent = split[len(split)-2]
	}
	command := split[len(split)-1]
	return fmt.Sprintf(fmTemplate, strings.Join(split, " "), parent, command)
}

const fmTemplate = `+++
title = "%s"
type = "reference"
parent = "%s"
command = "%s"
+++
`

func linkHandler(name string) string {
	base := strings.TrimSuffix(name, path.Ext(name))
	return filepath.Join(linkDir, strings.ToLower(base))
}
