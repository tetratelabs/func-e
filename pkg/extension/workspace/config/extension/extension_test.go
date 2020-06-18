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
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config"
)

var _ = Describe("Extension", func() {
	Describe("Unmarshal()", func() {
		Describe("in case of invalid input", func() {
			type testCase struct {
				input       string
				expectedErr string
			}
			DescribeTable("should fail with a proper error",
				func(given testCase) {
					var descriptor Descriptor
					err := config.Unmarshal([]byte(given.input), &descriptor)
					Expect(err).To(MatchError(given.expectedErr))
				},
				Entry("invalid programming language", testCase{
					input: `#
# Envoy Wasm extension created with getenvoy toolkit.
#
kind: Extension

name: mycompany.filters.http.custom_metrics

language: invalid
category: envoy.filters.http
`,
					expectedErr: `error unmarshaling JSON: "invalid" is not a valid programming language`,
				}),
				Entry("invalid extension category", testCase{
					input: `#
# Envoy Wasm extension created with getenvoy toolkit.
#
kind: Extension

name: mycompany.filters.http.custom_metrics

language: rust
category: invalid
`,
					expectedErr: `error unmarshaling JSON: "invalid" is not a valid extension category`,
				}),
			)
		})
	})

	//nolint:lll
	Describe("Validate()", func() {
		Describe("in case of invalid input", func() {
			type testCase struct {
				input       string
				expectedErr string
			}
			DescribeTable("should fail with a proper error",
				func(given testCase) {
					var descriptor Descriptor
					err := config.Unmarshal([]byte(given.input), &descriptor)
					Expect(err).ToNot(HaveOccurred())

					err = descriptor.Validate()
					Expect(err).To(MatchError(given.expectedErr))
				},
				Entry("empty", testCase{
					input:       ``,
					expectedErr: `4 errors occurred: extension name cannot be empty; extension category cannot be empty; programming language cannot be empty; runtime description is not valid: Envoy version cannot be empty`,
				}),
				Entry("invalid Envoy version", testCase{
					input: `#
# Envoy Wasm extension created with getenvoy toolkit.
#
kind: Extension

name: mycompany.filters.http.custom_metrics

language: rust
category: envoy.filters.http

# Runtime the extension is being developed against.
runtime:
  envoy:
    version: invalid value
`,
					expectedErr: `runtime description is not valid: Envoy version is not valid: "invalid value" is not a valid GetEnvoy reference. Expected format: <flavor>:<version>[/<platform>]`,
				}),
				Entry("missing extension name", testCase{
					input: `#
# Envoy Wasm extension created with getenvoy toolkit.
#
kind: Extension

category: envoy.filters.http
language: rust

# Runtime the extension is being developed against.
runtime:
  envoy:
    version: wasm:nightly
`,
					expectedErr: `extension name cannot be empty`,
				}),
				Entry("invalid extension name", testCase{
					input: `#
# Envoy Wasm extension created with getenvoy toolkit.
#
kind: Extension

name: ?!@#$%

category: envoy.filters.http
language: rust

# Runtime the extension is being developed against.
runtime:
  envoy:
    version: wasm:nightly
`,
					expectedErr: `"?!@#$%" is not a valid extension name. Extension name must match the format "^[a-z0-9_]+(\\.[a-z0-9_]+)*$". E.g., 'mycompany.filters.http.custom_metrics'`,
				}),
			)
		})
		Describe("in case of valid input", func() {
			type testCase struct {
				input string
			}
			DescribeTable("should not return any error",
				func(given testCase) {
					var descriptor Descriptor
					err := config.Unmarshal([]byte(given.input), &descriptor)
					Expect(err).ToNot(HaveOccurred())

					err = descriptor.Validate()
					Expect(err).ToNot(HaveOccurred())

					actual, err := config.Marshal(&descriptor)
					Expect(err).ToNot(HaveOccurred())
					Expect(actual).To(MatchYAML(given.input))
				},
				Entry("invalid Envoy version", testCase{
					input: `#
# Envoy Wasm extension created with getenvoy toolkit.
#
kind: Extension

name: mycompany.filters.http.custom_metrics

category: envoy.filters.http
language: rust

# Runtime the extension is being developed against.
runtime:
  envoy:
    version: wasm:nightly
`,
				}),
			)
		})
	})
})

var _ = Describe("ValidateExtensionName()", func() {
	DescribeTable("should accept valid names",
		func(given string) {
			err := ValidateExtensionName(given)
			Expect(err).ToNot(HaveOccurred())
		},
		Entry("Envoy-like name", "mycompany.filters.http.custom_metrics"),
		Entry("no segments", "myextension"),
		Entry("numbers", "911.i18n.v2"),
		Entry("'_'", "_._"),
	)
	//nolint:lll
	DescribeTable("should reject invalid names",
		func(given string) {
			err := ValidateExtensionName(given)
			Expect(err).To(MatchError(fmt.Sprintf(`%q is not a valid extension name. Extension name must match the format "^[a-z0-9_]+(\\.[a-z0-9_]+)*$". E.g., 'mycompany.filters.http.custom_metrics'`, given)))
		},
		Entry("trailing '.'", "myextension."),
		Entry("upper-case", "MYEXTENSION"),
		Entry("'-'", "-.-"),
		Entry("non alpha-num characters", `!@#$%^&*()-+<>?~:;"'\[]{}`),
	)
})

var _ = Describe("SanitizeExtensionName()", func() {
	type testCase struct {
		input    []string
		expected string
	}
	DescribeTable("should replace unsafe characters",
		func(given testCase) {
			actual := SanitizeExtensionName(given.input...)
			Expect(actual).To(Equal(given.expected))
			Expect(ValidateExtensionName(actual)).To(Succeed())
		},
		Entry("upper-case, non-alpha-num and empty", testCase{
			input:    []string{`My-C0mpany.com`, ``, `e!x@t#`},
			expected: `my_c0mpany_com.e_x_t_`,
		}),
	)
})

var _ = Describe("SanitizeExtensionNameSegment()", func() {
	type testCase struct {
		input    string
		expected string
	}
	DescribeTable("should replace unsafe characters",
		func(given testCase) {
			actual := SanitizeExtensionNameSegment(given.input)
			Expect(actual).To(Equal(given.expected))
		},
		Entry("upper-case", testCase{
			input:    `My-C0mpany.com`,
			expected: `my_c0mpany_com`,
		}),
		Entry("non alpha-num characters", testCase{
			input:    `!@#$%^&*()-+<>?~:;"'\[]{}`,
			expected: `_________________________`,
		}),
	)
})
