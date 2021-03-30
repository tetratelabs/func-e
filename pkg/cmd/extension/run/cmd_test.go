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

package run_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/test/cmd/extension"
	cmdutil "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

// relativeWorkspaceDir points to a usable pre-initialized workspace
const relativeWorkspaceDir = "testdata/workspace"

func TestGetEnvoyExtensionRunValidateFlag(t *testing.T) {
	type TestCase struct {
		name        string
		flags       []string
		flagValues  []string
		expectedErr string
	}

	tempDir, closer := extension.RequireNewTempDir(t)
	defer closer()

	// Create a fake envoy script so that we can verify execute bit is required.
	notExecutable := filepath.Join(tempDir, "envoy")
	err := ioutil.WriteFile(notExecutable, []byte(`#!/bin/sh`), 0600)
	require.NoError(t, err, `couldn't create fake envoy script'`)

	tests := []TestCase{
		{
			name:        "--envoy-options with imbalanced quotes",
			flags:       []string{"--envoy-options"},
			flagValues:  []string{"imbalanced ' quotes"},
			expectedErr: `"imbalanced ' quotes" is not a valid command line string`,
		},
		{
			name:        "--envoy-path file doesn't exist",
			flags:       []string{"--envoy-path"},
			flagValues:  []string{"non-existing-file"},
			expectedErr: `unable to find custom Envoy binary at "non-existing-file": stat non-existing-file: no such file or directory`,
		},
		{
			name:        "--envoy-path is a directory",
			flags:       []string{"--envoy-path"},
			flagValues:  []string{"."},
			expectedErr: `unable to find custom Envoy binary at ".": there is a directory at a given path instead of a regular file`,
		},
		{
			name:        "--envoy-path not executable",
			flags:       []string{"--envoy-path"},
			flagValues:  []string{notExecutable},
			expectedErr: fmt.Sprintf(`unable to find custom Envoy binary at "%s": file is not executable`, notExecutable),
		},
		{
			name:        "--envoy-version with invalid value",
			flags:       []string{"--envoy-version"},
			flagValues:  []string{"???"},
			expectedErr: `Envoy version is not valid: "???" is not a valid GetEnvoy reference. Expected format: <flavor>:<version>[/<platform>]`,
		},
		{
			name:        "--envoy-version and --envoy-path flags at the same time",
			flags:       []string{"--envoy-version", "--envoy-path"},
			flagValues:  []string{"standard:1.17.0", "envoy"},
			expectedErr: `only one of flags '--envoy-version' and '--envoy-path' can be used at a time`,
		},
		{
			name:        "--extension-config-file file doesn't exist",
			flags:       []string{"--extension-config-file"},
			flagValues:  []string{"non-existing-file"},
			expectedErr: `failed to read custom extension config from file "non-existing-file": open non-existing-file: no such file or directory`,
		},
		{
			name:        "--extension-config-file is a directory",
			flags:       []string{"--extension-config-file"},
			flagValues:  []string{"."},
			expectedErr: `failed to read custom extension config from file ".": read .: is a directory`,
		},
		{
			name:        "--extension-file file doesn't exist",
			flags:       []string{"--extension-file"},
			flagValues:  []string{"non-existing-file"},
			expectedErr: `unable to find a pre-built *.wasm file at "non-existing-file": stat non-existing-file: no such file or directory`,
		},
		{
			name:        "--extension-file is a directory",
			flags:       []string{"--extension-file"},
			flagValues:  []string{"."},
			expectedErr: `unable to find a pre-built *.wasm file at ".": there is a directory at a given path instead of a regular file`,
		},
		{
			name:        "--toolchain-container-options with invalid value",
			flags:       []string{"--toolchain-container-image"},
			flagValues:  []string{"?invalid value?"},
			expectedErr: `"?invalid value?" is not a valid image name: invalid reference format`,
		},
		{
			name:        "--toolchain-container-options with imbalanced quotes",
			flags:       []string{"--toolchain-container-options"},
			flagValues:  []string{"imbalanced ' quotes"},
			expectedErr: `"imbalanced ' quotes" is not a valid command line string`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			// Run "getenvoy extension run" with the flags we are testing
			cmd, stdout, stderr := extension.NewRootCommand()
			args := []string{"extension", "run"}
			for i := range test.flags {
				args = append(args, test.flags[i], test.flagValues[i])
			}
			cmd.SetArgs(args)
			err := cmdutil.Execute(cmd)
			require.EqualError(t, err, test.expectedErr, `expected an error running [%v]`, cmd)

			// Verify the command failed with the expected error
			require.Empty(t, stdout.String(), `expected no stdout running [%v]`, cmd)
			expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy extension run --help' for usage.\n", test.expectedErr)
			require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, cmd)
		})
	}
}

