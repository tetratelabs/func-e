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
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/otiai10/copy"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/manifest"
	"github.com/tetratelabs/getenvoy/pkg/test/cmd"
	manifesttest "github.com/tetratelabs/getenvoy/pkg/test/manifest"
	. "github.com/tetratelabs/getenvoy/pkg/test/morerequire"
	"github.com/tetratelabs/getenvoy/pkg/types"
	cmdutil "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

// relativeWorkspaceDir points to a usable pre-initialized workspace
const relativeWorkspaceDir = "testdata/workspace"

func TestGetEnvoyExtensionRunValidateFlag(t *testing.T) {
	type testCase struct {
		name        string
		args        []string
		expectedErr string
	}

	tempDir, closer := RequireNewTempDir(t)
	defer closer()

	// Create a fake envoy script so that we can verify execute bit is required.
	notExecutable := filepath.Join(tempDir, "envoy")
	err := os.WriteFile(notExecutable, []byte(`#!/bin/sh`), 0600)
	require.NoError(t, err, `couldn't create fake envoy script'`)

	tests := []testCase{
		{
			name:        "--envoy-options with imbalanced quotes",
			args:        []string{"--envoy-options", "imbalanced ' quotes"},
			expectedErr: `"imbalanced ' quotes" is not a valid command line string`,
		},
		{
			name:        "--envoy-path file doesn't exist",
			args:        []string{"--envoy-path", "non-existing-file"},
			expectedErr: `unable to find custom Envoy binary at "non-existing-file": stat non-existing-file: no such file or directory`,
		},
		{
			name:        "--envoy-path is a directory",
			args:        []string{"--envoy-path", "."},
			expectedErr: `unable to find custom Envoy binary at ".": there is a directory at a given path instead of a regular file`,
		},
		{
			name:        "--envoy-path not executable",
			args:        []string{"--envoy-path", notExecutable},
			expectedErr: fmt.Sprintf(`unable to find custom Envoy binary at "%s": file is not executable`, notExecutable),
		},
		{
			name:        "--envoy-version with invalid value",
			args:        []string{"--envoy-version", "???"},
			expectedErr: `envoy version is not valid: "???" is not a valid GetEnvoy reference. Expected format: <flavor>:<version>[/<platform>]`,
		},
		{
			name:        "--envoy-version and --envoy-path flags at the same time",
			args:        []string{"--envoy-version", "standard:1.17.1", "--envoy-path", "envoy"},
			expectedErr: `only one of flags '--envoy-version' and '--envoy-path' can be used at a time`,
		},
		{
			name:        "--extension-config-file file doesn't exist",
			args:        []string{"--extension-config-file", "non-existing-file"},
			expectedErr: `failed to read custom extension config from file "non-existing-file": open non-existing-file: no such file or directory`,
		},
		{
			name:        "--extension-config-file is a directory",
			args:        []string{"--extension-config-file", "."},
			expectedErr: `failed to read custom extension config from file ".": read .: is a directory`,
		},
		{
			name:        "--extension-file file doesn't exist",
			args:        []string{"--extension-file", "non-existing-file"},
			expectedErr: `unable to find a pre-built *.wasm file at "non-existing-file": stat non-existing-file: no such file or directory`,
		},
		{
			name:        "--extension-file is a directory",
			args:        []string{"--extension-file", "."},
			expectedErr: `unable to find a pre-built *.wasm file at ".": there is a directory at a given path instead of a regular file`,
		},
		{
			name:        "--toolchain-container-options with invalid value",
			args:        []string{"--toolchain-container-image", "?invalid value?"},
			expectedErr: `"?invalid value?" is not a valid image name: invalid reference format`,
		},
		{
			name:        "--toolchain-container-options with imbalanced quotes",
			args:        []string{"--toolchain-container-options", "imbalanced ' quotes"},
			expectedErr: `"imbalanced ' quotes" is not a valid command line string`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			// Run "getenvoy extension run" with the flags we are testing
			c, stdout, stderr := cmd.NewRootCommand()
			c.SetArgs(append([]string{"extension", "run"}, test.args...))
			err := cmdutil.Execute(c)
			require.EqualError(t, err, test.expectedErr, `expected an error running [%v]`, c)

			// Verify the command failed with the expected error
			require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
			expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy extension run --help' for usage.\n", test.expectedErr)
			require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, c)
		})
	}
}

