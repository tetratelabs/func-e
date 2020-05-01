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
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"

	uiutil "github.com/tetratelabs/getenvoy/pkg/util/ui"
)

var _ = Describe("interactive mode", func() {
	Describe("wizard", func() {
		Describe("all parameters are valid", func() {
			var tmpDir string

			BeforeEach(func() {
				dir, err := ioutil.TempDir("", "test")
				Expect(err).ToNot(HaveOccurred())
				tmpDir = dir
			})

			AfterEach(func() {
				if tmpDir != "" {
					Expect(os.RemoveAll(tmpDir)).To(Succeed())
				}
			})

			type testCase struct {
				noColors bool
				expected string
			}
			DescribeTable("should not require user interaction if all parameters are valid",
				func(givenFn func() testCase) {
					given := givenFn()

					uiutil.StylesEnabled = !given.noColors

					cmd := new(cobra.Command)
					out := new(bytes.Buffer)
					cmd.SetOut(out)

					params := newParams()
					params.Category.Value = "envoy.filters.http"
					params.Language.Value = "rust"
					params.OutputDir.Value = tmpDir

					err := newWizard(cmd).Fill(params)
					Expect(err).ToNot(HaveOccurred())

					Expect(out.String()).To(Equal(given.expected))
				},
				Entry("colors disabled", func() testCase {
					return testCase{
						noColors: true,
						expected: fmt.Sprintf(`What kind of extension would you like to create?
✔ Category HTTP Filter
✔ Language Rust
✔ Output directory %s
Great! Let me help you with that!

`, tmpDir),
					}
				}),
				Entry("colors enabled", func() testCase {
					return testCase{
						noColors: false,
						expected: "\x1b[4mWhat kind of extension would you like to create?\x1b[0m\n" +
							"\x1b[32m✔\x1b[0m \x1b[3mCategory\x1b[0m \x1b[2mHTTP Filter\x1b[0m\n" +
							"\x1b[32m✔\x1b[0m \x1b[3mLanguage\x1b[0m \x1b[2mRust\x1b[0m\n" +
							fmt.Sprintf("\x1b[32m✔\x1b[0m \x1b[3mOutput directory\x1b[0m \x1b[2m%s\x1b[0m\n", tmpDir) +
							"Great! Let me help you with that!\n" +
							"\n",
					}
				}),
			)
		})
	})
})
