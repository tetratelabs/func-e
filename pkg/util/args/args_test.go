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

package args_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/tetratelabs/getenvoy/pkg/util/args"
)

var _ = Describe("SplitCommandLine()", func() {
	type testCase struct {
		input    []string
		expected []string
	}
	DescribeTable("should split command line properly",
		func(given testCase) {
			Expect(SplitCommandLine(given.input...)).To(Equal(given.expected))
		},
		Entry("nil", testCase{
			input:    nil,
			expected: []string{},
		}),
		Entry("empty", testCase{
			input:    []string{},
			expected: []string{},
		}),
		Entry("already split", testCase{
			input:    []string{"-e", "VAR=VALUE"},
			expected: []string{"-e", "VAR=VALUE"},
		}),
		Entry("command line", testCase{
			input: []string{"-e VAR=VALUE -v /host/path:/container/path"},
			expected: []string{
				"-e",
				"VAR=VALUE",
				"-v",
				"/host/path:/container/path",
			},
		}),
		Entry("quoted command line", testCase{
			input: []string{`'-e VAR=VALUE' "-v /host/path:/container/path"`},
			expected: []string{
				"-e VAR=VALUE",
				"-v /host/path:/container/path",
			},
		}),
	)
})
