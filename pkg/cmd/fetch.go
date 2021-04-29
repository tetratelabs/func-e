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
	"fmt"

	"github.com/spf13/cobra"

	reference "github.com/tetratelabs/getenvoy/pkg"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/globals"
)

// NewFetchCmd create a command responsible for retrieving Envoy binaries
func NewFetchCmd(o *globals.GlobalOpts) *cobra.Command {
	return &cobra.Command{
		Use:   "fetch <reference>",
		Short: "Retrieve Envoy binaries from GetEnvoy.",
		Long: `
Retrieves the referenced Envoy binary from GetEnvoy. The reference can be a full or partial reference.
A complete list of available builds can be retrieved using` + "`getenvoy list`" + `.`,
		Example: fmt.Sprintf(`# Fetch using a partial manifest reference to retrieve a build suitable for your operating system.
getenvoy fetch %[1]s
		
# Fetch using a full manifest reference to retrieve a specific build for Linux. 
getenvoy fetch s%[1]s/linux-glibc`, reference.Latest),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing reference parameter")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := envoy.FetchIfNeeded(o, args[0])
			return err
		},
	}
}
