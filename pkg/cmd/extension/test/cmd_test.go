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

package test_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/tetratelabs/getenvoy/pkg/cmd"

	testcontext "github.com/tetratelabs/getenvoy/pkg/test/cmd/extension"

	cmdutil "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

var _ = Describe("getenvoy extension test", func() {

	var dockerDir string

	BeforeEach(func() {
		dir, err := filepath.Abs("../../../extension/workspace/toolchain/builtin/testdata/toolchain")
		Expect(err).ToNot(HaveOccurred())
		dockerDir = dir
	})

	var pathBackup string

	BeforeEach(func() {
		pathBackup = os.Getenv("PATH")

		// override PATH to overshadow `docker` executable during the test
		path := strings.Join([]string{dockerDir, pathBackup}, string(filepath.ListSeparator))
		os.Setenv("PATH", path)
	})

	AfterEach(func() {
		os.Setenv("PATH", pathBackup)
	})

	var cwdBackup string

	BeforeEach(func() {
		cwd, err := os.Getwd()
		Expect(err).ToNot(HaveOccurred())
		cwdBackup = cwd
	})

	AfterEach(func() {
		if cwdBackup != "" {
			Expect(os.Chdir(cwdBackup)).To(Succeed())
		}
	})

	testcontext.SetDefaultUser() // UID:GID == 1001:1002

	var stdout *bytes.Buffer
	var stderr *bytes.Buffer

	BeforeEach(func() {
		stdout = new(bytes.Buffer)
		stderr = new(bytes.Buffer)
	})

	var c *cobra.Command

	BeforeEach(func() {
		c = cmd.NewRoot()
		c.SetOut(stdout)
		c.SetErr(stderr)
	})

	It("should validate value of --toolchain-container-image flag", func() {
		By("running command")
		c.SetArgs([]string{"extension", "test", "--toolchain-container-image", "?invalid value?"})
		err := cmdutil.Execute(c)
		Expect(err).To(HaveOccurred())

		By("verifying command output")
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(Equal(`Error: "?invalid value?" is not a valid image name: invalid reference format

Run 'getenvoy extension test --help' for usage.
`))
	})

	It("should validate value of --toolchain-container-options flag", func() {
		By("running command")
		c.SetArgs([]string{"extension", "test", "--toolchain-container-options", "imbalanced ' quotes"})
		err := cmdutil.Execute(c)
		Expect(err).To(HaveOccurred())

		By("verifying command output")
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(Equal(`Error: "imbalanced ' quotes" is not a valid command line string

Run 'getenvoy extension test --help' for usage.
`))
	})

	chdir := func(path string) string {
		dir, err := filepath.Abs(path)
		Expect(err).ToNot(HaveOccurred())

		err = os.Chdir(dir)
		Expect(err).ToNot(HaveOccurred())

		return dir
	}

	//nolint:lll
	Context("inside a workspace directory", func() {
		It("should succeed", func() {
			By("changing to a workspace dir")
			workspaceDir := chdir("../build/testdata/workspace")

			By("running command")
			c.SetArgs([]string{"extension", "test"})
			err := cmdutil.Execute(c)
			Expect(err).ToNot(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(Equal(fmt.Sprintf("%s/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init getenvoy/extension-rust-builder:latest test\n", dockerDir, workspaceDir)))
			Expect(stderr.String()).To(Equal("docker stderr\n"))
		})

		It("should allow to override build image and add Docker cli options", func() {
			By("changing to a workspace dir")
			workspaceDir := chdir("../build/testdata/workspace")

			By("running command")
			c.SetArgs([]string{"extension", "test",
				"--toolchain-container-image", "build/image",
				"--toolchain-container-options", `-e 'VAR=VALUE' -v "/host:/container"`,
			})
			err := cmdutil.Execute(c)
			Expect(err).ToNot(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(Equal(fmt.Sprintf("%s/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init -e VAR=VALUE -v /host:/container build/image test\n", dockerDir, workspaceDir)))
			Expect(stderr.String()).To(Equal("docker stderr\n"))
		})

		It("should properly handle Docker build failing", func() {
			By("changing to a workspace dir")
			workspaceDir := chdir("../build/testdata/workspace")

			By("running command")
			c.SetArgs([]string{"extension", "test",
				"--toolchain-container-image", "build/image",
				"--toolchain-container-options", `-e EXIT_CODE=3`,
			})
			err := cmdutil.Execute(c)
			Expect(err).To(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(Equal(fmt.Sprintf("%s/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init -e EXIT_CODE=3 build/image test\n", dockerDir, workspaceDir)))
			Expect(stderr.String()).To(Equal(fmt.Sprintf(`docker stderr
Error: failed to unit test Envoy extension using "default" toolchain: failed to execute an external command "%s/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init -e EXIT_CODE=3 build/image test": exit status 3

Run 'getenvoy extension test --help' for usage.
`, dockerDir, workspaceDir)))
		})
	})

	Context("outside of a workspace directory", func() {
		It("should fail", func() {
			By("changing to a non-workspace dir")
			dir := chdir("../build/testdata")

			By("running command")
			c.SetArgs([]string{"extension", "test"})
			err := cmdutil.Execute(c)
			Expect(err).To(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(BeEmpty())
			Expect(stderr.String()).To(Equal(fmt.Sprintf(`Error: there is no extension directory at or above: %s

Run 'getenvoy extension test --help' for usage.
`, dir)))
		})
	})
})
