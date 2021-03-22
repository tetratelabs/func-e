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

package configdir_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/tetratelabs/getenvoy/pkg/extension/workspace/example/runtime/configdir"

	envoybootstrap "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"

	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/example/runtime"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
)

var _ = Describe("NewConfigDir()", func() {

	runContext := func(workspace model.Workspace, example model.Example) *runtime.RunContext {
		_, f := example.GetExtensionConfig()
		return &runtime.RunContext{
			Opts: runtime.RunOpts{
				Workspace: workspace,
				Example: runtime.ExampleOpts{
					Name:    "default",
					Example: example,
				},
				Extension: runtime.ExtensionOpts{
					WasmFile: `/path/to/extension.wasm`,
					Config:   *f,
				},
			},
		}
	}

	expectValidBootstrap := func(bootstrap *envoybootstrap.Bootstrap) {
		Expect(bootstrap).ToNot(BeNil())
		Expect(bootstrap.GetAdmin().GetAccessLogPath()).To(Equal("/dev/null"))
		Expect(bootstrap.GetAdmin().GetAddress().GetSocketAddress().GetAddress()).To(Equal("127.0.0.1"))
		Expect(bootstrap.GetAdmin().GetAddress().GetSocketAddress().GetPortValue()).To(Equal(uint32(9901)))
	}

	expectInvalidBootstrap := func(bootstrap *envoybootstrap.Bootstrap) {
		Expect(bootstrap).To(BeNil())
	}

	Describe("in case of valid input", func() {
		type testCase struct {
			workspaceDir    string
			isEnvoyTemplate func(string) bool
			expectBootstrap func(bootstrap *envoybootstrap.Bootstrap)
		}
		DescribeTable("should create a proper config directory",
			func(given testCase) {
				workspace, err := workspaces.GetWorkspaceAt(given.workspaceDir)
				Expect(err).ToNot(HaveOccurred())

				example, err := workspace.GetExample("default")
				Expect(err).ToNot(HaveOccurred())

				ctx := runContext(workspace, example)

				By("creating a config dir")
				configDir, err := NewConfigDir(ctx)
				Expect(err).ToNot(HaveOccurred())
				defer func() {
					By("removing config dir")
					Expect(configDir.Close()).To(Succeed())

					By("verifying config dir has been removed")
					_, err := os.Stat(configDir.GetDir())
					Expect(err).To(HaveOccurred())
					Expect(os.IsNotExist(err)).To(BeTrue())
				}()

				By("verifying the config dir")
				Expect(configDir.GetBootstrapFile()).ToNot(BeEmpty())
				given.expectBootstrap(configDir.GetBootstrap())

				By("verifying contents of the config dir")
				for _, fileName := range ctx.Opts.Example.GetFiles().GetNames() {
					expected, err := ioutil.ReadFile(filepath.Join(given.workspaceDir, "expected/getenvoy_extension_run", fileName))
					Expect(err).ToNot(HaveOccurred())
					actual, err := ioutil.ReadFile(filepath.Join(configDir.GetDir(), fileName))
					Expect(err).ToNot(HaveOccurred())

					if given.isEnvoyTemplate(fileName) {
						Expect(actual).To(MatchYAML(expected))
					} else {
						Expect(string(actual)).To(Equal(string(expected)))
					}

				}
			},
			Entry("envoy.tmpl.yaml", testCase{
				workspaceDir: "testdata/workspace1",
				isEnvoyTemplate: func(name string) bool {
					return name == "envoy.tmpl.yaml" //nolint:goconst
				},
				expectBootstrap: expectValidBootstrap,
			}),
			Entry("envoy.tmpl.yaml + lds.tmpl.yaml + cds.yaml", testCase{
				workspaceDir: "testdata/workspace2",
				isEnvoyTemplate: func(name string) bool {
					return name == "envoy.tmpl.yaml" || name == "lds.tmpl.yaml" || name == "cds.yaml"
				},
				expectBootstrap: expectValidBootstrap,
			}),
			Entry("envoy.tmpl.yaml: not a valid YAML", testCase{
				workspaceDir: "testdata/workspace3",
				isEnvoyTemplate: func(name string) bool {
					return false
				},
				expectBootstrap: expectInvalidBootstrap,
			}),
			Entry("envoy.tmpl.yaml: invalid paths to `lds` and `cds` files", testCase{
				workspaceDir: "testdata/workspace4",
				isEnvoyTemplate: func(name string) bool {
					return name == "envoy.tmpl.yaml"
				},
				expectBootstrap: expectValidBootstrap,
			}),
			Entry("envoy.tmpl.yaml: .txt configuration with \"//\" comment lines", testCase{
				workspaceDir: "testdata/workspace8",
				isEnvoyTemplate: func(name string) bool {
					return name == "envoy.tmpl.yaml"
				},
				expectBootstrap: expectValidBootstrap,
			}),
		)
	})

	Describe("in case of invalid input", func() {
		abs := func(path string) string {
			path, err := filepath.Abs(path)
			if err != nil {
				panic(err)
			}
			return path
		}

		type testCase struct {
			workspaceDir string
			expectedErr  string
		}
		//nolint:lll
		DescribeTable("should fail with a proper error",
			func(given testCase) {
				workspace, err := workspaces.GetWorkspaceAt(given.workspaceDir)
				Expect(err).ToNot(HaveOccurred())

				example, err := workspace.GetExample("default")
				Expect(err).ToNot(HaveOccurred())

				ctx := runContext(workspace, example)

				By("creating a config dir")
				configDir, err := NewConfigDir(ctx)
				Expect(configDir).To(BeNil())
				Expect(err).To(MatchError(given.expectedErr))
			},
			Entry("envoy.tmpl.yaml: invalid placeholder", testCase{
				workspaceDir: "testdata/workspace5",
				expectedErr:  fmt.Sprintf(`failed to process Envoy config template coming from %q: failed to render Envoy config template: template: :4:19: executing "" at <.GetEnvoy.DefaultValue>: error calling DefaultValue: unknown property "???"`, abs("testdata/workspace5/.getenvoy/extension/examples/default/envoy.tmpl.yaml")),
			}),
			Entry("envoy.tmpl.yaml + lds.tmpl.yaml + cds.yaml: invalid placeholder in lds.tmpl.yaml", testCase{
				workspaceDir: "testdata/workspace6",
				expectedErr:  fmt.Sprintf(`failed to process Envoy config template coming from %q: failed to render Envoy config template: template: :22:34: executing "" at <.GetEnvoy.Extension.Code>: error calling Code: unable to resolve Wasm module [???]: not supported yet`, abs("testdata/workspace6/.getenvoy/extension/examples/default/lds.tmpl.yaml")),
			}),
			Entry("envoy.tmpl.yaml + lds.tmpl.yaml + cds.tmpl.yaml: invalid placeholder in cds.tmpl.yaml", testCase{
				workspaceDir: "testdata/workspace7",
				expectedErr:  fmt.Sprintf(`failed to process Envoy config template coming from %q: failed to render Envoy config template: template: :1:18: executing "" at <.GetEnvoy.Extension.Config>: error calling Config: unable to resolve a named config [???]: not supported yet`, abs("testdata/workspace7/.getenvoy/extension/examples/default/cds.tmpl.yaml")),
			}),
		)
	})
})
