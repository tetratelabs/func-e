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
	"path/filepath"
	"runtime"
	"testing"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/require"

	reference "github.com/tetratelabs/getenvoy/pkg"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy/globals"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoytest"
	rootcmd "github.com/tetratelabs/getenvoy/pkg/cmd"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain/builtin"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
	"github.com/tetratelabs/getenvoy/pkg/test/cmd"
	manifesttest "github.com/tetratelabs/getenvoy/pkg/test/manifest"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

// relativeExtensionDir points to a usable pre-initialized workspace
const relativeExtensionDir = "testdata/workspace"

func TestGetEnvoyExtensionRunValidateFlag(t *testing.T) {
	type testCase struct {
		name        string
		args        []string
		expectedErr string
	}
	tests := []testCase{
		{
			name:        "--envoy-options with imbalanced quotes",
			args:        []string{"--envoy-options", "imbalanced ' quotes"},
			expectedErr: `"imbalanced ' quotes" is not a valid command line string`,
		},
		{
			name:        "--envoy-version with invalid value",
			args:        []string{"--envoy-version", "???"},
			expectedErr: `envoy version is not valid: "???" is not a valid GetEnvoy reference. Expected format: <flavor>:<version>[/<platform>]`,
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
			c, stdout, stderr := cmd.NewRootCommand(&globals.GlobalOpts{})
			c.SetArgs(append([]string{"extension", "run"}, test.args...))
			err := rootcmd.Execute(c)

			// Verify the command failed with the expected error
			require.EqualError(t, err, test.expectedErr, `expected an error running [%v]`, c)
			require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
			expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy extension run --help' for usage.\n", test.expectedErr)
			require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, c)
		})
	}
}

func TestGetEnvoyExtensionRunFailsOutsideExtensionDirectory(t *testing.T) {
	// Change to a non-workspace dir
	o := &globals.GlobalOpts{ExtensionDir: morerequire.RequireAbs(t, relativeExtensionDir+"/..")}

	// Run "getenvoy extension run"
	c, stdout, stderr := cmd.NewRootCommand(o)
	c.SetArgs([]string{"extension", "run"})
	err := rootcmd.Execute(c)

	// Verify the command failed with the expected error
	expectedErr := fmt.Sprintf(`not an extension directory %q`, o.ExtensionDir)
	require.EqualError(t, err, expectedErr, `expected an error running [%v]`, c)
	require.Empty(t, stdout.String(), `expected no stdout running [%v]`, c)
	expectedStderr := fmt.Sprintf("Error: %s\n\nRun 'getenvoy extension run --help' for usage.\n", expectedErr)
	require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, c)
}

func TestGetEnvoyExtensionRun(t *testing.T) {
	o, cleanup := setupTest(t, relativeExtensionDir)
	defer cleanup()

	// Run "getenvoy extension run"
	c, stdout, stderr := cmd.NewRootCommand(&o.GlobalOpts)
	c.SetArgs([]string{"extension", "run"})
	err := rootcmd.Execute(c)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, c)

	// We expect docker to build from the correct path, as the current user and mount a volume for the correct workspace.
	// Then, we expect envoy to run the generated configuration.
	expectedStdout := fmt.Sprintf(`docker wd: %s
docker bin: %s
docker args: run -u %s --rm -e GETENVOY_GOOS=%s -t -v %s:/source -w /source --init getenvoy/extension-rust-builder:latest build --output-file target/getenvoy/extension.wasm
envoy wd: %s
envoy bin: %s
envoy args: -c envoy.yaml --admin-address-path admin-address.txt`,
		o.ExtensionDir, o.DockerPath, builtin.DefaultDockerUser, runtime.GOOS, o.ExtensionDir, o.WorkingDir, o.EnvoyPath)
	require.Equal(t, expectedStdout+"\n", stdout.String(), `expected stdout running [%v]`, c)
	require.Equal(t, "docker stderr\nenvoy stderr\n", stderr.String(), `expected stderr running [%v]`, c)

	// Verify the placeholders envoy would have ran substituted, including generated extension.wasm and escaped o.
	envoytest.RequireRestoreWorkingDir(t, o.WorkingDir, c)
	withoutSpace := expectedYAML(o.ExtensionDir, false, `{"key":"value"}`)
	withSpace := expectedYAML(o.ExtensionDir, true, `{"key":"value"}`)
	yaml := requirePlaceholdersYaml(t, o.WorkingDir)
	if withoutSpace != yaml {
		require.Equal(t, yaml, withSpace, `unexpected placeholders yaml after running [%v]`, c)
	}
}

