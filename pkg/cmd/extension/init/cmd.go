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
	"strings"

	"github.com/spf13/cobra"

	"github.com/tetratelabs/getenvoy/pkg/cmd/extension/globals"
	scaffold "github.com/tetratelabs/getenvoy/pkg/extension/init"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
)

var (
	// extension categories supported by the `init` command.
	supportedCategories = options{
		{Value: extension.EnvoyHTTPFilter.String(), DisplayText: "HTTP Filter"},
		{Value: extension.EnvoyNetworkFilter.String(), DisplayText: "Network Filter"},
		{Value: extension.EnvoyAccessLogger.String(), DisplayText: "Access Logger"},
	}
	// programming languages supported by the `init` command.
	supportedLanguages = options{
		{Value: extension.LanguageRust.String(), DisplayText: "Rust"},
	}
)

// NewCmd returns a command that generates the initial set of files
// to kick off development of a new extension.
func NewCmd() *cobra.Command {
	params := newParams()
	cmd := &cobra.Command{
		Use:   "init [DIR]",
		Short: "Scaffold a new Envoy extension.",
		Long: `
Scaffold a new Envoy extension in a language of your choice.`,
		Example: `
  # Scaffold a new Envoy HTTP filter in Rust in the current working directory.
  getenvoy extension init --category envoy.filters.http --language rust

  # Scaffold a new Envoy Access logger in Rust in the "my-access-logger" directory.
  getenvoy extension init my-access-logger --category envoy.access_loggers --language rust`,
		RunE: func(cmd *cobra.Command, args []string) error {
			outputDir := ""
			if len(args) > 0 {
				outputDir = args[0]
			}
			params.OutputDir.Value = outputDir

			usedWizard := false
			if err := params.Validate(); err != nil {
				if globals.NoPrompt {
					return err
				}
				if err := newWizard(cmd).Fill(params); err != nil {
					return err
				}
				usedWizard = true
			}

			opts := &scaffold.ScaffoldOpts{}
			opts.Category = params.Category.Value
			opts.Language = params.Language.Value
			opts.TemplateName = "default"
			opts.OutputDir = params.OutputDir.Value
			opts.ProgressHandler = &feedback{
				cmd:        cmd,
				opts:       opts,
				usedWizard: usedWizard,
			}
			return scaffold.Scaffold(opts)
		},
	}
	cmd.PersistentFlags().StringVar(&params.Category.Value, "category", "", "choose extension category. "+hintOneOf(supportedCategories.Values()...))
	cmd.PersistentFlags().StringVar(&params.Language.Value, "language", "", "choose programming language. "+hintOneOf(supportedLanguages.Values()...))
	return cmd
}

func hintOneOf(values ...string) string {
	texts := make([]string, len(values))
	for i := range values {
		texts[i] = fmt.Sprintf("%q", values[i])
	}
	return "One of: " + strings.Join(texts, ", ")
}
