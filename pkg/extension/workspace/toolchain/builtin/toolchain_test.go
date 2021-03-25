// +build !windows

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

package builtin_test

import (
	"bytes"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config"
	builtinconfig "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/toolchain/builtin"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
	. "github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain/builtin"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain/types"
	ioutil "github.com/tetratelabs/getenvoy/pkg/util/io"
)

var _ = Describe("built-in toolchain", func() {

	var pathBackup string

	BeforeEach(func() {
		pathBackup = os.Getenv("PATH")
	})

	AfterEach(func() {
		os.Setenv("PATH", pathBackup)
	})

	BeforeEach(func() {
		// override PATH to overshadow `docker` executable during the test
		path := strings.Join([]string{"testdata/toolchain", pathBackup}, string(filepath.ListSeparator))
		os.Setenv("PATH", path)
	})

	var getCurrentUserBackup func() (*user.User, error)

	BeforeEach(func() {
		getCurrentUserBackup = GetCurrentUser
	})

	AfterEach(func() {
		GetCurrentUser = getCurrentUserBackup
	})

	BeforeEach(func() {
		GetCurrentUser = func() (*user.User, error) {
			return &user.User{Uid: "1001", Gid: "1002"}, nil
		}
	})

	var workspace model.Workspace

	BeforeEach(func() {
		w, err := workspaces.GetWorkspaceAt("testdata/workspace")
		Expect(err).ToNot(HaveOccurred())
		workspace = w
	})

	var stdin *bytes.Buffer
	var stdout *bytes.Buffer
	var stderr *bytes.Buffer
	var stdio ioutil.StdStreams

	BeforeEach(func() {
		stdin = new(bytes.Buffer)
		stdout = new(bytes.Buffer)
		stderr = new(bytes.Buffer)
		stdio = ioutil.StdStreams{
			In:  stdin,
			Out: stdout,
			Err: stderr,
		}
	})

	type testCase struct {
		config         string
		tool           func(toolchain types.Toolchain) error
		expectedStdOut string
		expectedErr    string
	}

	parseConfig := func(yaml string) *builtinconfig.ToolchainConfig {
		var cfg builtinconfig.ToolchainConfig
		Expect(config.Unmarshal([]byte(yaml), &cfg)).To(Succeed())
		Expect(cfg.Validate()).To(Succeed())
		return &cfg
	}

	build := func(toolchain types.Toolchain) error {
		By("running 'build' tool")
		return toolchain.Build(types.BuildContext{IO: stdio})
	}

	test := func(toolchain types.Toolchain) error {
		By("running 'test' tool")
		return toolchain.Test(types.TestContext{IO: stdio})
	}

	clean := func(toolchain types.Toolchain) error {
		By("running 'clean' tool")
		return toolchain.Clean(types.CleanContext{IO: stdio})
	}

	DescribeTable("should execute `docker run` with a given container image and Docker cli options",
		func(givenFn func() testCase) {
			given := givenFn()
			cfg := parseConfig(given.config)

			By("creating builtin toolchain")
			toolchain := NewToolchain("test", cfg, workspace)

			By("using toolchain")
			err := given.tool(toolchain)
			if given.expectedErr == "" {
				Expect(err).ToNot(HaveOccurred())
			} else {
				Expect(err).To(MatchError(given.expectedErr))
			}

			By("verifying stdout and stderr")
			Expect(stdout.String()).To(Equal(given.expectedStdOut))
			Expect(stderr.String()).To(Equal("docker stderr\n"))
		},
		Entry("build using default container image", func() testCase {
			return testCase{
				config: `
                kind: BuiltinToolchain
                container:
                  image: default/image
`,
				tool:           build,
				expectedStdOut: fmt.Sprintf("testdata/toolchain/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init default/image build --output-file extension.wasm\n", workspace.GetDir().GetRootDir()),
			}
		}),
		Entry("test using default container image", func() testCase {
			return testCase{
				config: `
                kind: BuiltinToolchain
                container:
                  image: default/image
`,
				tool:           test,
				expectedStdOut: fmt.Sprintf("testdata/toolchain/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init default/image test\n", workspace.GetDir().GetRootDir()),
			}
		}),
		Entry("clean using default container image", func() testCase {
			return testCase{
				config: `
                kind: BuiltinToolchain
                container:
                  image: default/image
`,
				tool:           clean,
				expectedStdOut: fmt.Sprintf("testdata/toolchain/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init default/image clean\n", workspace.GetDir().GetRootDir()),
			}
		}),
		Entry("build using given container image", func() testCase {
			return testCase{
				config: `
                kind: BuiltinToolchain
                container:
                  image: default/image
                build:
                  container:
                    image: build/image
                  output:
                    wasmFile: output/file.wasm
                # 'test' tool config should not affect 'build'
                test:
                  container:
                    image: test/image
`,
				tool:           build,
				expectedStdOut: fmt.Sprintf("testdata/toolchain/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init build/image build --output-file output/file.wasm\n", workspace.GetDir().GetRootDir()),
			}
		}),
		Entry("build using given container image and Docker cli options", func() testCase {
			return testCase{
				config: `
                kind: BuiltinToolchain
                container:
                  image: default/image
                build:
                  container:
                    image: build/image
                    options:
                    - -e
                    - VAR=VALUE
                  output:
                    wasmFile: output/file.wasm
                # 'test' tool config should not affect 'build'
                test:
                  container:
                    image: test/image
                    options:
                    - -v
                    - /host/path=/container/path
`,
				tool:           build,
				expectedStdOut: fmt.Sprintf("testdata/toolchain/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init -e VAR=VALUE build/image build --output-file output/file.wasm\n", workspace.GetDir().GetRootDir()),
			}
		}),
		Entry("build fails with a non-0 exit code", func() testCase {
			return testCase{
				config: `
                kind: BuiltinToolchain
                container:
                  image: default/image
                build:
                  container:
                    image: build/image
                    options:
                    - -e
                    - EXIT_CODE=3
                  output:
                    wasmFile: output/file.wasm
`,
				tool:           build,
				expectedStdOut: fmt.Sprintf("testdata/toolchain/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init -e EXIT_CODE=3 build/image build --output-file output/file.wasm\n", workspace.GetDir().GetRootDir()),
				expectedErr:    fmt.Sprintf("failed to execute an external command \"testdata/toolchain/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init -e EXIT_CODE=3 build/image build --output-file output/file.wasm\": exit status 3", workspace.GetDir().GetRootDir()),
			}
		}),
		Entry("test using given container image", func() testCase {
			return testCase{
				config: `
                kind: BuiltinToolchain
                container:
                  image: default/image
                # 'build' tool config should not affect 'test'
                build:
                  container:
                    image: build/image
                  output:
                    wasmFile: output/file.wasm
                test:
                  container:
                    image: test/image
`,
				tool:           test,
				expectedStdOut: fmt.Sprintf("testdata/toolchain/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init test/image test\n", workspace.GetDir().GetRootDir()),
			}
		}),
		Entry("test using given container image and Docker cli options", func() testCase {
			return testCase{
				config: `
                kind: BuiltinToolchain
                container:
                  image: default/image
                # 'build' tool config should not affect 'test'
                build:
                  container:
                    image: build/image
                    options:
                    - -e
                    - VAR=VALUE
                  output:
                    wasmFile: output/file.wasm
                test:
                  container:
                    image: test/image
                    options:
                    - -v
                    - /host/path=/container/path
`,
				tool:           test,
				expectedStdOut: fmt.Sprintf("testdata/toolchain/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init -v /host/path=/container/path test/image test\n", workspace.GetDir().GetRootDir()),
			}
		}),
		Entry("test fails with a non-0 exit code", func() testCase {
			return testCase{
				config: `
                kind: BuiltinToolchain
                container:
                  image: default/image
                test:
                  container:
                    image: test/image
                    options:
                    - -e
                    - EXIT_CODE=3
`,
				tool:           test,
				expectedStdOut: fmt.Sprintf("testdata/toolchain/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init -e EXIT_CODE=3 test/image test\n", workspace.GetDir().GetRootDir()),
				expectedErr:    fmt.Sprintf("failed to execute an external command \"testdata/toolchain/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init -e EXIT_CODE=3 test/image test\": exit status 3", workspace.GetDir().GetRootDir()),
			}
		}),
		Entry("clean using given container image", func() testCase {
			return testCase{
				config: `
                kind: BuiltinToolchain
                container:
                  image: default/image
                # 'build' tool config should not affect 'test'
                build:
                  container:
                    image: build/image
                  output:
                    wasmFile: output/file.wasm
                clean:
                  container:
                    image: clean/image
`,
				tool:           clean,
				expectedStdOut: fmt.Sprintf("testdata/toolchain/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init clean/image clean\n", workspace.GetDir().GetRootDir()),
			}
		}),
		Entry("clean using given container image and Docker cli options", func() testCase {
			return testCase{
				config: `
                kind: BuiltinToolchain
                container:
                  image: default/image
                # 'build' tool config should not affect 'test'
                build:
                  container:
                    image: build/image
                    options:
                    - -e
                    - VAR=VALUE
                  output:
                    wasmFile: output/file.wasm
                clean:
                  container:
                    image: clean/image
                    options:
                    - -v
                    - /host/path=/container/path
`,
				tool:           clean,
				expectedStdOut: fmt.Sprintf("testdata/toolchain/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init -v /host/path=/container/path clean/image clean\n", workspace.GetDir().GetRootDir()),
			}
		}),
		Entry("clean fails with a non-0 exit code", func() testCase {
			return testCase{
				config: `
                kind: BuiltinToolchain
                container:
                  image: default/image
                clean:
                  container:
                    image: clean/image
                    options:
                    - -e
                    - EXIT_CODE=3
`,
				tool:           clean,
				expectedStdOut: fmt.Sprintf("testdata/toolchain/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init -e EXIT_CODE=3 clean/image clean\n", workspace.GetDir().GetRootDir()),
				expectedErr:    fmt.Sprintf("failed to execute an external command \"testdata/toolchain/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init -e EXIT_CODE=3 clean/image clean\": exit status 3", workspace.GetDir().GetRootDir()),
			}
		}),
	)
})