// TestGetEnvoyExtensionRunDockerFail ensures docker failures show useful information in stderr
func TestGetEnvoyExtensionRunDockerFail(t *testing.T) {
	o, cleanup := setupTest(t, relativeExtensionDir)
	defer cleanup()

	// "-e docker_exit=3" is a special instruction handled in the fake docker script
	toolchainOptions := "-e docker_exit=3"
	// Run "getenvoy extension run"
	c, stdout, stderr := cmd.NewRootCommand(&o.GlobalOpts)
	c.SetArgs([]string{"extension", "run", "--toolchain-container-options", toolchainOptions})
	err := rootcmd.Execute(c)

	// We expect the exit instruction to have gotten to the fake docker script, along with the default options.
	expectedDockerArgs := fmt.Sprintf(`run -u %s --rm -e GETENVOY_GOOS=%s -t -v %s:/source -w /source --init %s getenvoy/extension-rust-builder:latest build --output-file target/getenvoy/extension.wasm`,
		builtin.DefaultDockerUser, runtime.GOOS, o.ExtensionDir, toolchainOptions)
	expectedErr := fmt.Sprintf(`failed to build Envoy extension using "default" toolchain: failed to execute an external command "%s %s": exit status 3`, o.DockerPath, expectedDockerArgs)
	require.EqualError(t, err, expectedErr, `expected an error running [%v]`, c)

	// We should see stdout because the docker script was invoked.
	expectedStdout := fmt.Sprintf("docker wd: %s\ndocker bin: %s\ndocker args: %s\n",
		o.ExtensionDir, o.DockerPath, expectedDockerArgs)
	require.Equal(t, expectedStdout, stdout.String(), `expected stdout running [%v]`, c)

	// We also expect "docker stderr" in the output for the same reason.
	expectedStderr := fmt.Sprintf("docker stderr\nError: %s\n\nRun 'getenvoy extension run --help' for usage.\n", expectedErr)
	require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, c)
}

// TestGetEnvoyExtensionRunWithExplicitVersion only tests that a version override is used. It doesn't test things that
// aren't different between here and TestGetEnvoyExtensionRun.
func TestGetEnvoyExtensionRunWithExplicitVersion(t *testing.T) {
	o, cleanup := setupTest(t, relativeExtensionDir)
	defer cleanup()

	// Run "getenvoy extension run --envoy-version wasm:stable"
	c, stdout, _ := cmd.NewRootCommand(&o.GlobalOpts)
	c.SetArgs([]string{"extension", "run", "--envoy-version", "wasm:stable"})
	err := rootcmd.Execute(c)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, c)

	// verify the expected binary is in the command output
	require.Contains(t, stdout.String(), "envoy bin: "+o.EnvoyPath, `expected stdout running [%v]`, c)
}

func TestGetEnvoyExtensionRunFailWithUnknownVersion(t *testing.T) {
	o, cleanup := setupTest(t, relativeExtensionDir)
	defer cleanup()

	o.EnvoyPath = "" // force lookup of version flag
	c, _, stderr := cmd.NewRootCommand(&o.GlobalOpts)

	// Run "getenvoy extension run --envoy-version wasm:unknown"
	version := "wasm:unknown"
	c.SetArgs([]string{"extension", "run", "--envoy-version", version})
	err := rootcmd.Execute(c)

	// Verify the command failed with the expected error.
	r := version + "/" + o.platform
	expectedErr := fmt.Sprintf(`unable to find matching GetEnvoy build for reference "%s"`, r)
	require.EqualError(t, err, expectedErr, `expected an error running [%v]`, c)

	// We also expect "docker stderr" in the output for the same reason.
	expectedStderr := fmt.Sprintf("docker stderr\nError: %s\n\nRun 'getenvoy extension run --help' for usage.\n", expectedErr)
	require.Equal(t, expectedStderr, stderr.String(), `expected stderr running [%v]`, c)
}

