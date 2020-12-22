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

	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	e2e "github.com/tetratelabs/getenvoy/test/e2e/util"
)

var _ = Describe("getenvoy extension init", func() {

	type testCase e2e.CategoryLanguageTuple

	testCases := func() []TableEntry {
		testCases := make([]TableEntry, 0)
		for _, combination := range e2e.GetCategoryLanguageCombinations() {
			testCases = append(testCases, Entry(combination.String(), testCase(combination)))
		}
		return testCases
	}

	const extensionName = "my.extension"

	VerifyStdoutStderr := func(stdout string, stderr string, outputDir string) {
		Expect(stdout).To(Equal(``))
		Expect(stderr).To(MatchRegexp(`^\QScaffolding a new extension:\E\n`))
		Expect(stderr).To(MatchRegexp(`\QGenerating files in %s:\E\n`, outputDir))
		Expect(stderr).To(MatchRegexp(`\Q* .getenvoy/extension/extension.yaml\E\n`))
		Expect(stderr).To(MatchRegexp(`\QDone!\E\n$`))
	}

	VerifyOutputDir := func(given testCase, outputDir string) {
		workspace, err := workspaces.GetWorkspaceAt(outputDir)
		Expect(err).NotTo(HaveOccurred())
		Expect(workspace.GetExtensionDescriptor().Name).To(Equal(extensionName))
		Expect(workspace.GetExtensionDescriptor().Category).To(Equal(given.Category))
		Expect(workspace.GetExtensionDescriptor().Language).To(Equal(given.Language))
	}

	DescribeTable("should create extension in a new directory",
		func(given testCase) {
			By("choosing the output directory")
			outputDir := filepath.Join(tempDir, "new")

			By("running `extension init` command")
			stdout, stderr, err := GetEnvoy("extension init").
				Arg(outputDir).
				Arg("--category").Arg(given.Category.String()).
				Arg("--language").Arg(given.Language.String()).
				Arg("--name").Arg(extensionName).
				Exec()
			Expect(err).NotTo(HaveOccurred())

			By("verifying stdout/stderr")
			VerifyStdoutStderr(stdout, stderr, outputDir)

			By("verifying output directory")
			VerifyOutputDir(given, outputDir)
		},
		testCases()...,
	)

	DescribeTable("should create extension in the current directory",
		func(given testCase) {
			By("choosing the output directory")
			outputDir := tempDir
			defer CleanUpExtensionDir(outputDir)

			By("changing to the output directory")
			err := os.Chdir(outputDir)
			Expect(err).NotTo(HaveOccurred())

			By("running `extension init` command")
			stdout, stderr, err := GetEnvoy("extension init").
				Arg("--category").Arg(given.Category.String()).
				Arg("--language").Arg(given.Language.String()).
				Arg("--name").Arg(extensionName).
				Exec()
			Expect(err).NotTo(HaveOccurred())

			By("verifying stdout/stderr")
			VerifyStdoutStderr(stdout, stderr, outputDir)

			By("verifying output directory")
			VerifyOutputDir(given, outputDir)
		},
		testCases()...,
	)
})
