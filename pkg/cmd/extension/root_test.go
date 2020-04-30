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

	"github.com/tetratelabs/getenvoy/pkg/cmd"
	"github.com/tetratelabs/getenvoy/pkg/cmd/extension/globals"
)

var _ = Describe("getenvoy extension", func() {
	Describe("--no-prompt", func() {
		type testCase struct {
			args     []string
			expected bool
		}
		DescribeTable("should be available as `globals.NoPrompt` variable",
			func(given testCase) {
				c := cmd.NewRoot()
				c.SetOut(new(bytes.Buffer))

				c.SetArgs(append([]string{"extension"}, given.args...))
				Expect(c.Execute()).To(Succeed())

				Expect(globals.NoPrompt).To(Equal(given.expected))
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
				c := cmd.NewRoot()
				c.SetOut(new(bytes.Buffer))

				c.SetArgs(append([]string{"extension"}, given.args...))
				Expect(c.Execute()).To(Succeed())

				Expect(globals.NoColors).To(Equal(given.expected))
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
