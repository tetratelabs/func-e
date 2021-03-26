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

var _ = Describe("getenvoy extension examples remove", func() {

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

	It("should require --name flag", func() {
		By("running command")
		c.SetArgs([]string{"extension", "examples", "remove"})
		err := cmdutil.Execute(c)
		Expect(err).To(HaveOccurred())

		By("verifying command output")
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(Equal(`Error: example name cannot be empty

Run 'getenvoy extension examples remove --help' for usage.
`))
	})

	//nolint:lll
	It("should validate --name flag", func() {
		By("running command")
		c.SetArgs([]string{"extension", "examples", "remove", "--name", "my:example"})
		err := cmdutil.Execute(c)
		Expect(err).To(HaveOccurred())

		By("verifying command output")
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(Equal(`Error: "my:example" is not a valid example name. Example name must match the format "^[a-z0-9._-]+$". E.g., 'my.example', 'my-example' or 'my_example'

Run 'getenvoy extension examples remove --help' for usage.
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

		It("should be able to remove the 'default' example setup", func() {
			By("simulating a workspace with the 'default' example setup")
			err := copy.Copy("testdata/workspace2", tempDir)
			Expect(err).NotTo(HaveOccurred())

			By("changing to a workspace dir")
			chdir(tempDir)

			By("running command")
			c.SetArgs([]string{"extension", "examples", "remove", "--name", "default"})
			err = cmdutil.Execute(c)
			Expect(err).ToNot(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(BeEmpty())
			Expect(stderr.String()).To(Equal(`Removing example setup:
* .getenvoy/extension/examples/default/envoy.tmpl.yaml
* .getenvoy/extension/examples/default/example.yaml
* .getenvoy/extension/examples/default/extension.json
Done!
`))
			By("verifying file system")
			Expect(filepath.Join(tempDir, ".getenvoy/extension/examples/default")).NotTo(BeAnExistingFile())
		})

		It("should be able to remove a non-default example setup", func() {
			By("simulating a workspace with a non-default example setup")
			err := copy.Copy("testdata/workspace3", tempDir)
			Expect(err).NotTo(HaveOccurred())

			By("changing to a workspace dir")
			chdir(tempDir)

			By("running command")
			c.SetArgs([]string{"extension", "examples", "remove", "--name", "another"})
			err = cmdutil.Execute(c)
			Expect(err).ToNot(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(BeEmpty())
			Expect(stderr.String()).To(Equal(`Removing example setup:
* .getenvoy/extension/examples/another/envoy.tmpl.yaml
* .getenvoy/extension/examples/another/example.yaml
* .getenvoy/extension/examples/another/extension.json
Done!
`))
			By("verifying file system")
			Expect(filepath.Join(tempDir, ".getenvoy/extension/examples/another")).NotTo(BeAnExistingFile())
		})

		It("should not fail if such example doesn't exist", func() {
			By("simulating a workspace with the 'default' example")
			err := copy.Copy("testdata/workspace2", tempDir)
			Expect(err).NotTo(HaveOccurred())

			By("changing to a workspace dir")
			chdir(tempDir)

			By("running command")
			c.SetArgs([]string{"extension", "examples", "remove", "--name", "non-existing-setup"})
			err = cmdutil.Execute(c)
			Expect(err).NotTo(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(BeEmpty())
			Expect(stderr.String()).To(Equal(`There is no example setup named "non-existing-setup".

Use "getenvoy extension examples list" to list existing example setups.
`))
		})
	})

	Context("outside of a workspace directory", func() {
		It("should fail", func() {
			By("changing to a non-workspace dir")
			dir := chdir("testdata")

			By("running command")
			c.SetArgs([]string{"extension", "examples", "remove", "--name", "default"})
			err := cmdutil.Execute(c)
			Expect(err).To(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(BeEmpty())
			Expect(stderr.String()).To(Equal(fmt.Sprintf(`Error: there is no extension directory at or above: %s

Run 'getenvoy extension examples remove --help' for usage.
`, dir)))
		})
	})
})
