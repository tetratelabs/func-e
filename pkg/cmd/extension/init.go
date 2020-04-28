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
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	scaffold "github.com/tetratelabs/getenvoy/pkg/extension/init"
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
		Args: cobra.MaximumNArgs(1),
		Example: `
  # Scaffold a new Envoy HTTP filter in Rust in the current working directory.
  getenvoy extension init --category envoy.filters.http --language rust`,
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

			outputDir, err := inferOutputDir(args[:1])
			if err != nil {
				return err
			}
			if err := ensureDirExists(outputDir); err != nil {
				return err
			}
			if empty, err := isEmptyDir(outputDir); err != nil || !empty {
				if err != nil {
					return err
				}
				return fmt.Errorf("cowardly refusing to scaffold a new extension because output directory is not empty: %v", outputDir)
			}
			opts.OutputDir = outputDir
			return scaffold.Scaffold(opts)
		},
	}
	cmd.PersistentFlags().StringVar(&category, "category", "", "choose extension category. "+hintOneOf(allSupportedCategories...))
	cmd.PersistentFlags().StringVar(&language, "language", "", "choose programming language. "+hintOneOf(allSupportedLanguages...))
	return cmd
}

func inferOutputDir(args []string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if len(args) > 0 {
		dir := filepath.Clean(args[0])
		if path.IsAbs(dir) {
			return dir, nil
		}
		return filepath.Join(cwd, dir), nil
	}
	return cwd, nil
}

func ensureDirExists(name string) error {
	if err := os.MkdirAll(name, os.ModeDir|0755); err != nil {
		return err
	}
	return nil
}

func isEmptyDir(name string) (empty bool, errs error) {
	dir, err := os.Open(filepath.Clean(name))
	if err != nil {
		return false, err
	}
	defer func() {
		if e := dir.Close(); e != nil {
			errs = multierror.Append(errs, e)
		}
	}()
	files, err := dir.Readdirnames(1)
	if err != nil && err != io.EOF {
		return false, err
	}
	return len(files) == 0, nil
}
