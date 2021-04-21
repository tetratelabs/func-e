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
	"os/user"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	builtinconfig "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/toolchain/builtin"
	. "github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain/builtin"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain/types"
	"github.com/tetratelabs/getenvoy/pkg/test/cmd"
	. "github.com/tetratelabs/getenvoy/pkg/test/morerequire"
	ioutil "github.com/tetratelabs/getenvoy/pkg/util/io"
)

func TestBuiltinToolchain(t *testing.T) {
	// We use a fake docker command to capture the commandline that would be invoked
	dockerDir, revertPath := RequireOverridePath(t, "testdata/toolchain")
	defer revertPath()

	// Fake the current user so we can test it is used in the docker args
	expectedUser := user.User{Uid: "1001", Gid: "1002"}
	revertGetCurrentUser := cmd.OverrideGetCurrentUser(&expectedUser)
	defer revertGetCurrentUser()

	workspace, err := workspaces.GetWorkspaceAt("testdata/workspace")
	require.NoError(t, err)

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
		tool:           build,
		expectedStdout: fmt.Sprintf("%s/docker run -u 1001:1002 --rm -e GETENVOY_GOOS=%s -t -v %s:/source:delegated -w /source --init default/image build --output-file extension.wasm\n", dockerDir, runtime.GOOS, workspace.GetDir().GetRootDir()),
	},
		{
			name: "test using default container image",
			config: `
                kind: BuiltinToolchain
                container:
                  image: default/image
`,
			tool:           test,
			expectedStdout: fmt.Sprintf("%s/docker run -u 1001:1002 --rm -e GETENVOY_GOOS=%s -t -v %s:/source:delegated -w /source --init default/image test\n", dockerDir, runtime.GOOS, workspace.GetDir().GetRootDir()),
		},
		{
			name: "clean using default container image",
			config: `
                kind: BuiltinToolchain
                container:
                  image: default/image
`,
			tool:           clean,
			expectedStdout: fmt.Sprintf("%s/docker run -u 1001:1002 --rm -e GETENVOY_GOOS=%s -t -v %s:/source:delegated -w /source --init default/image clean\n", dockerDir, runtime.GOOS, workspace.GetDir().GetRootDir()),
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
			tool:           build,
			expectedStdout: fmt.Sprintf("%s/docker run -u 1001:1002 --rm -e GETENVOY_GOOS=%s -t -v %s:/source:delegated -w /source --init build/image build --output-file output/file.wasm\n", dockerDir, runtime.GOOS, workspace.GetDir().GetRootDir()),
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
			tool:           build,
			expectedStdout: fmt.Sprintf("%s/docker run -u 1001:1002 --rm -e GETENVOY_GOOS=%s -t -v %s:/source:delegated -w /source --init -e VAR=VALUE build/image build --output-file output/file.wasm\n", dockerDir, runtime.GOOS, workspace.GetDir().GetRootDir()),
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
                    - DOCKER_EXIT_CODE=3
                  output:
                    wasmFile: output/file.wasm
`,
			tool:           build,
			expectedStdout: fmt.Sprintf("%s/docker run -u 1001:1002 --rm -e GETENVOY_GOOS=%s -t -v %s:/source:delegated -w /source --init -e DOCKER_EXIT_CODE=3 build/image build --output-file output/file.wasm\n", dockerDir, runtime.GOOS, workspace.GetDir().GetRootDir()),
			expectedErr:    fmt.Sprintf("failed to execute an external command \"%s/docker run -u 1001:1002 --rm -e GETENVOY_GOOS=%s -t -v %s:/source:delegated -w /source --init -e DOCKER_EXIT_CODE=3 build/image build --output-file output/file.wasm\": exit status 3", dockerDir, runtime.GOOS, workspace.GetDir().GetRootDir()),
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
			tool:           test,
			expectedStdout: fmt.Sprintf("%s/docker run -u 1001:1002 --rm -e GETENVOY_GOOS=%s -t -v %s:/source:delegated -w /source --init test/image test\n", dockerDir, runtime.GOOS, workspace.GetDir().GetRootDir()),
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
			tool:           test,
			expectedStdout: fmt.Sprintf("%s/docker run -u 1001:1002 --rm -e GETENVOY_GOOS=%s -t -v %s:/source:delegated -w /source --init -v /host/path=/container/path test/image test\n", dockerDir, runtime.GOOS, workspace.GetDir().GetRootDir()),
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
                    - DOCKER_EXIT_CODE=3
`,
			tool:           test,
			expectedStdout: fmt.Sprintf("%s/docker run -u 1001:1002 --rm -e GETENVOY_GOOS=%s -t -v %s:/source:delegated -w /source --init -e DOCKER_EXIT_CODE=3 test/image test\n", dockerDir, runtime.GOOS, workspace.GetDir().GetRootDir()),
			expectedErr:    fmt.Sprintf("failed to execute an external command \"%s/docker run -u 1001:1002 --rm -e GETENVOY_GOOS=%s -t -v %s:/source:delegated -w /source --init -e DOCKER_EXIT_CODE=3 test/image test\": exit status 3", dockerDir, runtime.GOOS, workspace.GetDir().GetRootDir()),
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
			tool:           clean,
			expectedStdout: fmt.Sprintf("%s/docker run -u 1001:1002 --rm -e GETENVOY_GOOS=%s -t -v %s:/source:delegated -w /source --init clean/image clean\n", dockerDir, runtime.GOOS, workspace.GetDir().GetRootDir()),
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
			tool:           clean,
			expectedStdout: fmt.Sprintf("%s/docker run -u 1001:1002 --rm -e GETENVOY_GOOS=%s -t -v %s:/source:delegated -w /source --init -v /host/path=/container/path clean/image clean\n", dockerDir, runtime.GOOS, workspace.GetDir().GetRootDir()),
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
                    - DOCKER_EXIT_CODE=3
`,
			tool:           clean,
			expectedStdout: fmt.Sprintf("%s/docker run -u 1001:1002 --rm -e GETENVOY_GOOS=%s -t -v %s:/source:delegated -w /source --init -e DOCKER_EXIT_CODE=3 clean/image clean\n", dockerDir, runtime.GOOS, workspace.GetDir().GetRootDir()),
			expectedErr:    fmt.Sprintf("failed to execute an external command \"%s/docker run -u 1001:1002 --rm -e GETENVOY_GOOS=%s -t -v %s:/source:delegated -w /source --init -e DOCKER_EXIT_CODE=3 clean/image clean\": exit status 3", dockerDir, runtime.GOOS, workspace.GetDir().GetRootDir()),
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
