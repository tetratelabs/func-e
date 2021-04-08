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

package configdir_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	envoybootstrap "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	"github.com/stretchr/testify/require"

	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/example/runtime"
	. "github.com/tetratelabs/getenvoy/pkg/extension/workspace/example/runtime/configdir"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
)

func TestNewConfigDir(t *testing.T) {
	tests := []struct {
		name             string
		workspaceDir     string
		isEnvoyTemplate  func(string) bool
		requireBootstrap func(t *testing.T, bootstrap *envoybootstrap.Bootstrap)
	}{
		{
			name:         "envoy.tmpl.yaml",
			workspaceDir: "testdata/workspace1",
			isEnvoyTemplate: func(name string) bool {
				return name == "envoy.tmpl.yaml" //nolint:goconst
			},
			requireBootstrap: expectValidBootstrap,
		},
		{
			name:         "envoy.tmpl.yaml + lds.tmpl.yaml + cds.yaml",
			workspaceDir: "testdata/workspace2",
			isEnvoyTemplate: func(name string) bool {
				return name == "envoy.tmpl.yaml" || name == "lds.tmpl.yaml" || name == "cds.yaml"
			},
			requireBootstrap: expectValidBootstrap,
		},
		{
			name:         "envoy.tmpl.yaml: not a valid YAML",
			workspaceDir: "testdata/workspace3",
			isEnvoyTemplate: func(name string) bool {
				return false
			},
			requireBootstrap: expectInvalidBootstrap,
		},
		{
			name:         "envoy.tmpl.yaml: invalid paths to `lds` and `cds` files",
			workspaceDir: "testdata/workspace4",
			isEnvoyTemplate: func(name string) bool {
				return name == "envoy.tmpl.yaml"
			},
			requireBootstrap: expectValidBootstrap,
		},
		{
			name:         "envoy.tmpl.yaml: .txt configuration",
			workspaceDir: "testdata/workspace8",
			isEnvoyTemplate: func(name string) bool {
				return name == "envoy.tmpl.yaml"
			},
			requireBootstrap: expectValidBootstrap,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			workspace, err := workspaces.GetWorkspaceAt(test.workspaceDir)
			require.NoError(t, err)

			example, err := workspace.GetExample("default")
			require.NoError(t, err)

			ctx := runContext(workspace, example)

			configDir, err := NewConfigDir(ctx)
			require.NoError(t, err)

			defer func() {
				err := configDir.Close()
				require.NoError(t, err)

				// verifying config dir has been removed
				_, err = os.Stat(configDir.GetDir())
				require.Error(t, err)
				require.True(t, os.IsNotExist(err))
			}()

			// verify the config dir
			require.FileExists(t, configDir.GetBootstrapFile())
			test.requireBootstrap(t, configDir.GetBootstrap())

			// verify contents of the config dir
			for _, fileName := range ctx.Opts.Example.GetFiles().GetNames() {
				expected, err := ioutil.ReadFile(filepath.Join(test.workspaceDir, "expected/getenvoy_extension_run", fileName))
				require.NoError(t, err)
				actual, err := ioutil.ReadFile(filepath.Join(configDir.GetDir(), fileName))
				require.NoError(t, err)
				if test.isEnvoyTemplate(fileName) {
					require.YAMLEq(t, string(expected), string(actual))
				} else {
					require.Equal(t, string(expected), string(actual))
				}
			}
		})
	}
}

func runContext(workspace model.Workspace, example model.Example) *runtime.RunContext {
	_, f := example.GetExtensionConfig()
	return &runtime.RunContext{
		Opts: runtime.RunOpts{
			Workspace: workspace,
			Example: runtime.ExampleOpts{
				Name:    "default",
				Example: example,
			},
			Extension: runtime.ExtensionOpts{
				WasmFile: `/path/to/extension.wasm`,
				Config:   *f,
			},
		},
	}
}

func expectValidBootstrap(t *testing.T, bootstrap *envoybootstrap.Bootstrap) {
	require.NotNil(t, bootstrap)
	require.Equal(t, "/dev/null", bootstrap.GetAdmin().GetAccessLogPath())
	require.Equal(t, "127.0.0.1", bootstrap.GetAdmin().GetAddress().GetSocketAddress().GetAddress())
	require.Equal(t, uint32(9901), bootstrap.GetAdmin().GetAddress().GetSocketAddress().GetPortValue())
}

func expectInvalidBootstrap(t *testing.T, bootstrap *envoybootstrap.Bootstrap) {
	require.Nil(t, bootstrap)
}

func TestNewConfigDirValidates(t *testing.T) {
	abs := func(path string) string {
		path, err := filepath.Abs(path)
		if err != nil {
			panic(err)
		}
		return path
	}
	tests := []struct {
		name         string
		workspaceDir string
		expectedErr  string
	}{
		{
			name:         "envoy.tmpl.yaml: invalid placeholder",
			workspaceDir: "testdata/workspace5",
			expectedErr:  fmt.Sprintf(`failed to process Envoy config template coming from %q: failed to render Envoy config template: template: :4:19: executing "" at <.GetEnvoy.DefaultValue>: error calling DefaultValue: unknown property "???"`, abs("testdata/workspace5/.getenvoy/extension/examples/default/envoy.tmpl.yaml")),
		},
		{
			name:         "envoy.tmpl.yaml + lds.tmpl.yaml + cds.yaml: invalid placeholder in lds.tmpl.yaml",
			workspaceDir: "testdata/workspace6",
			expectedErr:  fmt.Sprintf(`failed to process Envoy config template coming from %q: failed to render Envoy config template: template: :22:34: executing "" at <.GetEnvoy.Extension.Code>: error calling Code: unable to resolve Wasm module [???]: not supported yet`, abs("testdata/workspace6/.getenvoy/extension/examples/default/lds.tmpl.yaml")),
		},
		{
			name:         "envoy.tmpl.yaml + lds.tmpl.yaml + cds.tmpl.yaml: invalid placeholder in cds.tmpl.yaml",
			workspaceDir: "testdata/workspace7",
			expectedErr:  fmt.Sprintf(`failed to process Envoy config template coming from %q: failed to render Envoy config template: template: :1:18: executing "" at <.GetEnvoy.Extension.Config>: error calling Config: unable to resolve a named config [???]: not supported yet`, abs("testdata/workspace7/.getenvoy/extension/examples/default/cds.tmpl.yaml")),
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			workspace, err := workspaces.GetWorkspaceAt(test.workspaceDir)
			require.NoError(t, err)

			example, err := workspace.GetExample("default")
			require.NoError(t, err)

			ctx := runContext(workspace, example)

			configDir, err := NewConfigDir(ctx)
			require.Nil(t, configDir)
			require.EqualError(t, err, test.expectedErr)
		})
	}
}
