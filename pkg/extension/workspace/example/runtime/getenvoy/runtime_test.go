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

package getenvoy_test

import (
	"bytes"
	"fmt"
	stdioutil "io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/tetratelabs/getenvoy/pkg/extension/workspace/example/runtime/getenvoy"

	"github.com/tetratelabs/getenvoy/pkg/common"
	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/example/runtime"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"

	argutil "github.com/tetratelabs/getenvoy/pkg/util/args"
	ioutil "github.com/tetratelabs/getenvoy/pkg/util/io"
)

var _ = Describe("runtime", func() {
	Describe("Run()", func() {

		var backupHomeDir string

		BeforeEach(func() {
			backupHomeDir = common.HomeDir
		})

		AfterEach(func() {
			common.HomeDir = backupHomeDir
		})

		var tempHomeDir string

		BeforeEach(func() {
			dir, err := stdioutil.TempDir("", "getenvoy-home")
			Expect(err).NotTo(HaveOccurred())
			tempHomeDir = dir
		})

		AfterEach(func() {
			if tempHomeDir != "" {
				Expect(os.RemoveAll(tempHomeDir)).To(Succeed())
			}
		})

		BeforeEach(func() {
			common.HomeDir = tempHomeDir
		})

		envoyPath := func() string {
			path, err := filepath.Abs("testdata/envoy/bin/envoy")
			Expect(err).ToNot(HaveOccurred())
			return path
		}

		var stdout *bytes.Buffer
		var stderr *bytes.Buffer

		BeforeEach(func() {
			stdout = new(bytes.Buffer)
			stderr = new(bytes.Buffer)
		})

		runContext := func(workspace model.Workspace, example model.Example) *runtime.RunContext {
			return &runtime.RunContext{
				Opts: runtime.RunOpts{
					Workspace: workspace,
					Example: runtime.ExampleOpts{
						Name:    "default",
						Example: example,
					},
					Extension: runtime.ExtensionOpts{
						WasmFile: `/path/to/extension.wasm`,
						Config: model.File{
							Source:  "/path/to/config",
							Content: []byte(`{"key2":"value2"}`),
						},
					},
					Envoy: runtime.EnvoyOpts{
						Path: envoyPath(),
					},
				},
				IO: ioutil.StdStreams{
					Out: stdout,
					Err: stderr,
				},
			}
		}

		Describe("in case of valid input", func() {
			type testCase struct {
				workspaceDir    string
				isEnvoyTemplate func(string) bool
			}
			DescribeTable("should run Envoy with a proper config",
				func(given testCase) {
					workspace, err := workspaces.GetWorkspaceAt(given.workspaceDir)
					Expect(err).ToNot(HaveOccurred())

					example, err := workspace.GetExample("default")
					Expect(err).ToNot(HaveOccurred())

					ctx := runContext(workspace, example)

					By("running Envoy")
					err = NewRuntime().Run(ctx)
					Expect(err).ToNot(HaveOccurred())

					By("verifying Envoy output")
					Expect(stdout.String()).NotTo(BeEmpty())
					Expect(stderr.String()).To(Equal("envoy stderr\n"))

					By("verifying Envoy arguments")
					args, err := argutil.SplitCommandLine(stdout.String())
					Expect(err).ToNot(HaveOccurred())
					Expect(args).To(HaveLen(3))
					Expect(args[0]).To(Equal(ctx.Opts.Envoy.Path))
					Expect(args[1]).To(Equal("-c"))
				},
				Entry("envoy.tmpl.yaml", testCase{
					workspaceDir: "../configdir/testdata/workspace1",
					isEnvoyTemplate: func(name string) bool {
						return name == "envoy.tmpl.yaml"
					},
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

					By("running Envoy")
					err = NewRuntime().Run(ctx)
					Expect(err).To(MatchError(given.expectedErr))
				},
				Entry("envoy.tmpl.yaml: invalid placeholder", testCase{
					workspaceDir: "../configdir/testdata/workspace5",
					expectedErr:  fmt.Sprintf(`failed to process Envoy config template coming from %q: failed to render Envoy config template: template: :4:19: executing "" at <.GetEnvoy.DefaultValue>: error calling DefaultValue: unknown property "???"`, abs("../configdir/testdata/workspace5/.getenvoy/extension/examples/default/envoy.tmpl.yaml")),
				}),
			)
		})
	})
})
