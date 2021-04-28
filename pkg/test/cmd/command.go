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
	"bytes"

	"github.com/spf13/cobra"

	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/globals"
	"github.com/tetratelabs/getenvoy/pkg/cmd"
)

// NewRootCommand initializes a command with buffers for stdout and stderr.
func NewRootCommand(o *globals.GlobalOpts) (c *cobra.Command, stdout, stderr *bytes.Buffer) {
	stdout = new(bytes.Buffer)
	stderr = new(bytes.Buffer)
	c = cmd.NewRoot(o)
	c.SetOut(stdout)
	c.SetErr(stderr)
	return c, stdout, stderr
}
