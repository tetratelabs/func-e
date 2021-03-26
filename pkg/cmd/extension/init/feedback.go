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
	"io"

	"github.com/spf13/cobra"

	scaffold "github.com/tetratelabs/getenvoy/pkg/extension/init"
	uiutil "github.com/tetratelabs/getenvoy/pkg/util/ui"
)

// feedback communicates to a user progress of the `getenvoy extension init` command.
type feedback struct {
	cmd        *cobra.Command
	opts       *scaffold.ScaffoldOpts
	usedWizard bool
	w          io.Writer
}

func (f *feedback) OnStart() {
	fmt.Fprintln(f.w, uiutil.Underline("Scaffolding a new extension:"))
	fmt.Fprintln(f.w, uiutil.Style("Generating files in {{ . | faint }}:").Apply(f.opts.OutputDir))
}

func (f *feedback) OnFile(file string) {
	fmt.Fprintln(f.w, uiutil.Style(`{{ icon "good" | green }} {{ . }}`).Apply(file))
}

func (f feedback) OnComplete() {
	fmt.Fprintln(f.w, "Done!")
	if f.usedWizard {
		fmt.Fprintln(f.w)
		fmt.Fprintln(f.w, uiutil.Style(`{{ . | underline | faint }}`).Apply("Hint:"))
		fmt.Fprintln(f.w, uiutil.Faint("Next time you can skip the wizard by running"))
		fmt.Fprintln(f.w, uiutil.Faint(
			fmt.Sprintf("  %s --category %s --language %s --name %s %s",
				f.cmd.CommandPath(),
				f.opts.Extension.Category,
				f.opts.Extension.Language,
				f.opts.Extension.Name,
				f.opts.OutputDir)))
	}
}
