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
	"fmt"

	"github.com/spf13/cobra"
	scaffold "github.com/tetratelabs/getenvoy/pkg/extension/init"
	uiutil "github.com/tetratelabs/getenvoy/pkg/util/ui"
)

// feedback communicates to a user progress of the `init` command.
type feedback struct {
	cmd        *cobra.Command
	opts       *scaffold.ScaffoldOpts
	usedWizard bool
}

func (f *feedback) OnStart() {
	f.cmd.Println(uiutil.Underline("Scaffolding a new extension:"))
	f.cmd.Println(uiutil.Style("Generating files in {{ . | faint }}:").Apply(f.opts.OutputDir))
}

func (f *feedback) OnFile(file string) {
	f.cmd.Println(uiutil.Style(fmt.Sprintf(`{{ "%s" | green }} {{ . }}`, uiutil.IconGood)).Apply(file))
}

func (f feedback) OnComplete() {
	f.cmd.Println("Done!")
	if f.usedWizard {
		f.cmd.Println()
		f.cmd.Println(uiutil.Style(`{{ . | underline | faint }}`).Apply("Hint:"))
		f.cmd.Println(uiutil.Faint("Next time you can skip the wizard by running"))
		f.cmd.Println(uiutil.Faint(
			fmt.Sprintf("  %s --category %s --language %s %s", f.cmd.CommandPath(), f.opts.Category, f.opts.Language, f.opts.OutputDir)))
	}
}
