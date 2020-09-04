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
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"

	"github.com/tetratelabs/getenvoy/pkg/cmd"

	cmdutil "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

var _ = Describe("getenvoy extension examples list", func() {

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
		It("should support a case with no examples", func() {
			By("changing to a workspace dir")
			chdir("testdata/workspace1")

			By("running command")
			c.SetArgs([]string{"extension", "examples", "list"})
			err := cmdutil.Execute(c)
			Expect(err).ToNot(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(BeEmpty())
			Expect(stderr.String()).To(Equal(`Extension has no example setups.

Use "getenvoy extension examples add --help" for more information on how to add one.
`))
		})

		It("should support a case with 1 example", func() {
			By("changing to a workspace dir")
			chdir("testdata/workspace2")

			By("running command")
			c.SetArgs([]string{"extension", "examples", "list"})
			err := cmdutil.Execute(c)
			Expect(err).ToNot(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(Equal(`EXAMPLE
default
`))
			Expect(stderr.String()).To(BeEmpty())
		})

		It("should support a case with multiple examples", func() {
			By("changing to a workspace dir")
			chdir("testdata/workspace3")

			By("running command")
			c.SetArgs([]string{"extension", "examples", "list"})
			err := cmdutil.Execute(c)
			Expect(err).ToNot(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(Equal(`EXAMPLE
another
default
`))
			Expect(stderr.String()).To(BeEmpty())
		})
	})

	Context("outside of a workspace directory", func() {
		It("should fail", func() {
			By("changing to a non-workspace dir")
			dir := chdir("testdata")

			By("running command")
			c.SetArgs([]string{"extension", "examples", "list"})
			err := cmdutil.Execute(c)
			Expect(err).To(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(BeEmpty())
			Expect(stderr.String()).To(Equal(fmt.Sprintf(`Error: there is no extension directory at or above: %s

Run 'getenvoy extension examples list --help' for usage.
`, dir)))
		})
	})
})
