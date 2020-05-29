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

package extension_test

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/tetratelabs/getenvoy/pkg/cmd"
	"github.com/tetratelabs/getenvoy/pkg/cmd/extension/globals"

	cmdutil "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

var _ = Describe("getenvoy extension", func() {

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

	Describe("--no-prompt", func() {
		type testCase struct {
			args     []string
			expected bool
		}
		DescribeTable("should be available as `globals.NoPrompt` variable",
			func(given testCase) {
				By("running command")
				c.SetArgs(append([]string{"extension"}, given.args...))
				err := cmdutil.Execute(c)
				Expect(err).ToNot(HaveOccurred())

				By("verifying side effects")
				Expect(globals.NoPrompt).To(Equal(given.expected))

				By("verifying command output")
				Expect(stdout.String()).ToNot(BeEmpty())
				Expect(stderr.String()).To(BeEmpty())
			},
			Entry("--no-prompt", testCase{
				args:     []string{"--no-prompt"},
				expected: true,
			}),
			Entry("--no-prompt=false", testCase{
				args:     []string{"--no-prompt=false"},
				expected: false,
			}),
		)
	})
	Describe("--no-colors", func() {
		type testCase struct {
			args     []string
			expected bool
		}
		DescribeTable("should be available as `globals.NoColors` variable",
			func(given testCase) {
				By("running command")
				c.SetArgs(append([]string{"extension"}, given.args...))
				err := cmdutil.Execute(c)
				Expect(err).ToNot(HaveOccurred())

				By("verifying side effects")
				Expect(globals.NoColors).To(Equal(given.expected))

				By("verifying command output")
				Expect(stdout.String()).ToNot(BeEmpty())
				Expect(stderr.String()).To(BeEmpty())
			},
			Entry("--no-colors", testCase{
				args:     []string{"--no-colors"},
				expected: true,
			}),
			Entry("--no-colors=false", testCase{
				args:     []string{"--no-colors=false"},
				expected: false,
			}),
		)
	})
})