func TestGetEnvoyExtensionRunFailsOutsideWorkspaceDirectory(t *testing.T) {
	// Change to a non-workspace dir
	config, cleanup := setupTest(t, relativeWorkspaceDir+"/..")
	defer cleanup()

	// Run "getenvoy extension run"
	c, stdout, stderr := cmd.NewRootCommand()
	c.SetArgs([]string{"extension", "run"})
	err := cmdutil.Execute(c)

	// Verify the command failed with the expected error
	expectedErr := "there is no extension directory at or above: " + config.workspaceDir
	require.EqualError(t, err, expectedErr, `expected an error running [%v]`, c)
	require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
	expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy extension run --help' for usage.\n", expectedErr)
	require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, c)
}

func TestGetEnvoyExtensionRun(t *testing.T) {
	config, cleanup := setupTest(t, relativeWorkspaceDir)
	defer cleanup()

	// Run "getenvoy extension run"
	c, stdout, stderr := cmd.NewRootCommand()
	c.SetArgs([]string{"extension", "run", "--home-dir", config.envoyHome,
		// prevents generated admin-address-path which makes assertions difficult
		"--envoy-options", "--admin-address-path /admin-address.txt",
	})
	err := cmdutil.Execute(c)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, c)

	// The working directory of envoy isn't the same as docker or the workspace
	envoyWd := cmd.ParseEnvoyWorkDirectory(t, stdout.String(), `couldn't find envoy wd running [%v]`, c)

	envoyBin := filepath.Join(config.envoyHome, "builds/standard/1.17.1", config.platform, "/bin/envoy")
	// We expect docker to build from the correct path, as the current user and mount a volume for the correct workspace.
	expectedStdout := fmt.Sprintf(`%s/docker run -u %s --rm -e GETENVOY_GOOS=%s -t -v %s:/source -w /source --init getenvoy/extension-rust-builder:latest build --output-file target/getenvoy/extension.wasm
envoy pwd: %s
envoy bin: %s
envoy args: -c %s/envoy.tmpl.yaml --admin-address-path /admin-address.txt`,
		config.dockerDir, config.expectedUidGid, runtime.GOOS, config.workspaceDir, envoyWd, envoyBin, envoyWd)
	require.Equal(t, expectedStdout+"\n", stdout.String(), `expected stdout running [%v]`, c)
	require.Equal(t, "docker stderr\nenvoy stderr\n", stderr.String(), `expected stderr running [%v]`, c)

	// Verify the placeholders envoy would have ran substituted, including generated extension.wasm and escaped config.
	withoutSpace := expectedYAML(config.workspaceDir, false, `{"key":"value"}`)
	withSpace := expectedYAML(config.workspaceDir, true, `{"key":"value"}`)
	yaml := requirePlaceholdersYaml(t, config.envoyHome)
	if withoutSpace != yaml {
		require.Equal(t, yaml, withSpace, `unexpected placeholders yaml after running [%v]`, c)
	}
}

// TestGetEnvoyExtensionRunDockerFail ensures docker failures show useful information in stderr
func TestGetEnvoyExtensionRunDockerFail(t *testing.T) {
	config, cleanup := setupTest(t, relativeWorkspaceDir)
	defer cleanup()

	// "-e DOCKER_EXIT_CODE=3" is a special instruction handled in the fake docker script
	toolchainOptions := "-e DOCKER_EXIT_CODE=3"
	// Run "getenvoy extension run"
	c, _, stderr := cmd.NewRootCommand()
	c.SetArgs([]string{"extension", "run", "--toolchain-container-options", toolchainOptions})
	err := cmdutil.Execute(c)

	// We expect the exit instruction to have gotten to the fake docker script, along with the default options.
	expectedDockerExec := fmt.Sprintf("%s/docker run -u %s --rm -e GETENVOY_GOOS=%s -t -v %s:/source -w /source --init %s getenvoy/extension-rust-builder:latest build --output-file target/getenvoy/extension.wasm",
		config.dockerDir, config.expectedUidGid, runtime.GOOS, config.workspaceDir, toolchainOptions)

	// Verify the command failed with the expected error.
	expectedErr := fmt.Sprintf(`failed to build Envoy extension using "default" toolchain: failed to execute an external command "%s": exit status 3`, expectedDockerExec)
	require.EqualError(t, err, expectedErr, `expected an error running [%v]`, c)

	// We also expect "docker stderr" in the output for the same reason.
	expectedStderr := fmt.Sprintf("docker stderr\nError: %s\n\nRun 'getenvoy extension run --help' for usage.\n", expectedErr)
	require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, c)
}

