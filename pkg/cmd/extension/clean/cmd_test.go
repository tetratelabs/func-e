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

package clean_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	rootcmd "github.com/tetratelabs/getenvoy/pkg/cmd"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain/builtin"
	"github.com/tetratelabs/getenvoy/pkg/globals"
	"github.com/tetratelabs/getenvoy/pkg/test/cmd"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

// relativeExtensionDir points to a usable pre-initialized workspace
const relativeExtensionDir = "testdata/workspace"

func TestGetEnvoyExtensionCleanValidateFlag(t *testing.T) {
	type testCase struct {
		name        string
		args        []string
		expectedErr string
	}

	tests := []testCase{
		{
			name:        "--toolchain-container-image with invalid value",
			args:        []string{"extension", "clean", "--toolchain-container-image", "?invalid value?"},
			expectedErr: `"?invalid value?" is not a valid image name: invalid reference format`,
		},
		{
			name:        "--toolchain-container-options has imbalanced quotes",
			args:        []string{"extension", "clean", "--toolchain-container-options", "imbalanced ' quotes"},
			expectedErr: `"imbalanced ' quotes" is not a valid command line string`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			// Run "getenvoy extension clean" with the args we are testing
			c, stdout, stderr := cmd.NewRootCommand(&globals.GlobalOpts{})
			c.SetArgs(test.args)
			err := rootcmd.Execute(c)

			require.EqualError(t, err, test.expectedErr, `expected an error running [%v]`, c)

			// Verify the command failed with the expected error
			require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
			expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy extension clean --help' for usage.\n", test.expectedErr)
			require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, c)
		})
	}
}

func TestGetEnvoyExtensionCleanFailsOutsideExtensionDirectory(t *testing.T) {
	// Use to a non-workspace dir
	o := &globals.GlobalOpts{ExtensionDir: morerequire.RequireAbs(t, ".")}

	// Run "getenvoy extension clean"
	c, stdout, stderr := cmd.NewRootCommand(o)
	c.SetArgs([]string{"extension", "clean"})
	err := rootcmd.Execute(c)

	// Verify the command failed with the expected error
	expectedErr := fmt.Sprintf("not an extension directory %q", o.ExtensionDir)
	require.EqualError(t, err, expectedErr, `expected an error running [%v]`, c)
	require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
	expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy extension clean --help' for usage.\n", expectedErr)
	require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, c)
}

func TestGetEnvoyExtensionClean(t *testing.T) {
	o, cleanup := morerequire.RequireGlobalOpts(t, relativeExtensionDir)
	defer cleanup()

	// Run "getenvoy extension clean"
	c, stdout, stderr := cmd.NewRootCommand(o)
	c.SetArgs([]string{"extension", "clean"})
	err := rootcmd.Execute(c)

	// We expect docker to run from the correct path, as the current user and mount a volume for the correct workspace.
	expectedStdout := fmt.Sprintf(`docker wd: %s
docker bin: %s
docker args: run -u %s --rm -e GETENVOY_GOOS=%s -t -v %s:/source -w /source --init getenvoy/extension-rust-builder:latest clean
`, o.ExtensionDir, o.DockerPath, builtin.DefaultDockerUser, runtime.GOOS, o.ExtensionDir)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, c)
	require.Equal(t, expectedStdout, stdout.String(), `expected stdout running [%v]`, c)
	require.Equal(t, "docker stderr\n", stderr.String(), `expected stderr running [%v]`, c)
}

// This tests --toolchain-container flags become docker command options
func TestGetEnvoyExtensionCleanWithDockerOptions(t *testing.T) {
	o, cleanup := morerequire.RequireGlobalOpts(t, relativeExtensionDir)
	defer cleanup()

	// Run "getenvoy extension clean"
	c, stdout, stderr := cmd.NewRootCommand(o)
	c.SetArgs([]string{"extension", "clean",
		"--toolchain-container-image", "clean/image",
		"--toolchain-container-options", `-e 'VAR=VALUE' -v "/host:/container"`,
	})
	err := rootcmd.Execute(c)

	// Verify the command's stdout includes the init args. TestGetEnvoyExtensionClean tests the rest of stdout.
	require.NoError(t, err, `expected no error running [%v]`, c)
	require.Regexp(t, ".*--init -e VAR=VALUE -v /host:/container clean/image clean.*", stdout.String(), `expected stdout running [%v]`, c)
	require.Equal(t, "docker stderr\n", stderr.String(), `expected stderr running [%v]`, c)
}

// TestGetEnvoyExtensionCleanFail ensures clean failures show useful information in stderr
func TestGetEnvoyExtensionCleanFail(t *testing.T) {
	o, cleanup := morerequire.RequireGlobalOpts(t, relativeExtensionDir)
	defer cleanup()

	// "-e docker_exit=3" is a special instruction handled in the fake docker script
	toolchainOptions := "-e docker_exit=3"

	// Run "getenvoy extension clean"
	c, stdout, stderr := cmd.NewRootCommand(o)
	c.SetArgs([]string{"extension", "clean", "--toolchain-container-options", toolchainOptions})
	err := rootcmd.Execute(c)

	// We expect the exit instruction to have gotten to the fake docker script, along with the default options.
	expectedDockerArgs := fmt.Sprintf(`run -u %s --rm -e GETENVOY_GOOS=%s -t -v %s:/source -w /source --init %s getenvoy/extension-rust-builder:latest clean`,
		builtin.DefaultDockerUser, runtime.GOOS, o.ExtensionDir, toolchainOptions)
	expectedErr := fmt.Sprintf(`failed to clean build directory of Envoy extension using "default" toolchain: failed to execute an external command "%s %s": exit status 3`, o.DockerPath, expectedDockerArgs)
	require.EqualError(t, err, expectedErr, `expected an error running [%v]`, c)

	// We should see stdout because the docker script was invoked.
	expectedStdout := fmt.Sprintf("docker wd: %s\ndocker bin: %s\ndocker args: %s\n",
		o.ExtensionDir, o.DockerPath, expectedDockerArgs)
	require.Equal(t, expectedStdout, stdout.String(), `expected stdout running [%v]`, c)

	// We also expect "docker stderr" in the output for the same reason.
	expectedStderr := fmt.Sprintf("docker stderr\nError: %s\n\nRun 'getenvoy extension clean --help' for usage.\n", expectedErr)
	require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, c)
}
