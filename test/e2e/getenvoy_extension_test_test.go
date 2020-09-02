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
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	e2e "github.com/tetratelabs/getenvoy/test/e2e/util"
)

var _ = Describe("getenvoy extension test", func() {

	type testCase e2e.CategoryLanguageTuple

	testCases := func() []TableEntry {
		testCases := make([]TableEntry, 0)
		for _, combination := range e2e.GetCategoryLanguageCombinations() {
			testCases = append(testCases, Entry(combination.String(), testCase(combination)))
		}
		return testCases
	}

	const extensionName = "my.extension"

	DescribeTable("should run unit tests",
		func(given testCase) {
			By("choosing the output directory")
			outputDir := filepath.Join(tempDir, "new")

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

			By("running `extension test` command")
			stdout, stderr, err := GetEnvoy("extension test").
				Args(e2e.Env.GetBuiltinContainerOptions()...).
				Exec()
			Expect(err).NotTo(HaveOccurred())

			By("verifying stdout/stderr")
			// apparently, use of `-t` option in `docker run` causes stderr to be incorporated into stdout
			Expect(stdout).NotTo(BeEmpty())
			Expect(stderr).To(BeEmpty())
		},
		testCases()...,
	)
})