// TestGetEnvoyExtensionRunWithExplicitVersion only tests that a version override is used. It doesn't test things that
// aren't different between here and TestGetEnvoyExtensionRun.
func TestGetEnvoyExtensionRunWithExplicitVersion(t *testing.T) {
	config, cleanup := setupTest(t, relativeWorkspaceDir)
	defer cleanup()

	// Run "getenvoy extension run --envoy-version wasm:stable"
	c, stdout, _ := cmd.NewRootCommand()
	c.SetArgs([]string{"extension", "run", "--home-dir", config.envoyHome, "--envoy-version", "wasm:stable"})
	err := cmdutil.Execute(c)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, c)

	// verify the expected binary is in the command output
	envoyBin := filepath.Join(config.envoyHome, "builds/wasm/stable", config.platform, "/bin/envoy")
	require.Contains(t, stdout.String(), "envoy bin: "+envoyBin, `expected stdout running [%v]`, c)
}

func TestGetEnvoyExtensionRunFailWithUnknownVersion(t *testing.T) {
	config, cleanup := setupTest(t, relativeWorkspaceDir)
	defer cleanup()

	version := "wasm:unknown"
	// Run "getenvoy extension run --envoy-version wasm:unknown"
	c, _, stderr := cmd.NewRootCommand()
	c.SetArgs([]string{"extension", "run", "--home-dir", config.envoyHome, "--envoy-version", version})
	err := cmdutil.Execute(c)

	// Verify the command failed with the expected error.
	reference := version + "/" + config.platform
	expectedErr := fmt.Sprintf(`failed to run "default" example: unable to find matching GetEnvoy build for reference "%s"`, reference)
	require.EqualError(t, err, expectedErr, `expected an error running [%v]`, c)

	// We also expect "docker stderr" in the output for the same reason.
	expectedStderr := fmt.Sprintf("docker stderr\nError: %s\n\nRun 'getenvoy extension run --help' for usage.\n", expectedErr)
	require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, c)
}

func TestGetEnvoyExtensionRunWithCustomBinary(t *testing.T) {
	config, cleanup := setupTest(t, relativeWorkspaceDir)
	defer cleanup()

	// Run "getenvoy extension run --envoy-path $ENVOY_HOME/bin/envoy"
	envoyBin := filepath.Join(config.envoyHome, "bin/envoy")
	c, stdout, _ := cmd.NewRootCommand()
	c.SetArgs([]string{"extension", "run", "--envoy-path", envoyBin})
	err := cmdutil.Execute(c)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, c)

	// The only way we can see "envoy bin: ..." in stdout, is if our fake envoy script was read
	require.Contains(t, stdout.String(), "envoy bin: "+envoyBin, `expected stdout running [%v]`, c)
}

func TestGetEnvoyExtensionRunWithOptions(t *testing.T) {
	config, cleanup := setupTest(t, relativeWorkspaceDir)
	defer cleanup()

	// Run "getenvoy extension run ..."
	c, stdout, _ := cmd.NewRootCommand()
	c.SetArgs([]string{"extension", "run", "--home-dir", config.envoyHome,
		"--envoy-options", "'--concurrency 2 --component-log-level wasm:debug,config:trace'"})
	err := cmdutil.Execute(c)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, c)

	// The working directory of envoy is a temp directory not controlled by this test, so we have to parse it.
	envoyWd := cmd.ParseEnvoyWorkDirectory(t, stdout.String(), `couldn't find envoy wd running [%v]`, c)

	envoyArgs := fmt.Sprintf(`-c %s/envoy.tmpl.yaml --concurrency 2 --component-log-level wasm:debug,config:trace`, envoyWd)
	require.Contains(t, stdout.String(), "envoy args: "+envoyArgs, `expected stdout running [%v]`, c)
}

