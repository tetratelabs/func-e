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

package example_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/otiai10/copy"

	"github.com/spf13/cobra"

	"github.com/tetratelabs/getenvoy/pkg/cmd"

	cmdutil "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

var _ = Describe("getenvoy extension examples add", func() {

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

	var tempDir string

	BeforeEach(func() {
		dir, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		tempDir = dir
	})

	AfterEach(func() {
		if tempDir != "" {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		}
	})

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

	//nolint:lll
	It("should validate --name flag", func() {
		By("running command")
		c.SetArgs([]string{"extension", "examples", "add", "--name", "my:example"})
		err := cmdutil.Execute(c)
		Expect(err).To(HaveOccurred())

		By("verifying command output")
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(Equal(`Error: "my:example" is not a valid example name. Example name must match the format "^[a-z0-9._-]+$". E.g., 'my.example', 'my-example' or 'my_example'

Run 'getenvoy extension examples add --help' for usage.
`))
	})

	chdir := func(path string) string {
		dir, err := filepath.Abs(path)
		Expect(err).ToNot(HaveOccurred())

		dir, err = filepath.EvalSymlinks(dir)
		Expect(err).ToNot(HaveOccurred())

		err = os.Chdir(dir)
		Expect(err).ToNot(HaveOccurred())

		return dir
	}

	//nolint:lll
	Context("inside a workspace directory", func() {
		It("should create 'default' example setup when no --name is omitted", func() {
			By("simulating a workspace without any examples")
			err := copy.Copy("testdata/workspace1", tempDir)
			Expect(err).NotTo(HaveOccurred())

			By("changing to a workspace dir")
			chdir(tempDir)

			By("running command")
			c.SetArgs([]string{"extension", "examples", "add"})
			err = cmdutil.Execute(c)
			Expect(err).ToNot(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(BeEmpty())
			Expect(stderr.String()).To(Equal(`Scaffolding a new example setup:
* .getenvoy/extension/examples/default/README.md
* .getenvoy/extension/examples/default/envoy.tmpl.yaml
* .getenvoy/extension/examples/default/example.yaml
* .getenvoy/extension/examples/default/extension.json
Done!
`))
			By("verifying file system")
			readmePath := filepath.Join(tempDir, ".getenvoy/extension/examples/default/README.md")
			Expect(readmePath).To(BeAnExistingFile())
			Expect(filepath.Join(tempDir, ".getenvoy/extension/examples/default/envoy.tmpl.yaml")).To(BeAnExistingFile())
			Expect(filepath.Join(tempDir, ".getenvoy/extension/examples/default/example.yaml")).To(BeAnExistingFile())
			Expect(filepath.Join(tempDir, ".getenvoy/extension/examples/default/extension.json")).To(BeAnExistingFile())

			// Check README substitution: EXTENSION_CONFIG_FILE_NAME must be replaced with "extension.json".
			data, err := ioutil.ReadFile(readmePath)
			Expect(err).ToNot(HaveOccurred())
			readme := string(data)
			Expect(readme).To(ContainSubstring("extension.json"))
			Expect(readme).To(Not(ContainSubstring("EXTENSION_CONFIG_FILE_NAME")))
		})

		It("should create example setup with a given --name", func() {
			By("simulating a workspace without any examples")
			err := copy.Copy("testdata/workspace1", tempDir)
			Expect(err).NotTo(HaveOccurred())

			By("changing to a workspace dir")
			chdir(tempDir)

			By("running command")
			c.SetArgs([]string{"extension", "examples", "add", "--name", "advanced"})
			err = cmdutil.Execute(c)
			Expect(err).ToNot(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(BeEmpty())
			Expect(stderr.String()).To(Equal(`Scaffolding a new example setup:
* .getenvoy/extension/examples/advanced/README.md
* .getenvoy/extension/examples/advanced/envoy.tmpl.yaml
* .getenvoy/extension/examples/advanced/example.yaml
* .getenvoy/extension/examples/advanced/extension.json
Done!
`))
			By("verifying file system")
			Expect(filepath.Join(tempDir, ".getenvoy/extension/examples/advanced/README.md")).To(BeAnExistingFile())
			Expect(filepath.Join(tempDir, ".getenvoy/extension/examples/advanced/envoy.tmpl.yaml")).To(BeAnExistingFile())
			Expect(filepath.Join(tempDir, ".getenvoy/extension/examples/advanced/example.yaml")).To(BeAnExistingFile())
			Expect(filepath.Join(tempDir, ".getenvoy/extension/examples/advanced/extension.json")).To(BeAnExistingFile())
		})

		It("should fail if such example already exists", func() {
			By("simulating a workspace with the 'default' example")
			err := copy.Copy("testdata/workspace2", tempDir)
			Expect(err).NotTo(HaveOccurred())

			By("changing to a workspace dir")
			chdir(tempDir)

			By("running command")
			c.SetArgs([]string{"extension", "examples", "add"})
			err = cmdutil.Execute(c)
			Expect(err).To(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(BeEmpty())
			Expect(stderr.String()).To(Equal(`Error: example setup "default" already exists

Run 'getenvoy extension examples add --help' for usage.
`))
		})

		It("should create 'default' example setup when no --name is omitted for TinyGo", func() {
			By("simulating a workspace without any examples")
			err := copy.Copy("testdata/workspace4", tempDir)
			Expect(err).NotTo(HaveOccurred())

			By("changing to a workspace dir")
			chdir(tempDir)

			By("running command")
			c.SetArgs([]string{"extension", "examples", "add"})
			err = cmdutil.Execute(c)
			Expect(err).ToNot(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(BeEmpty())
			Expect(stderr.String()).To(Equal(`Scaffolding a new example setup:
* .getenvoy/extension/examples/default/README.md
* .getenvoy/extension/examples/default/envoy.tmpl.yaml
* .getenvoy/extension/examples/default/example.yaml
* .getenvoy/extension/examples/default/extension.txt
Done!
`))
			By("verifying file system")
			readmePath := filepath.Join(tempDir, ".getenvoy/extension/examples/default/README.md")
			Expect(readmePath).To(BeAnExistingFile())
			Expect(filepath.Join(tempDir, ".getenvoy/extension/examples/default/envoy.tmpl.yaml")).To(BeAnExistingFile())
			Expect(filepath.Join(tempDir, ".getenvoy/extension/examples/default/example.yaml")).To(BeAnExistingFile())
			Expect(filepath.Join(tempDir, ".getenvoy/extension/examples/default/extension.txt")).To(BeAnExistingFile())

			data, err := ioutil.ReadFile(readmePath)
			Expect(err).ToNot(HaveOccurred())
			readme := string(data)
			Expect(readme).To(ContainSubstring("extension.txt"))
		})
	})

	Context("outside of a workspace directory", func() {
		It("should fail", func() {
			By("changing to a non-workspace dir")
			dir := chdir("testdata")

			By("running command")
			c.SetArgs([]string{"extension", "examples", "add"})
			err := cmdutil.Execute(c)
			Expect(err).To(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(BeEmpty())
			Expect(stderr.String()).To(Equal(fmt.Sprintf(`Error: there is no extension directory at or above: %s

Run 'getenvoy extension examples add --help' for usage.
`, dir)))
		})
	})
})
