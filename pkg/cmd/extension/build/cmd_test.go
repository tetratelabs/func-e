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

package build_test

import (
	"fmt"
	"os/user"
	"testing"

	"github.com/stretchr/testify/require"

	cmd2 "github.com/tetratelabs/getenvoy/pkg/test/cmd"
	cmdutil "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

// relativeWorkspaceDir points to a usable pre-initialized workspace
const relativeWorkspaceDir = "testdata/workspace"

func TestGetEnvoyExtensionBuildValidateFlag(t *testing.T) {
	type testCase struct {
		flag        string
		flagValue   string
		expectedErr string
	}

	tests := []testCase{
		{
			flag:        "--toolchain-container-image",
			flagValue:   "?invalid value?",
			expectedErr: `"?invalid value?" is not a valid image name: invalid reference format`,
		},
		{
			flag:        "--toolchain-container-options",
			flagValue:   "imbalanced ' quotes",
			expectedErr: `"imbalanced ' quotes" is not a valid command line string`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.flag+"="+test.flagValue, func(t *testing.T) {
			// Run "getenvoy extension build" with the flags we are testing
			cmd, stdout, stderr := cmd2.NewRootCommand()
			cmd.SetArgs([]string{"extension", "build", test.flag, test.flagValue})
			err := cmdutil.Execute(cmd)
			require.EqualError(t, err, test.expectedErr, `expected an error running [%v]`, cmd)

			// Verify the command failed with the expected error
			require.Empty(t, stdout.String(), `expected no stdout running [%v]`, cmd)
			expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy extension build --help' for usage.\n", test.expectedErr)
			require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, cmd)
		})
	}
}

func TestGetEnvoyExtensionBuildFailsOutsideWorkspaceDirectory(t *testing.T) {
	// Change to a non-workspace dir
	dir, revertWd := cmd2.RequireChDir(t, relativeWorkspaceDir+"/..")
	defer revertWd()

	// Run "getenvoy extension build"
	cmd, stdout, stderr := cmd2.NewRootCommand()
	cmd.SetArgs([]string{"extension", "build"})
	err := cmdutil.Execute(cmd)

	// Verify the command failed with the expected error
	expectedErr := "there is no extension directory at or above: " + dir
	require.EqualError(t, err, expectedErr, `expected an error running [%v]`, cmd)
	require.Empty(t, stdout.String(), `expected no stdout running [%v]`, cmd)
	expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy extension build --help' for usage.\n", expectedErr)
	require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, cmd)
}

func TestGetEnvoyExtensionBuild(t *testing.T) {
	// We use a fake docker command to capture the commandline that would be invoked
	dockerDir, revertPath := cmd2.RequireOverridePath(t, cmd2.FakeDockerDir)
	defer revertPath()

	// "getenvoy extension build" must be in a valid workspace directory
	workspaceDir, revertWd := cmd2.RequireChDir(t, relativeWorkspaceDir)
	defer revertWd()

	// Fake the current user so we can test it is used in the docker args
	expectedUser := user.User{Uid: "1001", Gid: "1002"}
	revertGetCurrentUser := cmd2.OverrideGetCurrentUser(&expectedUser)
	defer revertGetCurrentUser()

	// Run "getenvoy extension build"
	cmd, stdout, stderr := cmd2.NewRootCommand()
	cmd.SetArgs([]string{"extension", "build"})
	err := cmdutil.Execute(cmd)

	// We expect docker to run from the correct path, as the current user and mount a volume for the correct workspace.
	expectedDockerExec := fmt.Sprintf("%s/docker run -u %s:%s --rm -t -v %s:/source -w /source --init getenvoy/extension-rust-builder:latest build --output-file target/getenvoy/extension.wasm",
		dockerDir, expectedUser.Uid, expectedUser.Gid, workspaceDir)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, cmd)
	require.Equal(t, expectedDockerExec+"\n", stdout.String(), `expected stdout running [%v]`, cmd)
	require.Equal(t, "docker stderr\n", stderr.String(), `expected stderr running [%v]`, cmd)
}

// This tests --toolchain-container flags become docker command options
func TestGetEnvoyExtensionBuildWithDockerOptions(t *testing.T) {
	// We use a fake docker command to capture the commandline that would be invoked
	_, revertPath := cmd2.RequireOverridePath(t, cmd2.FakeDockerDir)
	defer revertPath()

	// "getenvoy extension build" must be in a valid workspace directory
	_, revertWd := cmd2.RequireChDir(t, relativeWorkspaceDir)
	defer revertWd()

	// Run "getenvoy extension build"
	cmd, stdout, stderr := cmd2.NewRootCommand()
	cmd.SetArgs([]string{"extension", "build",
		"--toolchain-container-image", "build/image",
		"--toolchain-container-options", `-e 'VAR=VALUE' -v "/host:/container"`,
	})
	err := cmdutil.Execute(cmd)

	// Verify the command's stdout includes the init args. TestGetEnvoyExtensionBuild tests the rest of stdout.
	require.NoError(t, err, `expected no error running [%v]`, cmd)
	require.Regexp(t, ".*--init -e VAR=VALUE -v /host:/container build/image build.*", stdout.String(), `expected stdout running [%v]`, cmd)
	require.Equal(t, "docker stderr\n", stderr.String(), `expected stderr running [%v]`, cmd)
}

// TestGetEnvoyExtensionBuildFail ensures build failures show useful information in stderr
func TestGetEnvoyExtensionBuildFail(t *testing.T) {
	// We use a fake docker command to capture the commandline that would be invoked, and force a failure.
	dockerDir, revertPath := cmd2.RequireOverridePath(t, cmd2.FakeDockerDir)
	defer revertPath()

	// "getenvoy extension build" must be in a valid workspace directory
	workspaceDir, revertWd := cmd2.RequireChDir(t, relativeWorkspaceDir)
	defer revertWd()

	// Fake the current user so we can test it is used in the docker args
	expectedUser := user.User{Uid: "1001", Gid: "1002"}
	revertGetCurrentUser := cmd2.OverrideGetCurrentUser(&expectedUser)
	defer revertGetCurrentUser()

	// "-e DOCKER_EXIT_CODE=3" is a special instruction handled in the fake docker script
	toolchainOptions := "-e DOCKER_EXIT_CODE=3"
	// Run "getenvoy extension build"
	cmd, stdout, stderr := cmd2.NewRootCommand()
	cmd.SetArgs([]string{"extension", "build", "--toolchain-container-options", toolchainOptions})
	err := cmdutil.Execute(cmd)

	// We expect the exit instruction to have gotten to the fake docker script, along with the default options.
	expectedDockerExec := fmt.Sprintf("%s/docker run -u %s:%s --rm -t -v %s:/source -w /source --init %s getenvoy/extension-rust-builder:latest build --output-file target/getenvoy/extension.wasm",
		dockerDir, expectedUser.Uid, expectedUser.Gid, workspaceDir, toolchainOptions)

	// Verify the command failed with the expected error.
	expectedErr := fmt.Sprintf(`failed to build Envoy extension using "default" toolchain: failed to execute an external command "%s": exit status 3`, expectedDockerExec)
	require.EqualError(t, err, expectedErr, `expected an error running [%v]`, cmd)

	// We should see stdout because the docker script was invoked.
	require.Equal(t, expectedDockerExec+"\n", stdout.String(), `expected stdout running [%v]`, cmd)

	// We also expect "docker stderr" in the output for the same reason.
	expectedStderr := fmt.Sprintf("docker stderr\nError: %s\n\nRun 'getenvoy extension build --help' for usage.\n", expectedErr)
	require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, cmd)
}
