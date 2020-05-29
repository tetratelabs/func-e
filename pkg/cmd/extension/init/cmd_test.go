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

package init_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"

	"github.com/tetratelabs/getenvoy/pkg/cmd"
	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	cmdutil "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

var _ = Describe("getenvoy extension init", func() {

	var stdout *bytes.Buffer
	var stderr *bytes.Buffer

	BeforeEach(func() {
		stdout = new(bytes.Buffer)
		stderr = new(bytes.Buffer)
	})

	var c *cobra.Command

	BeforeEach(func() {
		c = cmd.NewRoot()
		c.SetOut(stdout)
		c.SetErr(stderr)
	})

	Describe("should validate parameters", func() {
		type testCase struct {
			args           []string
			expectedStdErr string
		}
		type testCaseFn func() testCase
		give := func(given testCase) testCaseFn {
			return func() testCase {
				return given
			}
		}
		DescribeTable("should fail if a parameter is missing or has an invalid value",
			func(givenFn testCaseFn) {
				given := givenFn()

				By("running command")
				c.SetArgs(append([]string{"extension", "init", "--no-prompt"}, given.args...))
				err := cmdutil.Execute(c)
				Expect(err).To(HaveOccurred())

				By("verifying command output")
				Expect(stdout.String()).To(BeEmpty())
				Expect(stderr.String()).To(Equal(given.expectedStdErr))
			},
			Entry("extension category is missing", give(testCase{
				args: []string{},
				expectedStdErr: `Error: "" is not a supported extension category

Run 'getenvoy extension init --help' for usage.
`,
			})),
			Entry("extension category is not valid", give(testCase{
				args: []string{"--category", "invalid.category"},
				expectedStdErr: `Error: "invalid.category" is not a supported extension category

Run 'getenvoy extension init --help' for usage.
`,
			})),
			Entry("programming language is missing", give(testCase{
				args: []string{"--category", "envoy.filters.http"},
				expectedStdErr: `Error: "" is not a supported programming language

Run 'getenvoy extension init --help' for usage.
`,
			})),
			Entry("programming language is not valid", give(testCase{
				args: []string{"--category", "envoy.filters.http", "--language", "invalid.language"},
				expectedStdErr: `Error: "invalid.language" is not a supported programming language

Run 'getenvoy extension init --help' for usage.
`,
			})),
			Entry("output directory exists but is not empty", func() testCase {
				cwd, err := os.Getwd()
				Expect(err).ToNot(HaveOccurred())
				return testCase{
					args: []string{"--category", "envoy.filters.http", "--language", "rust"},
					expectedStdErr: fmt.Sprintf(`Error: output directory must be empty or new: %s

Run 'getenvoy extension init --help' for usage.
`, cwd),
				}
			}),
		)
	})

	Describe("should generate source code when all parameters are valid", func() {
		type testCase struct {
			category string
			language string
		}
		DescribeTable("should generate source code in the output directory",
			func(given testCase) {
				outputDirName, err := ioutil.TempDir("", "test")
				Expect(err).ToNot(HaveOccurred())
				defer func() {
					Expect(os.RemoveAll(outputDirName)).To(Succeed())
				}()

				By("running command")
				c.SetArgs([]string{"extension", "init", "--no-colors", "--category", given.category, "--language", given.language, outputDirName})
				err = cmdutil.Execute(c)
				Expect(err).ToNot(HaveOccurred())

				By("verifying that extension files have been generated")
				outputDir, err := os.Open(outputDirName)
				Expect(err).ToNot(HaveOccurred())
				names, err := outputDir.Readdirnames(-1)
				Expect(err).ToNot(HaveOccurred())
				Expect(names).NotTo(BeEmpty())

				By("verifying that a workspace has been created")
				workspace, err := workspaces.GetWorkspaceAt(outputDirName)
				Expect(err).ToNot(HaveOccurred())

				By("verifying contents of the extension descriptor file")
				descriptor := workspace.GetExtensionDescriptor()
				Expect(descriptor.Category.String()).To(Equal(given.category))
				Expect(descriptor.Language.String()).To(Equal(given.language))

				By("verifying command output")
				Expect(stdout.String()).ToNot(BeEmpty())
				Expect(stderr.String()).To(BeEmpty())
			},
			func() []TableEntry {
				entries := []TableEntry{}
				for _, category := range []string{"envoy.filters.http", "envoy.filters.network", "envoy.access_loggers"} {
					for _, language := range []string{"rust"} {
						entries = append(entries, Entry(fmt.Sprintf("category=%s language=%s", category, language), testCase{
							category: category,
							language: language,
						}))
					}
				}
				return entries
			}()...,
		)
	})
})