// TestGetEnvoyExtensionRunWithWasm shows docker isn't run when the user supplies an "--extension-file"
func TestGetEnvoyExtensionRunWithWasm(t *testing.T) {
	config, cleanup := setupTest(t, relativeWorkspaceDir)
	defer cleanup()

	// As all scripts invoked are fakes, we only need to touch a file as it isn't read
	wasmFile := filepath.Join(config.tempDir, "extension.wasm")
	err := os.WriteFile(wasmFile, []byte{}, 0600)
	require.NoError(t, err, `expected no error creating extension.wasm: %s`, wasmFile)

	// Run "getenvoy extension run --extension-file /path/to/extension.wasm"
	c, stdout, stderr := cmd.NewRootCommand()
	c.SetArgs([]string{"extension", "run", "--home-dir", config.envoyHome, "--extension-file", wasmFile,
		// prevents generated admin-address-path which makes assertions difficult
		"--envoy-options", "--admin-address-path /admin-address.txt",
	})
	err = cmdutil.Execute(c)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, c)

	envoyBin := filepath.Join(config.envoyHome, "builds/standard/1.17.1", config.platform, "/bin/envoy")

	// The working directory of envoy is a temp directory not controlled by this test, so we have to parse it.
	envoyWd := cmd.ParseEnvoyWorkDirectory(t, stdout.String(), `couldn't find envoy wd running [%v]`, c)

	// We expect docker to not have ran, since we supplied a pre-existing wasm. However, envoy should have.
	expectedStdout := fmt.Sprintf(`envoy pwd: %s
envoy bin: %s
envoy args: -c %s/envoy.tmpl.yaml --admin-address-path /admin-address.txt`,
		envoyWd, envoyBin, envoyWd)
	require.Equal(t, expectedStdout+"\n", stdout.String(), `expected stdout running [%v]`, c)
	require.Equal(t, "envoy stderr\n", stderr.String(), `expected stderr running [%v]`, c)

	// Verify the placeholders envoy would have ran substituted, notably including the specified extension.wasm
	yamlExtensionCode := fmt.Sprintf(`'extension.code': {"local":{"filename":"%s"}}`, wasmFile)
	yaml := requirePlaceholdersYaml(t, config.envoyHome)
	require.Contains(t, yaml, yamlExtensionCode, `unexpected placeholders yaml after running [%v]`, c)
}

// TestGetEnvoyExtensionRunWithConfig shows extension config passed as an argument ends up readable by envoy.
func TestGetEnvoyExtensionRunWithConfig(t *testing.T) {
	config, cleanup := setupTest(t, relativeWorkspaceDir)
	defer cleanup()

	// As all scripts invoked are fakes, we only need to touch a file as it isn't read
	configFile := filepath.Join(config.tempDir, "config.json")
	err := os.WriteFile(configFile, []byte(`{"key2":"value2"}`), 0600)
	require.NoError(t, err, `expected no error creating extension.wasm: %s`, configFile)

	// Run "getenvoy extension run --extension-config-file /path/to/config.json"
	c, _, _ := cmd.NewRootCommand()
	c.SetArgs([]string{"extension", "run", "--home-dir", config.envoyHome, "--extension-config-file", configFile})
	err = cmdutil.Execute(c)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, c)

	// Verify the placeholders envoy would have ran substituted, including generated extension.wasm and escaped config.
	withoutSpace := expectedYAML(config.workspaceDir, false, `{"key2":"value2"}`)
	withSpace := expectedYAML(config.workspaceDir, true, `{"key2":"value2"}`)
	yaml := requirePlaceholdersYaml(t, config.envoyHome)
	if withoutSpace != yaml {
		require.Equal(t, yaml, withSpace, `unexpected placeholders yaml after running [%v]`, c)
	}
}

// Google made json formatting (json.prepareNext) intentionally unstable, technically by adding a space randomly.
// https://github.com/golang/protobuf/issues/920 requested an option for stability, but it was closed and locked.
// https://github.com/golang/protobuf/issues/1121 remains open, but unlikely to change.
// Hence, we have to check two possible formats via the shouldSpace parameter.
func expectedYAML(workspaceDir string, shouldSpace bool, extensionConfigValue string) string {
	space := ""
	if shouldSpace {
		space = " "
	}
	// Verify the placeholders envoy would have ran substituted, notably including the generated extension.wasm
	return fmt.Sprintf(`'extension.name': "mycompany.filters.http.custom_metrics"
'extension.code': {"local":{"filename":"%s/target/getenvoy/extension.wasm"}}
'extension.config': {"@type":"type.googleapis.com/google.protobuf.StringValue",%s"value":%q}
`, workspaceDir, space, extensionConfigValue)
}

func TestGetEnvoyExtensionRunCreatesExampleWhenMissing(t *testing.T) {
	// Use the workspace from the "extension build" test as it doesn't include examples.
	config, cleanup := setupTest(t, "../build/testdata/workspace")
	defer cleanup()

	// Run "getenvoy extension run"
	c, _, stderr := cmd.NewRootCommand()
	c.SetArgs([]string{"extension", "run", "--home-dir", config.envoyHome})
	err := cmdutil.Execute(c)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, c)

	// Verify a new example was scaffolded prior to running docker and envoy
	require.Equal(t, `Scaffolding a new example setup:
* .getenvoy/extension/examples/default/README.md
* .getenvoy/extension/examples/default/envoy.tmpl.yaml
* .getenvoy/extension/examples/default/example.yaml
* .getenvoy/extension/examples/default/extension.json
Done!
docker stderr
envoy stderr
`, stderr.String(), `expected stderr running [%v]`, c)
}