func TestGetEnvoyExtensionRunWithOptions(t *testing.T) {
	o, cleanup := setupTest(t, relativeExtensionDir)
	defer cleanup()

	// Run "getenvoy extension run ..."
	c, stdout, _ := cmd.NewRootCommand(&o.GlobalOpts)
	envoyOpts := "--concurrency 2 --component-log-level wasm:debug,o:trace"
	c.SetArgs([]string{"extension", "run", "--envoy-options", fmt.Sprintf("'%s'", envoyOpts)})
	err := rootcmd.Execute(c)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, c)

	envoyArgs := fmt.Sprintf("-c envoy.yaml %s --admin-address-path admin-address.txt", envoyOpts)
	require.Contains(t, stdout.String(), "envoy args: "+envoyArgs, `expected stdout running [%v]`, c)
}

// TestGetEnvoyExtensionRunWithWasm shows docker isn't run when the user supplies an "--extension-file"
func TestGetEnvoyExtensionRunWithWasm(t *testing.T) {
	o, cleanup := setupTest(t, relativeExtensionDir)
	defer cleanup()

	// As all scripts invoked are fakes, we only need to touch a file as it isn't read
	wasmFile := filepath.Join(o.tempDir, "extension.wasm")
	err := os.WriteFile(wasmFile, []byte{}, 0600)
	require.NoError(t, err, `expected no error creating extension.wasm: %s`, wasmFile)

	c, stdout, stderr := cmd.NewRootCommand(&o.GlobalOpts)

	// Run "getenvoy extension run --extension-file /path/to/extension.wasm"
	c.SetArgs([]string{"extension", "run", "--extension-file", wasmFile})
	err = rootcmd.Execute(c)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, c)

	// We expect docker to not have ran, since we supplied a pre-existing wasm. However, envoy should have.
	expectedStdout := fmt.Sprintf(`envoy wd: %s
envoy bin: %s
envoy args: -c envoy.yaml --admin-address-path admin-address.txt`,
		o.WorkingDir, o.EnvoyPath)
	require.Equal(t, expectedStdout+"\n", stdout.String(), `expected stdout running [%v]`, c)
	require.Equal(t, "envoy stderr\n", stderr.String(), `expected stderr running [%v]`, c)

	// Verify the placeholders envoy would have ran substituted, notably including the specified extension.wasm
	envoytest.RequireRestoreWorkingDir(t, o.WorkingDir, c)
	yaml := requirePlaceholdersYaml(t, o.WorkingDir)
	yamlExtensionCode := fmt.Sprintf(`'extension.code': {"local":{"filename":"%s"}}`, wasmFile)
	require.Contains(t, yaml, yamlExtensionCode, `unexpected placeholders yaml after running [%v]`, c)
}

// TestGetEnvoyExtensionRunWithConfig shows extension o passed as an argument ends up readable by envoy.
func TestGetEnvoyExtensionRunWithConfig(t *testing.T) {
	o, cleanup := setupTest(t, relativeExtensionDir)
	defer cleanup()

	// As all scripts invoked are fakes, we only need to touch a file as it isn't read
	configFile := filepath.Join(o.tempDir, "o.json")
	err := os.WriteFile(configFile, []byte(`{"key2":"value2"}`), 0600)
	require.NoError(t, err, `expected no error creating extension.wasm: %s`, configFile)

	c, _, _ := cmd.NewRootCommand(&o.GlobalOpts)

	// Run "getenvoy extension run --extension-config-file /path/to/o.json"
	c.SetArgs([]string{"extension", "run", "--extension-config-file", configFile})
	err = rootcmd.Execute(c)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, c)

	// Verify the placeholders envoy would have ran substituted, including generated extension.wasm and escaped o.
	envoytest.RequireRestoreWorkingDir(t, o.WorkingDir, c)
	withoutSpace := expectedYAML(o.ExtensionDir, false, `{"key2":"value2"}`)
	withSpace := expectedYAML(o.ExtensionDir, true, `{"key2":"value2"}`)
	yaml := requirePlaceholdersYaml(t, o.WorkingDir)
	if withoutSpace != yaml {
		require.Equal(t, yaml, withSpace, `unexpected placeholders yaml after running [%v]`, c)
	}
}

