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
	"errors"

	"github.com/spf13/cobra"
	"github.com/tetratelabs/getenvoy/pkg/binary/getenvoy"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
)

// NewFetchCmd create a command responsible for retrieving Envoy binaries
func NewFetchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fetch [binary]",
		Short: "Retreives the passed Envoy binary.",
		Long: `
Retreives the passed Envoy binary. 
Location can be a full or partial manifest reference, use getenvoy list for a complete list of available builds.`,
		Example: `# Fetch using a partial manifest reference to retrieve a build for your operating system.
getenvoy fetch standard:1.11.1
		
# Fetch using a full manifest reference to retrieve a specific build. 
getenvoy fetch standard:1.11.1/debian`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing binary parameter")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := manifest.NewKey(args[0])
			if err != nil {
				return err
			}
			location, err := manifest.Locate(key, manifestURL)
			if err != nil {
				return err
			}
			runtime, err := getenvoy.New()
			if err != nil {
				return err
			}
			return runtime.Fetch(key, location)
		},
	}
}
