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
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	scaffold "github.com/tetratelabs/getenvoy/pkg/extension/init"
	osutil "github.com/tetratelabs/getenvoy/pkg/util/os"
)

var (
	// extension categories supported by the `init` command.
	supportedCategories = options{
		{Value: "envoy.filters.http", DisplayText: "HTTP Filter"},
		{Value: "envoy.filters.network", DisplayText: "Network Filter"},
		{Value: "envoy.access_loggers", DisplayText: "Access Logger"},
	}
	// programming languages supported by the `init` command.
	supportedLanguages = options{
		{Value: "rust", DisplayText: "Rust"},
	}
)

// NewCmd returns a command that generates the initial set of files
// to kick off development of a new extension.
func NewCmd() *cobra.Command {
	var category string
	var language string
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
			opts := &scaffold.ScaffoldOpts{}
			if !supportedCategories.Contains(category) {
				return fmt.Errorf("%q is not a supported extension category", category)
			}
			opts.Category = category
			if !supportedLanguages.Contains(language) {
				return fmt.Errorf("%q is not a supported programming language", language)
			}
			opts.Language = language
			opts.TemplateName = "default"

			outputDir := ""
			if len(args) > 0 {
				outputDir = args[0]
			}
			outputDir, err := filepath.Abs(filepath.Clean(outputDir))
			if err != nil {
				return err
			}
			err = osutil.EnsureDirExists(outputDir)
			if err != nil {
				return err
			}
			empty, err := osutil.IsEmptyDir(outputDir)
			if err != nil {
				return err
			}
			if !empty {
				return fmt.Errorf("output directory must be empty or new: %v", outputDir)
			}
			opts.OutputDir = outputDir
			opts.ProgressHandler = &feedback{cmd: cmd, opts: opts}
			return scaffold.Scaffold(opts)
		},
	}
	cmd.PersistentFlags().StringVar(&category, "category", "", "choose extension category. "+hintOneOf(supportedCategories.Values()...))
	cmd.PersistentFlags().StringVar(&language, "language", "", "choose programming language. "+hintOneOf(supportedLanguages.Values()...))
	return cmd
}

func hintOneOf(values ...string) string {
	texts := make([]string, len(values))
	for i := range values {
		texts[i] = fmt.Sprintf("%q", values[i])
	}
	return "One of: " + strings.Join(texts, ", ")
}
