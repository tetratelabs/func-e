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
	"os"

	"github.com/spf13/cobra"

	"github.com/tetratelabs/getenvoy/pkg/globals"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
)

// NewListCmd returns command that lists available Envoy binaries
func NewListCmd(o *globals.GlobalOpts) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available Envoy version references you can run",
		RunE: func(cmd *cobra.Command, _ []string) error {
			m, err := manifest.FetchManifest(o.ManifestURL)
			if err != nil {
				return err
			}
			return manifest.Print(m, os.Stdout)
		},
	}
}