func TestGetEnvoyExtensionRunFailsOutsideWorkspaceDirectory(t *testing.T) {
	// Change to a non-workspace dir
	config, cleanup := setupTest(t, relativeWorkspaceDir+"/..")
	defer cleanup()

	// Run "getenvoy extension run"
	cmd, stdout, stderr := extension.NewRootCommand()
	cmd.SetArgs([]string{"extension", "run"})
	err := cmdutil.Execute(cmd)

	// Verify the command failed with the expected error
	expectedErr := "there is no extension directory at or above: " + config.workspaceDir
	require.EqualError(t, err, expectedErr, `expected an error running [%v]`, cmd)
	require.Empty(t, stdout.String(), `expected no stdout running [%v]`, cmd)
	expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy extension run --help' for usage.\n", expectedErr)
	require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, cmd)
}

func TestGetEnvoyExtensionRun(t *testing.T) {
	config, cleanup := setupTest(t, relativeWorkspaceDir)
	defer cleanup()

	// Run "getenvoy extension run"
	cmd, stdout, stderr := extension.NewRootCommand()
	cmd.SetArgs([]string{"extension", "run", "--home-dir", config.envoyHome})
	err := cmdutil.Execute(cmd)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, cmd)

	envoyBin := filepath.Join(config.envoyHome, "builds/standard/1.17.0", config.platform, "/bin/envoy")
	// The working directory of envoy isn't the same as docker or the workspace
	envoyWd := extension.ParseEnvoyWorkDirectory(stdout)

	// We expect docker to build from the correct path, as the current user and mount a volume for the correct workspace.
	expectedStdout := fmt.Sprintf(`%s/docker run -u %s --rm -t -v %s:/source -w /source --init getenvoy/extension-rust-builder:latest build --output-file target/getenvoy/extension.wasm
envoy pwd: %s
envoy bin: %s
envoy args: -c %s/envoy.tmpl.yaml`,
		config.dockerDir, config.expectedUidGid, config.workspaceDir, envoyWd, envoyBin, envoyWd)
	require.Equal(t, expectedStdout+"\n", stdout.String(), `expected stdout running [%v]`, cmd)
	require.Equal(t, "docker stderr\nenvoy stderr\n", stderr.String(), `expected stderr running [%v]`, cmd)

	// Verify the placeholders envoy would have ran substituted, notably including the generated extension.wasm
	expectedYaml := fmt.Sprintf(`'extension.name': "mycompany.filters.http.custom_metrics"
'extension.code': {"local":{"filename":"%s/target/getenvoy/extension.wasm"}}
'extension.config': {"@type":"type.googleapis.com/google.protobuf.StringValue","value":"{\"key\":\"value\"}"}
`, config.workspaceDir)
	yaml := requirePlaceholdersYaml(t, config.envoyHome)
	require.Equal(t, expectedYaml, yaml, `unexpected placeholders yaml after running [%v]`, cmd)
}

// TestGetEnvoyExtensionRunDockerFail ensures docker failures show useful information in stderr
func TestGetEnvoyExtensionRunDockerFail(t *testing.T) {
	config, cleanup := setupTest(t, relativeWorkspaceDir)
	defer cleanup()

	// "-e DOCKER_EXIT_CODE=3" is a special instruction handled in the fake docker script
	toolchainOptions := "-e DOCKER_EXIT_CODE=3"
	// Run "getenvoy extension run"
	cmd, _, stderr := extension.NewRootCommand()
	cmd.SetArgs([]string{"extension", "run", "--toolchain-container-options", toolchainOptions})
	err := cmdutil.Execute(cmd)

	// We expect the exit instruction to have gotten to the fake docker script, along with the default options.
	expectedDockerExec := fmt.Sprintf("%s/docker run -u %s --rm -t -v %s:/source -w /source --init %s getenvoy/extension-rust-builder:latest build --output-file target/getenvoy/extension.wasm",
		config.dockerDir, config.expectedUidGid, config.workspaceDir, toolchainOptions)

	// Verify the command failed with the expected error.
	expectedErr := fmt.Sprintf(`failed to build Envoy extension using "default" toolchain: failed to execute an external command "%s": exit status 3`, expectedDockerExec)
	require.EqualError(t, err, expectedErr, `expected an error running [%v]`, cmd)

	// We also expect "docker stderr" in the output for the same reason.
	expectedStderr := fmt.Sprintf("docker stderr\nError: %s\n\nRun 'getenvoy extension run --help' for usage.\n", expectedErr)
	require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, cmd)
}

