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
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	scaffold "github.com/tetratelabs/getenvoy/pkg/extension/init"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
	uiutil "github.com/tetratelabs/getenvoy/pkg/util/ui"
)

func TestFeedbackArgs(t *testing.T) {
	type testCase struct {
		name       string
		noColors   bool
		usedWizard bool
		expected   string
	}
	tests := []testCase{
		{
			name:       "--no-colors",
			noColors:   true,
			usedWizard: false,
			expected: `Scaffolding a new extension:
Generating files in /path/to/dir:
* Cargo.toml
* src/lib.rs
Done!
`,
		},
		{
			name:       "--no-colors + wizard",
			noColors:   true,
			usedWizard: true,
			expected: `Scaffolding a new extension:
Generating files in /path/to/dir:
* Cargo.toml
* src/lib.rs
Done!

Hint:
Next time you can skip the wizard by running
  init --category envoy.filters.http --language rust --name my_company.my_http_filter /path/to/dir
`,
		}, {
			name:       "--no-colors=false",
			noColors:   false,
			usedWizard: false,
			expected: "\x1b[4mScaffolding a new extension:\x1b[0m\n" +
				"Generating files in \x1b[2m/path/to/dir\x1b[0m:\n" +
				"\x1b[32m✔\x1b[0m Cargo.toml\n" +
				"\x1b[32m✔\x1b[0m src/lib.rs\n" +
				"Done!\n",
		}, {
			name:       "--no-colors=false + wizard",
			noColors:   false,
			usedWizard: true,
			expected: "\x1b[4mScaffolding a new extension:\x1b[0m\n" +
				"Generating files in \x1b[2m/path/to/dir\x1b[0m:\n" +
				"\x1b[32m✔\x1b[0m Cargo.toml\n" +
				"\x1b[32m✔\x1b[0m src/lib.rs\n" +
				"Done!\n" +
				"\n" +
				"\x1b[2m\x1b[4mHint:\x1b[0m\n" +
				"\x1b[2mNext time you can skip the wizard by running\x1b[0m\n" +
				"\x1b[2m  init --category envoy.filters.http --language rust --name my_company.my_http_filter /path/to/dir\x1b[0m\n",
		},
	}
	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			uiutil.StylesEnabled = !test.noColors

			c := &cobra.Command{
				Use: "init",
			}
			stdout := new(bytes.Buffer)
			stderr := new(bytes.Buffer)
			c.SetOut(stdout)
			c.SetErr(stderr)

			f := &feedback{
				cmd: c,
				opts: &scaffold.ScaffoldOpts{
					Extension: &extension.Descriptor{
						Category: extension.EnvoyHTTPFilter,
						Language: extension.LanguageRust,
						Name:     "my_company.my_http_filter",
					},
					OutputDir: "/path/to/dir",
				},
				usedWizard: test.usedWizard,
				w:          c.ErrOrStderr(),
			}

			f.OnStart()
			f.OnFile("Cargo.toml")
			f.OnFile("src/lib.rs")
			f.OnComplete()

			require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
			require.Equal(t, test.expected, stderr.String(), `unexpected stderr running [%v]`, c)
		})
	}
}
