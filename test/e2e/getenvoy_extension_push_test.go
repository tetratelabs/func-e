package e2e_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	e2e "github.com/tetratelabs/getenvoy/test/e2e/util"

	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	toolchains "github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain"
)

const (
	localRegistryWasmImageRef = "localhost:5000/getenvoy/sample"
)

var _ = Describe("getenvoy extension push", func() {

	type testCase e2e.CategoryLanguageTuple

	testCases := func() []TableEntry {
		testCases := make([]TableEntry, 0)
		for _, combination := range e2e.GetCategoryLanguageCombinations() {
			testCases = append(testCases, Entry(combination.String(), testCase(combination)))
		}
		return testCases
	}

	//TODO(musaprg): write teardown process for local registries if it's needed

	const extensionName = "my.extension"

	DescribeTable("should push a *.wasm file",
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

			By("running `extension build` command")
			stdout, stderr, err := GetEnvoy("extension build").
				Args(e2e.Env.GetBuiltinContainerOptions()...).
				Exec()
			Expect(err).NotTo(HaveOccurred())

			By("verifying stdout/stderr")
			// apparently, use of `-t` option in `docker run` causes stderr to be incorporated into stdout
			Expect(stdout).NotTo(BeEmpty())
			Expect(stderr).To(BeEmpty())

			By("verifying *.wasm file")
			workspace, err := workspaces.GetWorkspaceAt(outputDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(workspace).NotTo(BeNil())
			toolchain, err := toolchains.LoadToolchain(toolchains.Default, workspace)
			Expect(err).NotTo(HaveOccurred())
			Expect(toolchain).NotTo(BeNil())

			By("running `extension push` command")
			_, _, err = GetEnvoy("extension push").Arg(localRegistryWasmImageRef).Exec()
			Expect(stdout).NotTo(BeEmpty())
			Expect(stderr).To(BeEmpty())
		},
		testCases()...,
	)
})