// TestGetEnvoyExtensionRunWithExplicitVersion only tests that a version override is used. It doesn't test things that
// aren't different between here and TestGetEnvoyExtensionRun.
func TestGetEnvoyExtensionRunWithExplicitVersion(t *testing.T) {
	config, cleanup := setupTest(t, relativeWorkspaceDir)
	defer cleanup()

	// Run "getenvoy extension run --envoy-version wasm:stable"
	cmd, stdout, _ := extension.NewRootCommand()
	cmd.SetArgs([]string{"extension", "run", "--home-dir", config.envoyHome, "--envoy-version", "wasm:stable"})
	err := cmdutil.Execute(cmd)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, cmd)

	// verify the expected binary is in the command output
	envoyBin := filepath.Join(config.envoyHome, "builds/wasm/stable", config.platform, "/bin/envoy")
	require.Contains(t, stdout.String(), "envoy bin: "+envoyBin, `expected stdout running [%v]`, cmd)
}

func TestGetEnvoyExtensionRunFailWithUnknownVersion(t *testing.T) {
	config, cleanup := setupTest(t, relativeWorkspaceDir)
	defer cleanup()

	version := "wasm:unknown"
	// Run "getenvoy extension run --envoy-version wasm:unknown"
	cmd, _, stderr := extension.NewRootCommand()
	cmd.SetArgs([]string{"extension", "run", "--home-dir", config.envoyHome, "--envoy-version", version})
	err := cmdutil.Execute(cmd)

	// Verify the command failed with the expected error.
	reference := version + "/" + config.platform
	expectedErr := fmt.Sprintf(`failed to run "default" example: unable to find matching GetEnvoy build for reference "%s"`, reference)
	require.EqualError(t, err, expectedErr, `expected an error running [%v]`, cmd)

	// We also expect "docker stderr" in the output for the same reason.
	expectedStderr := fmt.Sprintf("docker stderr\nError: %s\n\nRun 'getenvoy extension run --help' for usage.\n", expectedErr)
	require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, cmd)
}

func TestGetEnvoyExtensionRunWithCustomBinary(t *testing.T) {
	config, cleanup := setupTest(t, relativeWorkspaceDir)
	defer cleanup()

	// Run "getenvoy extension run --envoy-path $ENVOY_HOME/bin/envoy"
	envoyBin := filepath.Join(config.envoyHome, "bin/envoy")
	cmd, stdout, _ := extension.NewRootCommand()
	cmd.SetArgs([]string{"extension", "run", "--envoy-path", envoyBin})
	err := cmdutil.Execute(cmd)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, cmd)

	// The only way we can see "envoy bin: ..." in stdout, is if our fake envoy script was read
	require.Contains(t, stdout.String(), "envoy bin: "+envoyBin, `expected stdout running [%v]`, cmd)
}

func TestGetEnvoyExtensionRunWithOptions(t *testing.T) {
	config, cleanup := setupTest(t, relativeWorkspaceDir)
	defer cleanup()

	// Run "getenvoy extension run ..."
	cmd, stdout, _ := extension.NewRootCommand()
	cmd.SetArgs([]string{"extension", "run", "--home-dir", config.envoyHome,
		"--envoy-options", "'--concurrency 2 --component-log-level wasm:debug,config:trace'"})
	err := cmdutil.Execute(cmd)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, cmd)

	// The working directory of envoy is a temp directory not controlled by this test, so we have to parse it.
	envoyWd := extension.ParseEnvoyWorkDirectory(stdout)

	envoyArgs := fmt.Sprintf(`-c %s/envoy.tmpl.yaml --concurrency 2 --component-log-level wasm:debug,config:trace`, envoyWd)
	require.Contains(t, stdout.String(), "envoy args: "+envoyArgs, `expected stdout running [%v]`, cmd)
}