// TestGetEnvoyExtensionRunTinyGo ensures the docker command isn't pinned to rust projects
func TestGetEnvoyExtensionRunTinyGo(t *testing.T) {
	config, cleanup := setupTest(t, "testdata/workspace_tinygo")
	defer cleanup()

	// Run "getenvoy extension run"
	c, stdout, _ := cmd.NewRootCommand()
	c.SetArgs([]string{"extension", "run", "--home-dir", config.envoyHome})
	err := cmdutil.Execute(c)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, c)

	// Verify the docker command used the tinygo instead of the rust builder image
	require.Contains(t, stdout.String(), `--init getenvoy/extension-tinygo-builder:latest build`, `expected stdout running [%v]`, c)
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

	tempDir, deleteTempDir := RequireNewTempDir(t)
	tearDown = append(tearDown, deleteTempDir)
	result.tempDir = tempDir

	// We use a fake docker command to capture the commandline that would be invoked
	dockerDir, revertPath := RequireOverridePath(t, cmd.FakeDockerDir)
	tearDown = append(tearDown, revertPath)
	result.dockerDir = dockerDir

	envoyHome := filepath.Join(tempDir, "envoy_home")
	cmd.InitFakeEnvoyHome(t, envoyHome)
	result.envoyHome = envoyHome

	// create a new workspaceDir under tempDir
	workspaceDir := filepath.Join(tempDir, "workspace")
	err := os.Mkdir(workspaceDir, 0700)
	require.NoError(t, err, `error creating directory: %s`, workspaceDir)

	// Copy the template into the new workspaceDir to avoid tainting the source tree
	err = copy.Copy(RequireAbsDir(t, relativeWorkspaceTemplate), workspaceDir)
	require.NoError(t, err, `expected no error copying the directory: %s`, relativeWorkspaceTemplate)
	result.workspaceDir = workspaceDir

	// "getenvoy extension run" must be executed inside a valid workspace directory
	_, revertWd := RequireChDir(t, workspaceDir)
	tearDown = append(tearDown, revertWd)

	platform := requireManifestPlatform(t)
	shutdownTestServer := requireManifestTestServer(t, envoyHome)
	tearDown = append(tearDown, shutdownTestServer)
	result.platform = platform

	// Fake the current user so we can test it is used in the docker args
	expectedUser := user.User{Uid: "1001", Gid: "1002"}
	revertGetCurrentUser := cmd.OverrideGetCurrentUser(&expectedUser)
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
	b, err := os.ReadFile(placeholders)
	require.NoError(t, err, `expected no error reading placeholders: %s`, placeholders)
	return string(b)
}

// requireManifestPlatform returns the current platform as used in manifests.
func requireManifestPlatform(t *testing.T) string {
	key, err := manifest.NewKey("standard:1.17.1")
	require.NoError(t, err, `error resolving manifest for key: %s`, key)
	return key.Platform
}

// requireManifestTestServer calls manifest.SetURL to a test new tests server.
// The function returned stops that server and calls manifest.SetURL with the original URL.
func requireManifestTestServer(t *testing.T, envoySubstituteArchiveDir string) func() {
	testManifest, err := manifesttest.NewSimpleManifest("standard:1.17.1", "wasm:1.15", "wasm:stable")

	require.NoError(t, err, `error creating test manifest`)

	manifestServer := manifesttest.NewServer(&manifesttest.ServerOpts{
		Manifest: testManifest,
		GetArtifactDir: func(uri string) (string, error) {
			ref, e := types.ParseReference(uri)
			if e != nil {
				return "", e
			}
			if ref.Flavor == "wasm" {
				return envoySubstituteArchiveDir, nil
			}
			if ref.Flavor == "standard" {
				ver, e := semver.NewVersion(ref.Version)
				if e == nil && ver.Major() >= 1 && ver.Minor() >= 17 {
					return envoySubstituteArchiveDir, nil
				}
			}
			return "", fmt.Errorf("unexpected version of Envoy %q", uri)
		},
		OnError: func(err error) {
			require.NoError(t, err, `unexpected error from test manifest server`)
		},
	})

	// override location of the GetEnvoy manifest
	previous := manifest.GetURL()
	u := manifestServer.GetManifestURL()
	err = manifest.SetURL(u)
	require.NoError(t, err, `error manifest URL to: %s`, u)

	return func() {
		e := manifest.SetURL(previous)
		manifestServer.Close() // before require to ensure this occurs
		require.NoError(t, e, `error reverting manifest URL to: %s`, previous)
	}
}
