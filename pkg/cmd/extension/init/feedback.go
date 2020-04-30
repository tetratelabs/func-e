// Copyright 2020 Tetrate
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

package init

import (
	"github.com/spf13/cobra"
	scaffold "github.com/tetratelabs/getenvoy/pkg/extension/init"
)

// feedback communicates to a user progress of the `init` command.
type feedback struct {
	cmd  *cobra.Command
	opts *scaffold.ScaffoldOpts
}

func (f *feedback) OnStart() {
	f.cmd.Printf("Scaffolding a new extension in %s:\n", f.opts.OutputDir)
	f.cmd.Println()
	f.cmd.Println("* Generating files:")
}

func (f *feedback) OnFile(file string) {
	f.cmd.Printf("  √ %s\n", file)
}

func (f feedback) OnComplete() {
	f.cmd.Println()
	f.cmd.Println("Done!")
}