// TestGetEnvoyExtensionRunWithWasm shows docker isn't run when the user supplies an "--extension-file"
func TestGetEnvoyExtensionRunWithWasm(t *testing.T) {
	config, cleanup := setupTest(t, relativeWorkspaceDir)
	defer cleanup()

	// As all scripts invoked are fakes, we only need to touch a file as it isn't read
	wasmFile := filepath.Join(config.tempDir, "extension.wasm")
	err := ioutil.WriteFile(wasmFile, []byte{}, 0600)
	require.NoError(t, err, `expected no error creating extension.wasm: %s`, wasmFile)

	// Run "getenvoy extension run --extension-file /path/to/extension.wasm"
	cmd, stdout, stderr := extension.NewRootCommand()
	cmd.SetArgs([]string{"extension", "run", "--home-dir", config.envoyHome, "--extension-file", wasmFile})
	err = cmdutil.Execute(cmd)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, cmd)

	envoyBin := filepath.Join(config.envoyHome, "builds/standard/1.17.0", config.platform, "/bin/envoy")

	// The working directory of envoy is a temp directory not controlled by this test, so we have to parse it.
	envoyWd := extension.ParseEnvoyWorkDirectory(stdout)

	// We expect docker to not have ran, since we supplied a pre-existing wasm. However, envoy should have.
	expectedStdout := fmt.Sprintf(`envoy pwd: %s
envoy bin: %s
envoy args: -c %s/envoy.tmpl.yaml`,
		envoyWd, envoyBin, envoyWd)
	require.Equal(t, expectedStdout+"\n", stdout.String(), `expected stdout running [%v]`, cmd)
	require.Equal(t, "envoy stderr\n", stderr.String(), `expected stderr running [%v]`, cmd)

	// Verify the placeholders envoy would have ran substituted, notably including the specified extension.wasm
	yamlExtensionCode := fmt.Sprintf(`'extension.code': {"local":{"filename":"%s"}}`, wasmFile)
	yaml := requirePlaceholdersYaml(t, config.envoyHome)
	require.Contains(t, yaml, yamlExtensionCode, `unexpected placeholders yaml after running [%v]`, cmd)
}

// TestGetEnvoyExtensionRunWithConfig shows extension config passed as an argument ends up readable by envoy.
func TestGetEnvoyExtensionRunWithConfig(t *testing.T) {
	config, cleanup := setupTest(t, relativeWorkspaceDir)
	defer cleanup()

	// As all scripts invoked are fakes, we only need to touch a file as it isn't read
	configFile := filepath.Join(config.tempDir, "config.json")
	err := ioutil.WriteFile(configFile, []byte(`{"key2":"value2"}`), 0600)
	require.NoError(t, err, `expected no error creating extension.wasm: %s`, configFile)

	// Run "getenvoy extension run --extension-config-file /path/to/config.json"
	cmd, _, _ := extension.NewRootCommand()
	cmd.SetArgs([]string{"extension", "run", "--home-dir", config.envoyHome, "--extension-config-file", configFile})
	err = cmdutil.Execute(cmd)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, cmd)

	// Verify the placeholders envoy would have ran substituted, notably including the escaped config
	yamlExtensionConfig := `'extension.config': {"@type":"type.googleapis.com/google.protobuf.StringValue","value":"{\"key2\":\"value2\"}"}`
	yaml := requirePlaceholdersYaml(t, config.envoyHome)
	require.Contains(t, yaml, yamlExtensionConfig, `unexpected placeholders yaml after running [%v]`, cmd)
}

func TestGetEnvoyExtensionRunCreatesExampleWhenMissing(t *testing.T) {
	// Use the workspace from the "extension build" test as it doesn't include examples.
	config, cleanup := setupTest(t, "../build/testdata/workspace")
	defer cleanup()

	// Run "getenvoy extension run"
	cmd, _, stderr := extension.NewRootCommand()
	cmd.SetArgs([]string{"extension", "run", "--home-dir", config.envoyHome})
	err := cmdutil.Execute(cmd)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, cmd)

	// Verify a new example was scaffolded prior to running docker and envoy
	require.Equal(t, `Scaffolding a new example setup:
* .getenvoy/extension/examples/default/README.md
* .getenvoy/extension/examples/default/envoy.tmpl.yaml
* .getenvoy/extension/examples/default/example.yaml
* .getenvoy/extension/examples/default/extension.json
Done!
docker stderr
envoy stderr
`, stderr.String(), `expected stderr running [%v]`, cmd)
}

