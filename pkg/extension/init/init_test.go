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
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
)

var _ = Describe("interpolate()", func() {
	type testCase struct {
		extension   *extension.Descriptor
		fileName    string
		fileContent string
		expected    string
	}
	DescribeTable("should interpolate extension name",
		func(given testCase) {
			actual, err := interpolate(given.extension)(given.fileName, []byte(given.fileContent))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(actual)).To(Equal(given.expected))
		},
		Entry("src/factory.rs", testCase{
			extension: &extension.Descriptor{
				Name: "my_company.my_extension",
			},
			fileName: "src/factory.rs",
			fileContent: `
impl<'a> ExtensionFactory for SampleHttpFilterFactory<'a> {
    type Extension = SampleHttpFilter<'a>;

    /// The reference name for Sample HTTP Filter.
    ///
    /// This name appears in "Envoy" configuration as a value of "root_id" field
    /// (also known as "group_name").
	const NAME: &'static str = "{{ .Extension.Name }}";
}
`,
			expected: `
impl<'a> ExtensionFactory for SampleHttpFilterFactory<'a> {
    type Extension = SampleHttpFilter<'a>;

    /// The reference name for Sample HTTP Filter.
    ///
    /// This name appears in "Envoy" configuration as a value of "root_id" field
    /// (also known as "group_name").
	const NAME: &'static str = "my_company.my_extension";
}
`,
		}),
	)
})

var _ = Describe("Scaffold()", func() {
	var tmpDir string

	BeforeEach(func() {
		dir, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		tmpDir = dir
	})

	AfterEach(func() {
		if tmpDir != "" {
			Expect(os.RemoveAll(tmpDir)).To(Succeed())
		}
	})

	type testCase struct {
		extension *extension.Descriptor
		file      string
		expected  string
	}
	DescribeTable("should interpolate extension name",
		func(given testCase) {
			opts := &ScaffoldOpts{
				Extension:    given.extension,
				TemplateName: "default",
				OutputDir:    tmpDir,
			}

			err := Scaffold(opts)
			Expect(err).NotTo(HaveOccurred())

			actual, err := ioutil.ReadFile(filepath.Join(opts.OutputDir, given.file))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(actual)).To(ContainSubstring(given.expected))
		},
		Entry("rust/filters/http", testCase{
			extension: &extension.Descriptor{
				Name:     "my_company.my_extension",
				Category: extension.EnvoyHTTPFilter,
				Language: extension.LanguageRust,
			},
			file:     "src/factory.rs",
			expected: `const NAME: &'static str = "my_company.my_extension";`,
		}),
		Entry("rust/filters/network", testCase{
			extension: &extension.Descriptor{
				Name:     "my_company.my_extension",
				Category: extension.EnvoyNetworkFilter,
				Language: extension.LanguageRust,
			},
			file:     "src/factory.rs",
			expected: `const NAME: &'static str = "my_company.my_extension";`,
		}),
		// TODO(yskopets): add a test case for `envoy.access_logger
	)
})
