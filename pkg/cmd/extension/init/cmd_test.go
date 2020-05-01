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
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tetratelabs/getenvoy/pkg/cmd"
)

var _ = Describe("getenvoy extension init", func() {

	Describe("parameters", func() {
		type testCase struct {
			args        []string
			expectedErr string
		}
		type testCaseFn func() testCase
		give := func(given testCase) testCaseFn {
			return func() testCase {
				return given
			}
		}
		DescribeTable("should validate parameters",
			func(givenFn testCaseFn) {
				given := givenFn()

				c := cmd.NewRoot()
				c.SetOut(new(bytes.Buffer))

				c.SetArgs(append([]string{"extension", "init", "--no-prompt"}, given.args...))
				err := c.Execute()

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(given.expectedErr))
			},
			Entry("missing category", give(testCase{
				args:        []string{},
				expectedErr: `"" is not a supported extension category`,
			})),
			Entry("invalid category", give(testCase{
				args:        []string{"--category", "invalid.category"},
				expectedErr: `"invalid.category" is not a supported extension category`,
			})),
			Entry("missing language", give(testCase{
				args:        []string{"--category", "envoy.filters.http"},
				expectedErr: `"" is not a supported programming language`,
			})),
			Entry("invalid language", give(testCase{
				args:        []string{"--category", "envoy.filters.http", "--language", "invalid.language"},
				expectedErr: `"invalid.language" is not a supported programming language`,
			})),
			Entry("non-empty output directory", func() testCase {
				cwd, err := os.Getwd()
				Expect(err).ToNot(HaveOccurred())
				return testCase{
					args:        []string{"--category", "envoy.filters.http", "--language", "rust"},
					expectedErr: fmt.Sprintf(`output directory must be empty or new: %s`, cwd),
				}
			}),
		)
	})

	Describe("generated source code", func() {
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

				c := cmd.NewRoot()
				c.SetOut(new(bytes.Buffer))

				c.SetArgs([]string{"extension", "init", "--no-colors", "--category", given.category, "--language", given.language, outputDirName})
				err = c.Execute()

				Expect(err).ToNot(HaveOccurred())

				outputDir, err := os.Open(outputDirName)
				Expect(err).ToNot(HaveOccurred())
				names, err := outputDir.Readdirnames(-1)
				Expect(err).ToNot(HaveOccurred())
				Expect(names).NotTo(BeEmpty())
			},
			func() []table.TableEntry {
				entries := []table.TableEntry{}
				for _, category := range []string{"envoy.filters.http", "envoy.filters.network", "envoy.access_loggers"} {
					for _, language := range []string{"rust"} {
						entries = append(entries, Entry(fmt.Sprintf("%s + %s", category, language), testCase{
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
