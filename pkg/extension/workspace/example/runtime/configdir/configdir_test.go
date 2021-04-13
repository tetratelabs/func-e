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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/example/runtime"
	. "github.com/tetratelabs/getenvoy/pkg/extension/workspace/example/runtime/configdir"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
)

func TestNewConfigDir(t *testing.T) {
	tests := []struct {
		name                 string
		workspaceDir         string
		expectValidBootstrap bool
	}{
		{
			name:                 "envoy.tmpl.yaml",
			workspaceDir:         "testdata/workspace1",
			expectValidBootstrap: true,
		},
		{
			name:                 "envoy.tmpl.yaml + lds.tmpl.yaml + cds.yaml",
			workspaceDir:         "testdata/workspace2",
			expectValidBootstrap: true,
		},
		{
			name:                 "envoy.tmpl.yaml: not a valid YAML",
			workspaceDir:         "testdata/workspace3",
			expectValidBootstrap: false,
		},
		{
			name:                 "envoy.tmpl.yaml: invalid paths to `lds` and `cds` files",
			workspaceDir:         "testdata/workspace4",
			expectValidBootstrap: true,
		},
		{
			name:                 "envoy.tmpl.yaml: .txt configuration",
			workspaceDir:         "testdata/workspace8",
			expectValidBootstrap: true,
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

			// verify the bootstrap file
			require.FileExists(t, configDir.GetBootstrapFile())
			if test.expectValidBootstrap {
				bootstrap := configDir.GetBootstrap()
				require.NotNil(t, bootstrap)
				require.Equal(t, "/dev/null", bootstrap.GetAdmin().GetAccessLogPath())
				require.Equal(t, "127.0.0.1", bootstrap.GetAdmin().GetAddress().GetSocketAddress().GetAddress())
				require.Equal(t, uint32(9901), bootstrap.GetAdmin().GetAddress().GetSocketAddress().GetPortValue())
			} else {
				require.Nil(t, configDir.GetBootstrap())
			}

			// verify contents of the config dir
			for _, fileName := range ctx.Opts.Example.GetFiles().GetNames() {
				expected, err := ioutil.ReadFile(filepath.Join(test.workspaceDir, "expected/getenvoy_extension_run", fileName))
				require.NoError(t, err)
				actual, err := ioutil.ReadFile(filepath.Join(configDir.GetDir(), fileName))
				require.NoError(t, err)

				switch {
				case strings.HasSuffix(fileName, ".yaml") && fileName != "envoy.tmpl.yaml":
					require.YAMLEq(t, string(expected), string(actual), `%s is not valid yaml`, fileName)
				case strings.HasSuffix(fileName, ".json"):
					require.JSONEq(t, string(expected), string(actual), `%s is not valid json`, fileName)
				case fileName != "envoy.tmpl.yaml": // we don't need to check our input template
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
