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

package toolchain_test

import (
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/fs"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
	toolchains "github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain"
)

var _ = Describe("LoadToolchain()", func() {
	It("should fail to load toolchains other than `default`", func() {
		workspace, err := workspaces.GetWorkspaceAt("testdata/workspace1")
		Expect(err).ToNot(HaveOccurred())

		_, err = toolchains.LoadToolchain("non-existing", workspace)
		Expect(err).To(MatchError(`unknown toolchain "non-existing". At the moment, only "default" toolchain is supported`))
	})

	//nolint:lll
	It("should fail to load toolchain of unknown kind", func() {
		workspace, err := workspaces.GetWorkspaceAt("testdata/workspace1")
		Expect(err).ToNot(HaveOccurred())

		_, err = toolchains.LoadToolchain(toolchains.Default, workspace)
		Expect(err).To(MatchError(fmt.Sprintf(`toolchain "default" has invalid configuration coming from "%s/.getenvoy/extension/toolchains/default.yaml": `+
			`unknown toolchain kind "UnknownToolchain"`, workspace.GetDir().GetRootDir())))
	})

	//nolint:lll
	It("should fail to load toolchain with invalid config", func() {
		workspace, err := workspaces.GetWorkspaceAt("testdata/workspace2")
		Expect(err).ToNot(HaveOccurred())

		_, err = toolchains.LoadToolchain(toolchains.Default, workspace)
		Expect(err).To(MatchError(fmt.Sprintf(`toolchain "default" has invalid configuration coming from "%s/.getenvoy/extension/toolchains/default.yaml": `+
			`configuration of 'build' tool is not valid: container configuration is not valid: "?invalid value?" is not a valid image name: invalid reference format`, workspace.GetDir().GetRootDir())))
	})

	It("should load toolchain with a valid config", func() {
		workspace, err := workspaces.GetWorkspaceAt("testdata/workspace3")
		Expect(err).ToNot(HaveOccurred())

		builder, err := toolchains.LoadToolchain(toolchains.Default, workspace)
		Expect(err).ToNot(HaveOccurred())

		toolchain, err := builder.Build()
		Expect(err).ToNot(HaveOccurred())
		Expect(toolchain).ToNot(BeNil())
	})

	It("should create default toolchain if missing", func() {
		tempDir, err := ioutil.TempDir("", "workspace")
		Expect(err).ToNot(HaveOccurred())
		defer func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		}()

		dir, err := fs.CreateWorkspaceDir(tempDir)
		Expect(err).ToNot(HaveOccurred())

		err = dir.WriteFile(model.DescriptorFile, []byte(`
kind: Extension
language: rust
category: envoy.filters.http
`))
		Expect(err).ToNot(HaveOccurred())

		workspace, err := workspaces.GetWorkspaceAt(tempDir)
		Expect(err).ToNot(HaveOccurred())

		builder, err := toolchains.LoadToolchain(toolchains.Default, workspace)
		Expect(err).ToNot(HaveOccurred())

		toolchain, err := builder.Build()
		Expect(err).ToNot(HaveOccurred())
		Expect(toolchain).ToNot(BeNil())

		Expect(workspace.HasToolchain(toolchains.Default)).To(BeTrue())
		_, data, err := workspace.GetToolchainConfigBytes(toolchains.Default)
		Expect(err).ToNot(HaveOccurred())
		Expect(data).ToNot(BeEmpty())
	})
})
