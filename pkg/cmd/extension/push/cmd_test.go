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

package push_test

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

const (
	localRegistryWasmImageRef = "localhost:5000/getenvoy/sample"
)

var _ = Describe("getenvoy extension push", func() {

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

	chdir := func(path string) string {
		dir, err := filepath.Abs(path)
		Expect(err).ToNot(HaveOccurred())

		err = os.Chdir(dir)
		Expect(err).ToNot(HaveOccurred())

		return dir
	}

	//TODO(musaprg): write teardown process for local registries if it's needed

	//nolint:lll
	Context("inside a workspace directory", func() {
		When("if the image ref is valid", func() {
			It("should succeed", func() {
				By("changing to a workspace dir")
				_ = chdir("testdata/workspace")

				By("push to local registry")
				c.SetArgs([]string{"extension", "push", localRegistryWasmImageRef})
				err := cmdutil.Execute(c)
				Expect(err).ToNot(HaveOccurred())

				By("verifying command output")
				//TODO(musaprg): implement me
			})
		})
		When("if the image ref is invalid", func() {
			It("should fail", func() {
				//TODO(musaprg): implement me
			})
		})
	})

	Context("outside of a workspace directory", func() {
		When("if the target wasm binary is specified", func() {
			It("should succeed", func() {
				By("changing to a non-workspace dir")
				dir := chdir("testdata")

				By("running command")
				c.SetArgs([]string{"extension", "push", "--extension-file", filepath.Join(dir, "workspace", "extension.wasm")})
				err := cmdutil.Execute(c)
				Expect(err).NotTo(HaveOccurred())

				By("verifying command output")
				Expect(stdout.String()).To(BeEmpty())
				//TODO(musaprg): implement me
			})
		})
		When("if no wasm binary specified", func() {
			It("should fail", func() {
				By("changing to a non-workspace dir")
				dir := chdir("testdata")

				By("running command")
				c.SetArgs([]string{"extension", "push"})
				err := cmdutil.Execute(c)
				Expect(err).To(HaveOccurred())

				By("verifying command output")
				Expect(stdout.String()).To(BeEmpty())
				Expect(stderr.String()).To(Equal(fmt.Sprintf(`Error: there is no extension directory at or above: %s

Run 'getenvoy extension build --help' for usage.
`, dir)))
			})
		})
	})
})
