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

package types_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/tetratelabs/getenvoy/pkg/types"
)

var _ = Describe("ParseReference()", func() {
	Describe("valid input", func() {
		type testCase struct {
			input    string
			expected Reference
		}
		DescribeTable("should parse valid input",
			func(given testCase) {
				ref, err := ParseReference(given.input)

				Expect(err).ToNot(HaveOccurred())
				Expect(*ref).To(Equal(given.expected))
			},
			Entry("standard:1.11.0", testCase{
				input:    `standard:1.11.0`,
				expected: Reference{Flavor: "standard", Version: "1.11.0", Platform: ""},
			}),
			Entry("standard:1.11.0/darwin", testCase{
				input:    `standard:1.11.0/darwin`,
				expected: Reference{Flavor: "standard", Version: "1.11.0", Platform: "darwin"},
			}),
			Entry("standard:1.11.0/linux-glibc", testCase{
				input:    `standard:1.11.0/linux-glibc`,
				expected: Reference{Flavor: "standard", Version: "1.11.0", Platform: "linux-glibc"},
			}),
			Entry("wasm:1.15", testCase{
				input:    `wasm:1.15`,
				expected: Reference{Flavor: "wasm", Version: "1.15", Platform: ""},
			}),
			Entry("wasm:1.15/darwin", testCase{
				input:    `wasm:1.15/darwin`,
				expected: Reference{Flavor: "wasm", Version: "1.15", Platform: "darwin"},
			}),
			Entry("wasm:1.15/linux-glibc", testCase{
				input:    `wasm:1.15/linux-glib`,
				expected: Reference{Flavor: "wasm", Version: "1.15", Platform: "linux-glib"},
			}),
			Entry("mixed case", testCase{
				input:    `Wasm:NightlY/LINUX-GLIBC`,
				expected: Reference{Flavor: "wasm", Version: "nightly", Platform: "linux-glibc"},
			}),
			Entry("trailing slash", testCase{
				input:    `standard:1.11.0/`,
				expected: Reference{Flavor: "standard", Version: "1.11.0", Platform: ""},
			}),
			Entry("special characters", testCase{
				input:    `abcd-EFGH.01234_:-56789.XYZ_/`,
				expected: Reference{Flavor: "abcd-efgh.01234_", Version: "-56789.xyz_", Platform: ""},
			}),
		)
	})
	Describe("invalid input", func() {
		type testCase struct {
			input       string
			expectedErr string
		}
		DescribeTable("should fail on invalid input",
			func(given testCase) {
				ref, err := ParseReference(given.input)

				Expect(err).To(MatchError(given.expectedErr))
				Expect(ref).To(BeNil())
			},
			Entry("empty string", testCase{
				input:       ``,
				expectedErr: `"" is not a valid GetEnvoy reference. Expected format: <flavor>:<version>[/<platform>]`,
			}),
			Entry("empty components", testCase{
				input:       `:/`,
				expectedErr: `":/" is not a valid GetEnvoy reference. Expected format: <flavor>:<version>[/<platform>]`,
			}),
			Entry("no flavor", testCase{
				input:       `:1.11.0/darwin`,
				expectedErr: `":1.11.0/darwin" is not a valid GetEnvoy reference. Expected format: <flavor>:<version>[/<platform>]`,
			}),
			Entry("no version", testCase{
				input:       `standard:/darwin`,
				expectedErr: `"standard:/darwin" is not a valid GetEnvoy reference. Expected format: <flavor>:<version>[/<platform>]`,
			}),
			Entry("invalid character in flavor", testCase{
				input:       `stan dard:1.11.0/darwin`,
				expectedErr: `"stan dard:1.11.0/darwin" is not a valid GetEnvoy reference. Expected format: <flavor>:<version>[/<platform>]`,
			}),
		)
	})
})
var _ = Describe("Reference", func() {
	Describe("ToString()", func() {
		type testCase struct {
			input    string
			expected string
		}
		DescribeTable("should parse valid input",
			func(given testCase) {
				ref, err := ParseReference(given.input)
				Expect(err).ToNot(HaveOccurred())

				actual := ref.String()
				Expect(actual).To(Equal(given.expected))

				ref2, err := ParseReference(actual)
				Expect(err).ToNot(HaveOccurred())
				Expect(ref2).To(Equal(ref))
			},
			Entry("standard:1.11.0", testCase{
				input:    `standard:1.11.0`,
				expected: `standard:1.11.0`,
			}),
			Entry("standard:1.11.0/darwin", testCase{
				input:    `standard:1.11.0/darwin`,
				expected: `standard:1.11.0/darwin`,
			}),
			Entry("standard:1.11.0/", testCase{
				input:    `standard:1.11.0/`,
				expected: `standard:1.11.0`,
			}),
		)
	})
})
