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
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	builtinconfig "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/toolchain/builtin"
	. "github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain/builtin"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain/types"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
	ioutil "github.com/tetratelabs/getenvoy/pkg/util/io"
)

func TestBuiltinToolchain(t *testing.T) {
	fakeDocker, removeFakeDocker := morerequire.RequireCaptureScript(t, "docker")
	defer removeFakeDocker()

	workspace, err := workspaces.GetWorkspaceAt("testdata/workspace")
	require.NoError(t, err)
	extensionDir := workspace.GetDir().GetRootDir()

	baseDockerArgs := fmt.Sprintf(`run -u %s --rm -e GETENVOY_GOOS=%s -t -v %s:/source -w /source --init`,
		DefaultDockerUser, runtime.GOOS, extensionDir)

	tests := []struct {
		name           string
		config         string
		tool           func(toolchain types.Toolchain, stdio ioutil.StdStreams) error
		expectedStdout string
		expectedErr    string
	}{{
		name: "build using default container image",
		config: `
                kind: BuiltinToolchain
                container:
                  image: default/image
`,
		tool: build,
		expectedStdout: fmt.Sprintf(`docker wd: %s
docker bin: %s
docker args: %s default/image build --output-file extension.wasm
`, extensionDir, fakeDocker, baseDockerArgs),
	},
		{
			name: "test using default container image",
			config: `
                kind: BuiltinToolchain
                container:
                  image: default/image
`,
			tool: test,
			expectedStdout: fmt.Sprintf("docker wd: %s\ndocker bin: %s\ndocker args: %s default/image test\n",
				extensionDir, fakeDocker, baseDockerArgs),
		},
		{
			name: "clean using default container image",
			config: `
                kind: BuiltinToolchain
                container:
                  image: default/image
`,
			tool: clean,
			expectedStdout: fmt.Sprintf("docker wd: %s\ndocker bin: %s\ndocker args: %s default/image clean\n",
				extensionDir, fakeDocker, baseDockerArgs),
		},
		{
			name: "build using test container image",
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
			tool: build,
			expectedStdout: fmt.Sprintf("docker wd: %s\ndocker bin: %s\ndocker args: %s build/image build --output-file output/file.wasm\n",
				extensionDir, fakeDocker, baseDockerArgs),
		},
		{
			name: "build using test container image and Docker cli options",
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
			tool: build,
			expectedStdout: fmt.Sprintf("docker wd: %s\ndocker bin: %s\ndocker args: %s -e VAR=VALUE build/image build --output-file output/file.wasm\n",
				extensionDir, fakeDocker, baseDockerArgs),
		},
		{
			name: "build fails with a non-0 exit code",
			config: `
                kind: BuiltinToolchain
                container:
                  image: default/image
                build:
                  container:
                    image: build/image
                    options:
                    - -e
                    - docker_exit=3
                  output:
                    wasmFile: output/file.wasm
`,
			tool: build,
			expectedStdout: fmt.Sprintf("docker wd: %s\ndocker bin: %s\ndocker args: %s -e docker_exit=3 build/image build --output-file output/file.wasm\n",
				extensionDir, fakeDocker, baseDockerArgs),
			expectedErr: fmt.Sprintf("failed to execute an external command \"%s %s -e docker_exit=3 build/image build --output-file output/file.wasm\": exit status 3", fakeDocker, baseDockerArgs),
		},
		{
			name: "test using test container image",
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
			tool: test,
			expectedStdout: fmt.Sprintf("docker wd: %s\ndocker bin: %s\ndocker args: %s test/image test\n",
				extensionDir, fakeDocker, baseDockerArgs),
		},
		{
			name: "test using test container image and Docker cli options",
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
			tool: test,
			expectedStdout: fmt.Sprintf("docker wd: %s\ndocker bin: %s\ndocker args: %s -v /host/path=/container/path test/image test\n",
				extensionDir, fakeDocker, baseDockerArgs),
		},
		{
			name: "test fails with a non-0 exit code",
			config: `
                kind: BuiltinToolchain
                container:
                  image: default/image
                test:
                  container:
                    image: test/image
                    options:
                    - -e
                    - docker_exit=3
`,
			tool: test,
			expectedStdout: fmt.Sprintf("docker wd: %s\ndocker bin: %s\ndocker args: %s -e docker_exit=3 test/image test\n",
				extensionDir, fakeDocker, baseDockerArgs),
			expectedErr: fmt.Sprintf("failed to execute an external command \"%s %s -e docker_exit=3 test/image test\": exit status 3", fakeDocker, baseDockerArgs),
		},
		{
			name: "clean using test container image",
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
			tool: clean,
			expectedStdout: fmt.Sprintf("docker wd: %s\ndocker bin: %s\ndocker args: %s clean/image clean\n",
				extensionDir, fakeDocker, baseDockerArgs),
		},
		{
			name: "clean using test container image and Docker cli options",
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
			tool: clean,
			expectedStdout: fmt.Sprintf("docker wd: %s\ndocker bin: %s\ndocker args: %s -v /host/path=/container/path clean/image clean\n",
				extensionDir, fakeDocker, baseDockerArgs),
		},
		{
			name: "clean fails with a non-0 exit code",
			config: `
                kind: BuiltinToolchain
                container:
                  image: default/image
                clean:
                  container:
                    image: clean/image
                    options:
                    - -e
                    - docker_exit=3
`,
			tool: clean,
			expectedStdout: fmt.Sprintf("docker wd: %s\ndocker bin: %s\ndocker args: %s -e docker_exit=3 clean/image clean\n",
				extensionDir, fakeDocker, baseDockerArgs),
			expectedErr: fmt.Sprintf("failed to execute an external command \"%s %s -e docker_exit=3 clean/image clean\": exit status 3", fakeDocker, baseDockerArgs),
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			var cfg builtinconfig.ToolchainConfig
			err = yaml.Unmarshal([]byte(test.config), &cfg)
			require.NoError(t, err)

			err = cfg.Validate()
			require.NoError(t, err)

			cfg.Container.DockerPath = fakeDocker
			toolchain := NewToolchain("test", &cfg, workspace)
			stdin, stdout, stderr := new(bytes.Buffer), new(bytes.Buffer), new(bytes.Buffer)
			stdio := ioutil.StdStreams{In: stdin, Out: stdout, Err: stderr}

			// invoke the toolchain
			err := test.tool(toolchain, stdio)
			if test.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, test.expectedErr)
			}

			require.Equal(t, test.expectedStdout, stdout.String())
			require.Equal(t, "docker stderr\n", stderr.String())
		})
	}
}

func build(toolchain types.Toolchain, stdio ioutil.StdStreams) error {
	return toolchain.Build(types.BuildContext{IO: stdio})
}

func test(toolchain types.Toolchain, stdio ioutil.StdStreams) error {
	return toolchain.Test(types.TestContext{IO: stdio})
}

func clean(toolchain types.Toolchain, stdio ioutil.StdStreams) error {
	return toolchain.Clean(types.CleanContext{IO: stdio})
}
