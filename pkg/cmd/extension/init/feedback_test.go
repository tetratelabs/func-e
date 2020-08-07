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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"

	scaffold "github.com/tetratelabs/getenvoy/pkg/extension/init"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
	uiutil "github.com/tetratelabs/getenvoy/pkg/util/ui"
)

var _ = Describe("feedback", func() {
	Describe("console output", func() {
		type testCase struct {
			noColors   bool
			usedWizard bool
			expected   string
		}
		DescribeTable("should depend on whether colors are enabled",
			func(given testCase) {
				uiutil.StylesEnabled = !given.noColors

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
					usedWizard: given.usedWizard,
					w:          c.ErrOrStderr(),
				}

				f.OnStart()
				f.OnFile("Cargo.toml")
				f.OnFile("src/lib.rs")
				f.OnComplete()

				Expect(stdout.String()).To(BeEmpty())
				Expect(stderr.String()).To(Equal(given.expected))
			},
			Entry("--no-colors", testCase{
				noColors:   true,
				usedWizard: false,
				expected: `Scaffolding a new extension:
Generating files in /path/to/dir:
* Cargo.toml
* src/lib.rs
Done!
`,
			}),
			Entry("--no-colors + wizard", testCase{
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
			}),
			Entry("--no-colors=false", testCase{
				noColors:   false,
				usedWizard: false,
				expected: "\x1b[4mScaffolding a new extension:\x1b[0m\n" +
					"Generating files in \x1b[2m/path/to/dir\x1b[0m:\n" +
					"\x1b[32m✔\x1b[0m Cargo.toml\n" +
					"\x1b[32m✔\x1b[0m src/lib.rs\n" +
					"Done!\n",
			}),
			Entry("--no-colors=false + wizard", testCase{
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
			}),
		)
	})
})
