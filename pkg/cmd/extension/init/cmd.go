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

	reference "github.com/tetratelabs/getenvoy/pkg"
	scaffold "github.com/tetratelabs/getenvoy/pkg/extension/init"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
	"github.com/tetratelabs/getenvoy/pkg/globals"
	uiutil "github.com/tetratelabs/getenvoy/pkg/util/ui"
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
		{Value: extension.LanguageTinyGo.String(), DisplayText: "TinyGo"},
	}
	// Envoy release the extension templates have been developed against.
	//
	// Notice that WebAssembly support in Envoy is still in the alpha stage.
	// It is not possible to guarantee any compatibility between various Envoy releases.
	// That is why we have to pin Envoy version by default.
	//
	// The value defined here will be included into extension descriptor to indicate
	// what version of Envoy extension examples should run on if not specified explicitly.
	// Extension developers will be able to explicitly associate each extension example with
	// a separate version of Envoy.
	// Extension users will be able to force getenvoy command to run an extension example
	// on the Envoy version of choice.
	//
	// `getenvoy extension run` command will choose version of Envoy to run the extension example on
	// using to the following rules (from the highest priority to the lowest):
	// 1. according to command-line options
	// 2. otherwise, according to the example-specific configuration (.getenvoy/extension/examples/<example>/example.yaml)
	// 3. otherwise, according to extension descriptor (.getenvoy/extension/extension.yaml)
	defaultSupportedEnvoyVersion = reference.Latest
)

// NewCmd returns a command that generates the initial set of files
// to kick off development of a new extension.
func NewCmd(o *globals.GlobalOpts) *cobra.Command {
	params := newParams()
	cmd := &cobra.Command{
		Use:   "init [DIR]",
		Short: "Scaffold a new Envoy extension.",
		Long: `
Scaffold a new Envoy extension in a language of your choice.`,
		Example: `
  # Scaffold a new extension in interactive mode.
  getenvoy extension init

  # Scaffold a new extension according to command options: Envoy HTTP filter, in Rust, with a given name, in the current working directory.
  getenvoy extension init --category envoy.filters.http --language rust --name mycompany.filters.http.custom_metrics

  # Scaffold a new extension according to command options: Envoy Access logger, in Rust, with a given name, in the "my-access-logger" directory.
  getenvoy extension init my-access-logger --category envoy.access_loggers --language rust --name mycompany.access_loggers.custom_log`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.ExtensionDir == "" { // not overridden for tests
				extensionDir := ""
				if len(args) > 0 {
					extensionDir = args[0]
				}
				abs, err := filepath.Abs(extensionDir)
				if err != nil {
					return err
				}
				o.ExtensionDir = abs
			}

			// Reassign any incoming extension directory value here. The wizard may override it later.
			params.ExtensionDir.Value = o.ExtensionDir
			styleFuncs := uiutil.NewStyleFuncs(o.NoColors)
			usedWizard := false
			if err := params.Validate(); err != nil {
				if o.NoWizard {
					return err
				}
				w := newWizard(styleFuncs, cmd.OutOrStderr())
				if err := w.Fill(params); err != nil {
					return err
				}
				usedWizard = true
			}

			descriptor, err := scaffold.NewExtension(params.Category.Value, params.Language.Value)
			if err != nil {
				return err
			}
			descriptor.Name = params.Name.Value
			descriptor.Runtime.Envoy.Version = defaultSupportedEnvoyVersion

			opts := &scaffold.ScaffoldOpts{
				Extension:    descriptor,
				ExtensionDir: params.ExtensionDir.Value,
				TemplateName: "default",
			}
			opts.ProgressSink = &feedback{
				cmd:        cmd,
				opts:       opts, // << love circular refs right?
				usedWizard: usedWizard,
				styleFuncs: styleFuncs,
				w:          cmd.ErrOrStderr(),
			}
			return scaffold.Scaffold(opts)
		},
	}
	cmd.PersistentFlags().StringVar(&params.Category.Value, "category", "", "Choose extension category. "+hintOneOf(supportedCategories.Values()...))
	cmd.PersistentFlags().StringVar(&params.Language.Value, "language", "", "Choose programming language. "+hintOneOf(supportedLanguages.Values()...))
	cmd.PersistentFlags().StringVar(&params.Name.Value, "name", "", `Choose extension name, e.g. "mycompany.filters.http.custom_metrics"`)
	return cmd
}

func hintOneOf(values ...string) string {
	texts := make([]string, len(values))
	for i := range values {
		texts[i] = fmt.Sprintf("%q", values[i])
	}
	return "One of: " + strings.Join(texts, ", ")
}
