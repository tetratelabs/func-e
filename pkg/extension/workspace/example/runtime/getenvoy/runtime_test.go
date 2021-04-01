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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/example/runtime"
	. "github.com/tetratelabs/getenvoy/pkg/extension/workspace/example/runtime/getenvoy"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
	"github.com/tetratelabs/getenvoy/pkg/test/cmd"
	ioutil "github.com/tetratelabs/getenvoy/pkg/util/io"
)

const (
	relativeWorkspaceDir = "testdata/workspace"
	invalidWorkspaceDir  = "testdata/invalidWorkspace"
)

func TestRuntimeRun(t *testing.T) {
	workspace, err := workspaces.GetWorkspaceAt(relativeWorkspaceDir)
	require.NoError(t, err, `expected no error getting workspace from directory %s`, relativeWorkspaceDir)

	example, err := workspace.GetExample("default")
	require.NoError(t, err, `expected no error getting example from workspace %s`, workspace)

	fakeEnvoyPath, tearDown := setupFakeEnvoy(t)
	defer tearDown()

	// Create and run a new context that will invoke a fake envoy script
	ctx, stdout, stderr := runContext(workspace, example, fakeEnvoyPath)
	err = NewRuntime().Run(ctx)
	require.NoError(t, err, `expected no error running running [%v]`, ctx)

	// The working directory of envoy is a temp directory not controlled by this test, so we have to parse it.
	envoyWd := cmd.ParseEnvoyWorkDirectory(stdout)

	// Verify we executed the indicated envoy binary, and it captured the arguments we expected
	expectedStdout := fmt.Sprintf(`envoy pwd: %s
envoy bin: %s
envoy args: -c %s/envoy.tmpl.yaml
`, envoyWd, ctx.Opts.Envoy.Path, envoyWd)
	require.Equal(t, expectedStdout, stdout.String(), `expected stdout running [%v]`, ctx)

	// Verify we didn't accidentally combine the stderr of envoy into stdout, or otherwise dropped it.
	require.Equal(t, "envoy stderr\n", stderr.String(), `expected stderr running [%v]`, ctx)
}

func TestRuntimeRunFailsOnInvalidWorkspace(t *testing.T) {
	invalidWorkspaceDir := cmd.RequireAbsDir(t, invalidWorkspaceDir)
	workspace, err := workspaces.GetWorkspaceAt(invalidWorkspaceDir)
	require.NoError(t, err, `expected no error getting workspace from directory %s`, invalidWorkspaceDir)

	example, err := workspace.GetExample("default")
	require.NoError(t, err, `expected no error getting example from workspace %s`, workspace)

	fakeEnvoyPath, tearDown := setupFakeEnvoy(t)
	defer tearDown()

	// Create and run a new context that will invoke a fake envoy script
	ctx, stdout, stderr := runContext(workspace, example, fakeEnvoyPath)
	err = NewRuntime().Run(ctx)

	// Verify the error raised parsing the template from the input directory, before running envoy.
	invalidTemplate := invalidWorkspaceDir + "/.getenvoy/extension/examples/default/envoy.tmpl.yaml"
	expectedErr := fmt.Sprintf(`failed to process Envoy config template coming from "%s": failed to render Envoy config template: template: :18:19: executing "" at <.GetEnvoy.DefaultValue>: error calling DefaultValue: unknown property "???"`, invalidTemplate)
	require.EqualError(t, err, expectedErr, `expected an error running [%v]`, ctx)

	// Verify there was no stdout or stderr because envoy shouldn't have run, yet.
	require.Empty(t, stdout.String(), `expected no stdout running [%v]`, ctx)
	require.Empty(t, stderr.String(), `expected no stderr running [%v]`, ctx)
}

func runContext(workspace model.Workspace, example model.Example, envoyPath string) (ctx *runtime.RunContext, stdout, stderr *bytes.Buffer) {
	stdout = new(bytes.Buffer)
	stderr = new(bytes.Buffer)
	ctx = &runtime.RunContext{
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
				Path: envoyPath,
			},
		},
		IO: ioutil.StdStreams{
			Out: stdout,
			Err: stderr,
		},
	}
	return
}

// setupFakeEnvoy creates a fake envoy home and returns the path to the binary.
// Side effects are reversed in the returned tear-down function.
func setupFakeEnvoy(t *testing.T) (string, func()) {
	var tearDown []func()

	tempDir, deleteTempDir := cmd.RequireNewTempDir(t)
	tearDown = append(tearDown, deleteTempDir)

	envoyHome := filepath.Join(tempDir, "envoy_home")
	fakeEnvoyPath := cmd.InitFakeEnvoyHome(t, envoyHome)
	revertHomeDir := cmd.OverrideHomeDir(envoyHome)
	tearDown = append(tearDown, revertHomeDir)

	return fakeEnvoyPath, func() {
		for i := len(tearDown) - 1; i >= 0; i-- {
			tearDown[i]()
		}
	}
}
