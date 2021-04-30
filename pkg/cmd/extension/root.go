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

package extension

import (
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/tetratelabs/getenvoy/pkg/cmd/extension/build"
	"github.com/tetratelabs/getenvoy/pkg/cmd/extension/clean"
	"github.com/tetratelabs/getenvoy/pkg/cmd/extension/example"
	scaffold "github.com/tetratelabs/getenvoy/pkg/cmd/extension/init"
	"github.com/tetratelabs/getenvoy/pkg/cmd/extension/push"
	"github.com/tetratelabs/getenvoy/pkg/cmd/extension/run"
	"github.com/tetratelabs/getenvoy/pkg/cmd/extension/test"
	"github.com/tetratelabs/getenvoy/pkg/globals"
)

// NewCmd returns a command that aggregates all extension-related commands.
func NewCmd(o *globals.GlobalOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extension",
		Short: "Delve into Envoy extensions.",
		Long:  `Explore ready-to-use Envoy extensions or develop a new one.`,
	}
	cmd.AddCommand(scaffold.NewCmd(o))
	cmd.AddCommand(build.NewCmd(o))
	cmd.AddCommand(test.NewCmd(o))
	cmd.AddCommand(clean.NewCmd(o))
	cmd.AddCommand(example.NewCmd(o))
	cmd.AddCommand(run.NewCmd(o))
	cmd.AddCommand(push.NewCmd(o))
	if !o.NoWizard { // not overridden for tests
		cmd.PersistentFlags().BoolVar(&o.NoWizard, "no-prompt", noPromptDefault(),
			"disable automatic switching into interactive mode whenever a parameter is missing or not valid")
	}
	if !o.NoColors { // not overridden for tests
		cmd.PersistentFlags().BoolVar(&o.NoColors, "no-colors", noColorsDefault(), "disable colored output")
	}
	return cmd
}

func noPromptDefault() bool {
	return !(isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stderr.Fd()))
}

func noColorsDefault() bool {
	return !(isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stderr.Fd()))
}
