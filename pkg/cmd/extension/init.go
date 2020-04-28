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
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	scaffold "github.com/tetratelabs/getenvoy/pkg/extension/init"
	osutil "github.com/tetratelabs/getenvoy/pkg/util/os"
)

// extension categories supported by `init` command.
const (
	envoyHTTPFilter    = "envoy.filters.http"
	envoyNetworkFilter = "envoy.filters.network"
	envoyAccessLogger  = "envoy.access_loggers"
)

// programming languages supported by `init` command.
const (
	languageRust = "rust"
)

// options represents an exhaustive list of valid values.
type options []string

func (o options) Contains(value string) bool {
	for _, option := range o {
		if value == option {
			return true
		}
	}
	return false
}

var (
	allSupportedCategories = options{envoyHTTPFilter, envoyNetworkFilter, envoyAccessLogger}
	allSupportedLanguages  = options{languageRust}
)

// NewInitCmd returns a command that generates the initial set of files
// to kick off development of a new extension.
func NewInitCmd() *cobra.Command {
	var category string
	var language string
	cmd := &cobra.Command{
		Use:   "init [dir]",
		Short: "Scaffold a new Envoy extension.",
		Long: `
Scaffold a new Envoy extension in a language of your choice.`,
		Example: `
  # Scaffold a new Envoy HTTP filter in Rust in the current working directory.
  getenvoy extension init --category envoy.filters.http --language rust

  # Scaffold a new Envoy Access logger in Rust in the "my-access-logger" directory.
  getenvoy extension init my-access-logger --category envoy.access_loggers --language rust`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := scaffold.ScaffoldOpts{}
			if !allSupportedCategories.Contains(category) {
				return fmt.Errorf("extension %q has invalid value %q", "category", category)
			}
			opts.Category = category
			if !allSupportedLanguages.Contains(language) {
				return fmt.Errorf("extension %q has invalid value %q", "language", language)
			}
			opts.Language = language
			opts.TemplateName = "default"

			outputDir, err := filepath.Abs(filepath.Clean(optionalArg(args[:1]).ValueOr("")))
			if err != nil {
				return err
			}
			if err := osutil.EnsureDirExists(outputDir); err != nil {
				return err
			}
			if empty, err := osutil.IsEmptyDir(outputDir); err != nil || !empty {
				if err != nil {
					return err
				}
				if len(args) == 0 {
					return fmt.Errorf("unable to scaffold a new extension in the current working directory since it's not empty.\n"+
						"\nHint: consider providing a name for a new directory to scaffold in, e.g.\n"+
						"\n  getenvoy extension init my-new-extension --category=%s --language=%s\n",
						category, language,
					)
				}
				return fmt.Errorf("cowardly refusing to scaffold a new extension in a non-empty directory: %v", outputDir)
			}
			opts.OutputDir = outputDir
			opts.ProgressHandler = scaffold.ProgressFuncs{
				OnStartFunc: func() {
					cmd.Printf("Scaffolding a new extension in %s:\n", opts.OutputDir)
					cmd.Print("\n")
					cmd.Print("* Generating files:\n")
				},
				OnFileFunc: func(file string) {
					cmd.Printf("  √ %s\n", file)
				},
				OnCompleteFunc: func() {
					cmd.Print("\n")
					cmd.Print("Done!\n")
				},
			}
			return scaffold.Scaffold(&opts)
		},
	}
	cmd.PersistentFlags().StringVar(&category, "category", "", "choose extension category. "+hintOneOf(allSupportedCategories...))
	cmd.PersistentFlags().StringVar(&language, "language", "", "choose programming language. "+hintOneOf(allSupportedLanguages...))
	return cmd
}

// optionalArg represents an optional command-line argument.
type optionalArg []string

func (o optionalArg) ValueOr(defaultValue string) string {
	if len(o) > 0 {
		return o[0]
	}
	return defaultValue
}