// Google made json formatting (json.prepareNext) intentionally unstable, technically by adding a space randomly.
// https://github.com/golang/protobuf/issues/920 requested an option for stability, but it was closed and locked.
// https://github.com/golang/protobuf/issues/1121 remains open, but unlikely to change.
// Hence, we have to check two possible formats via the shouldSpace parameter.
func expectedYAML(extensionDir string, shouldSpace bool, extensionConfigValue string) string {
	space := ""
	if shouldSpace {
		space = " "
	}
	// Verify the placeholders envoy would have ran substituted, notably including the generated extension.wasm
	return fmt.Sprintf(`'extension.name': "mycompany.filters.http.custom_metrics"
'extension.code': {"local":{"filename":"%s/target/getenvoy/extension.wasm"}}
'extension.config': {"@type":"type.googleapis.com/google.protobuf.StringValue",%s"value":%q}
`, extensionDir, space, extensionConfigValue)
}

func TestGetEnvoyExtensionRunCreatesExampleWhenMissing(t *testing.T) {
	// Use the workspace from the "extension build" test as it doesn't include examples.
	o, cleanup := setupTest(t, "../build/testdata/workspace")
	defer cleanup()

	c, _, stderr := cmd.NewRootCommand(&o.GlobalOpts)

	// Run "getenvoy extension run"
	c.SetArgs([]string{"extension", "run"})
	err := rootcmd.Execute(c)

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
	o, cleanup := setupTest(t, "testdata/workspace_tinygo")
	defer cleanup()

	c, stdout, _ := cmd.NewRootCommand(&o.GlobalOpts)

	// Run "getenvoy extension run"
	c.SetArgs([]string{"extension", "run"})
	err := rootcmd.Execute(c)

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err, `expected no error running [%v]`, c)

	// Verify the docker command used the tinygo instead of the rust builder image
	require.Contains(t, stdout.String(), `--init getenvoy/extension-tinygo-builder:latest build`, `expected stdout running [%v]`, c)
}

type testEnvoyExtensionConfig struct {
	globals.GlobalOpts
	// tempDir is deleted on exit and contains many of the other directories
	tempDir string
	// platform is the types.Reference.Platform used in manifest commands
	platform string
}

// setupTest returns testEnvoyExtensionConfig and a tear-down function.
// The tear-down functions reverts side-effects such as temp directories and a fake manifest server.
// relativeWorkspaceTemplate is relative to the test file and will be copied into the resulting o.ExtensionDir.
func setupTest(t *testing.T, relativeWorkspaceTemplate string) (*testEnvoyExtensionConfig, func()) {
	result := testEnvoyExtensionConfig{}
	var tearDown []func()

	tempDir, deleteTempDir := morerequire.RequireNewTempDir(t)
	tearDown = append(tearDown, deleteTempDir)
	result.tempDir = tempDir

	// We use a fake docker command to capture the commandline that would be invoked
	fakeDocker, removeFakeDocker := morerequire.RequireCaptureScript(t, "docker")
	tearDown = append(tearDown, removeFakeDocker)
	result.DockerPath = fakeDocker

	result.HomeDir = filepath.Join(tempDir, "envoy_home")
	err := os.Mkdir(result.HomeDir, 0700)
	require.NoError(t, err, `error creating directory: %s`, result.HomeDir)

	// create a new ExtensionDir under tempDir
	result.ExtensionDir = filepath.Join(tempDir, "extension")
	err = os.Mkdir(result.ExtensionDir, 0700)
	require.NoError(t, err, `error creating directory: %s`, result.ExtensionDir)

	// Copy the template into the new ExtensionDir to avoid tainting the source tree
	err = copy.Copy(morerequire.RequireAbs(t, relativeWorkspaceTemplate), result.ExtensionDir)
	require.NoError(t, err, `expected no error copying the directory: %s`, relativeWorkspaceTemplate)

	key, err := manifest.NewKey(reference.Latest)
	require.NoError(t, err, `error resolving manifest for key: %s`, key)
	result.platform = key.Platform

	testManifest, err := manifesttest.NewSimpleManifest(key.String(), "wasm:1.15", "wasm:stable")
	require.NoError(t, err, `error creating test manifest`)

	manifestServer := manifesttest.RequireManifestTestServer(t, testManifest)
	result.ManifestURL = manifestServer.URL + "/manifest.json"
	tearDown = append(tearDown, manifestServer.Close)

	return &result, func() {
		for i := len(tearDown) - 1; i >= 0; i-- {
			tearDown[i]()
		}
	}
}

func requirePlaceholdersYaml(t *testing.T, debugDir string) string {
	placeholders := filepath.Join(debugDir, "placeholders.tmpl.yaml")
	b, err := os.ReadFile(placeholders)
	require.NoError(t, err, `expected no error reading placeholders: %s`, placeholders)
	return string(b)
}
