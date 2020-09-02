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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/tetratelabs/getenvoy/pkg/extension/init"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
)

var _ = Describe("GenerateExtensionName()", func() {
	type testCase struct {
		category  extension.Category
		outputDir string
		expected  string
	}
	DescribeTable("should generate a valid extension name",
		func(given testCase) {
			actual := GenerateExtensionName(given.category, given.outputDir)
			Expect(actual).To(Equal(given.expected))
		},
		Entry("http", testCase{
			category:  extension.EnvoyHTTPFilter,
			outputDir: `My-Ext3nsi0n.com`,
			expected:  `mycompany.filters.http.my_ext3nsi0n_com`,
		}),
		Entry("network", testCase{
			category:  extension.EnvoyNetworkFilter,
			outputDir: `My-Ext3nsi0n.com`,
			expected:  `mycompany.filters.network.my_ext3nsi0n_com`,
		}),
		Entry("access_logger", testCase{
			category:  extension.EnvoyAccessLogger,
			outputDir: `My-Ext3nsi0n.com`,
			expected:  `mycompany.access_loggers.my_ext3nsi0n_com`,
		}),
	)

	DescribeTable("should panic on unknown categories",
		func(given testCase) {
			Expect(func() {
				_ = GenerateExtensionName(given.category, given.outputDir)
			}).ToNot(Panic())
		},
		func() []TableEntry {
			testCases := make([]TableEntry, len(extension.Categories))
			for i, category := range extension.Categories {
				testCases[i] = Entry(category.String(), testCase{
					category:  category,
					outputDir: `my-extension`,
				})
			}
			return testCases
		}()...,
	)
})