// TestGetEnvoyExtensionRunTinyGo ensures the docker command isn't pinned to rust projects
func TestGetEnvoyExtensionRunTinyGo(t *testing.T) {
	config, cleanup := setupTest(t, "testdata/workspace_tinygo")
	defer cleanup()

	// Run "getenvoy extension run"
	cmd, stdout, _ := extension.NewRootCommand()
	cmd.SetArgs([]string{"extension", "run", "--home-dir", config.envoyHome})
	err := cmdutil.Execute(cmd)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, cmd)

	// Verify the docker command used the tinygo instead of the rust builder image
	require.Contains(t, stdout.String(), `--init getenvoy/extension-tinygo-builder:latest build`, `expected stdout running [%v]`, cmd)
}

type testEnvoyExtensionConfig struct {
	// tempDir is deleted on exit and contains many of the other directories
	tempDir string
	// dockerDir is the absolute location of extension.FakeDockerDir
	dockerDir string
	// workspaceDir will be the CWD of "getenvoy"
	workspaceDir string
	// envoyHome is a fake $tempDir/envoy_home, initialized with initFakeEnvoyHome
	envoyHome string
	// platform is the types.Reference.Platform used in manifest commands
	platform string
	// expectedUidGid corresponds to a fake user.User ex 1001:1002 the builtin toolchain will see.
	expectedUidGid string //nolint
}

// setupTest returns testEnvoyExtensionConfig and a tear-down function.
// The tear-down functions reverts side-effects such as temp directories and a fake manifest server.
// relativeWorkspaceTemplate is relative to the test file and will be copied into the resulting config.workspaceDir.
func setupTest(t *testing.T, relativeWorkspaceTemplate string) (*testEnvoyExtensionConfig, func()) {
	result := testEnvoyExtensionConfig{}
	var tearDown []func()

	tempDir, deleteTempDir := extension.RequireNewTempDir(t)
	tearDown = append(tearDown, deleteTempDir)
	result.tempDir = tempDir

	// We use a fake docker command to capture the commandline that would be invoked
	dockerDir, revertPath := extension.RequireOverridePath(t, extension.FakeDockerDir)
	tearDown = append(tearDown, revertPath)
	result.dockerDir = dockerDir

	envoyHome := filepath.Join(tempDir, "envoy_home")
	extension.InitFakeEnvoyHome(t, envoyHome)
	result.envoyHome = envoyHome

	// create a new workspaceDir under tempDir
	workspaceDir := filepath.Join(tempDir, "workspace")
	err := os.Mkdir(workspaceDir, 0700)
	require.NoError(t, err, `error creating directory: %s`, workspaceDir)

	// Copy the template into the new workspaceDir to avoid tainting the source tree
	err = copy.Copy(extension.RequireAbsDir(t, relativeWorkspaceTemplate), workspaceDir)
	require.NoError(t, err, `expected no error copying the directory: %s`, relativeWorkspaceTemplate)
	result.workspaceDir = workspaceDir

	// "getenvoy extension run" must be executed inside a valid workspace directory
	_, revertWd := extension.RequireChDir(t, workspaceDir)
	tearDown = append(tearDown, revertWd)

	platform := extension.RequireManifestPlatform(t)
	shutdownTestServer := extension.RequireManifestTestServer(t, envoyHome)
	tearDown = append(tearDown, shutdownTestServer)
	result.platform = platform

	// Fake the current user so we can test it is used in the docker args
	expectedUser := user.User{Uid: "1001", Gid: "1002"}
	revertGetCurrentUser := extension.OverrideGetCurrentUser(&expectedUser)
	tearDown = append(tearDown, revertGetCurrentUser)
	result.expectedUidGid = expectedUser.Uid + ":" + expectedUser.Gid

	return &result, func() {
		for i := len(tearDown) - 1; i >= 0; i-- {
			tearDown[i]()
		}
	}
}

func requirePlaceholdersYaml(t *testing.T, envoyHome string) string {
	placeholders := filepath.Join(envoyHome, "capture", "placeholders.tmpl.yaml")
	b, err := ioutil.ReadFile(placeholders)
	require.NoError(t, err, `expected no error reading placeholders: %s`, placeholders)
	return string(b)
}
