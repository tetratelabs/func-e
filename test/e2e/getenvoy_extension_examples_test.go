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

package e2e_test

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
	e2e "github.com/tetratelabs/getenvoy/test/e2e/util"
)

var _ = Describe("getenvoy extension examples", func() {

	type testCase e2e.CategoryLanguageTuple

	testCases := func() []TableEntry {
		testCases := make([]TableEntry, 0)
		for _, combination := range e2e.GetCategoryLanguageCombinations() {
			testCases = append(testCases, Entry(combination.String(), testCase(combination)))
		}
		return testCases
	}

	const extensionName = "my.extension"

	DescribeTable("should create extension in a new directory",
		func(given testCase) {
			By("choosing the output directory")
			outputDir := filepath.Join(tempDir, "new")
			defer CleanUpExtensionDir(outputDir)

			By("running `extension init` command")
			_, _, err := GetEnvoy("extension init").
				Arg(outputDir).
				Arg("--category").Arg(given.Category.String()).
				Arg("--language").Arg(given.Language.String()).
				Arg("--name").Arg(extensionName).
				Exec()
			Expect(err).NotTo(HaveOccurred())

			By("changing to the output directory")
			err = os.Chdir(outputDir)
			Expect(err).NotTo(HaveOccurred())

			By("running `extension examples list` command")
			stdout, stderr, err := GetEnvoy("extension examples list").Exec()
			Expect(err).NotTo(HaveOccurred())

			By("verifying stdout/stderr")
			Expect(stdout).To(Equal(``))
			Expect(stderr).To(Equal(`Extension has no example setups.

Use "getenvoy extension examples add --help" for more information on how to add one.
`))

			By("running `extension examples add` command")
			stdout, stderr, err = GetEnvoy("extension examples add").Exec()
			Expect(err).NotTo(HaveOccurred())

			By("verifying stdout/stderr")
			var extensionConfigFileName string
			switch given.Language {
			case extension.LanguageTinyGo:
				extensionConfigFileName = "extension.txt"
			default:
				extensionConfigFileName = "extension.json"
			}
			Expect(stdout).To(Equal(``))
			Expect(stderr).To(MatchRegexp(`^\QScaffolding a new example setup:\E\n`))
			Expect(stderr).To(MatchRegexp(`\Q* .getenvoy/extension/examples/default/README.md\E\n`))
			Expect(stderr).To(MatchRegexp(`\Q* .getenvoy/extension/examples/default/envoy.tmpl.yaml\E\n`))
			Expect(stderr).To(MatchRegexp(`\Q* .getenvoy/extension/examples/default/example.yaml\E\n`))
			Expect(stderr).To(MatchRegexp(
				fmt.Sprintf(`\Q* .getenvoy/extension/examples/default/%s\E\n`, extensionConfigFileName)))
			Expect(stderr).To(MatchRegexp(`\QDone!\E\n$`))

			By("verifying output directory")
			Expect(filepath.Join(outputDir, ".getenvoy/extension/examples/default/README.md")).To(BeAnExistingFile())
			Expect(filepath.Join(outputDir, ".getenvoy/extension/examples/default/envoy.tmpl.yaml")).To(BeAnExistingFile())
			Expect(filepath.Join(outputDir, ".getenvoy/extension/examples/default/example.yaml")).To(BeAnExistingFile())
			Expect(filepath.Join(outputDir,
				fmt.Sprintf(".getenvoy/extension/examples/default/%s", extensionConfigFileName))).To(BeAnExistingFile())

			By("running `extension examples list` command")
			stdout, stderr, err = GetEnvoy("extension examples list").Exec()
			Expect(err).NotTo(HaveOccurred())

			By("verifying stdout/stderr")
			Expect(stdout).To(Equal(`EXAMPLE
default
`))
			Expect(stderr).To(BeEmpty())

			By("running `extension examples remove` command")
			stdout, stderr, err = GetEnvoy("extension examples remove --name default").Exec()
			Expect(err).NotTo(HaveOccurred())

			By("verifying stdout/stderr")
			Expect(stdout).To(Equal(``))
			Expect(stderr).To(MatchRegexp(`^\QRemoving example setup:\E\n`))
			Expect(stderr).To(MatchRegexp(`\Q* .getenvoy/extension/examples/default/README.md\E\n`))
			Expect(stderr).To(MatchRegexp(`\Q* .getenvoy/extension/examples/default/envoy.tmpl.yaml\E\n`))
			Expect(stderr).To(MatchRegexp(`\Q* .getenvoy/extension/examples/default/example.yaml\E\n`))
			Expect(stderr).To(MatchRegexp(fmt.Sprintf(
				`\Q* .getenvoy/extension/examples/default/%s\E\n`, extensionConfigFileName)))
			Expect(stderr).To(MatchRegexp(`\QDone!\E\n$`))

			By("running `extension examples list` command")
			stdout, stderr, err = GetEnvoy("extension examples list").Exec()
			Expect(err).NotTo(HaveOccurred())

			By("verifying stdout/stderr")
			Expect(stdout).To(Equal(``))
			Expect(stderr).To(Equal(`Extension has no example setups.

Use "getenvoy extension examples add --help" for more information on how to add one.
`))
		},
		testCases()...,
	)
})
