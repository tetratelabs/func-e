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
					expectedErr: `3 errors occurred: extension category cannot be empty; programming language cannot be empty; runtime description is not valid: Envoy version cannot be empty`,
				}),
				Entry("invalid Envoy version", testCase{
					input: `#
# Envoy Wasm extension created with getenvoy toolkit.
#
kind: Extension

language: rust
category: envoy.filters.http

# Runtime the extension is being developed against.
runtime:
  envoy:
    version: invalid value
`,
					expectedErr: `runtime description is not valid: Envoy version is not valid: "invalid value" is not a valid GetEnvoy reference. Expected format: <flavor>:<version>[/<platform>]`,
				}),
			)
		})
		Describe("in case of valid input", func() {
			type testCase struct {
				input    string
				expected Descriptor
			}
			DescribeTable("should not return any error",
				func(given testCase) {
					var descriptor Descriptor
					err := config.Unmarshal([]byte(given.input), &descriptor)
					Expect(err).ToNot(HaveOccurred())

					err = descriptor.Validate()
					Expect(err).ToNot(HaveOccurred())

					Expect(descriptor).To(Equal(given.expected))
				},
				Entry("invalid Envoy version", testCase{
					input: `#
# Envoy Wasm extension created with getenvoy toolkit.
#
kind: Extension

language: rust
category: envoy.filters.http

# Runtime the extension is being developed against.
runtime:
  envoy:
    version: wasm:nightly
`,
					expected: Descriptor{
						Meta: config.Meta{
							Kind: Kind,
						},
						Language: LanguageRust,
						Category: EnvoyHTTPFilter,
						Runtime: Runtime{
							Envoy: EnvoyRuntime{
								Version: "wasm:nightly",
							},
						},
					},
				}),
			)
		})
	})
})
