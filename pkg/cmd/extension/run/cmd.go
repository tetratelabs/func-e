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

package run

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	reference "github.com/tetratelabs/getenvoy/pkg"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/getenvoy/pkg/cmd/extension/build"
	"github.com/tetratelabs/getenvoy/pkg/cmd/extension/common"
	examplecmd "github.com/tetratelabs/getenvoy/pkg/cmd/extension/example"
	"github.com/tetratelabs/getenvoy/pkg/cmd/run"
	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	builtinconfig "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/toolchain/builtin"
	examples "github.com/tetratelabs/getenvoy/pkg/extension/workspace/example"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/example/runtime"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/example/runtime/configdir"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
	"github.com/tetratelabs/getenvoy/pkg/globals"
	commontypes "github.com/tetratelabs/getenvoy/pkg/types"
	argutil "github.com/tetratelabs/getenvoy/pkg/util/args"
	cmdutil "github.com/tetratelabs/getenvoy/pkg/util/cmd"
	osutil "github.com/tetratelabs/getenvoy/pkg/util/os"
	scaffoldutil "github.com/tetratelabs/getenvoy/pkg/util/scaffold"
	uiutil "github.com/tetratelabs/getenvoy/pkg/util/ui"
)

// cmdOpts represents configuration options of the `run` command.
type cmdOpts struct {
	// Toolchain to use to build the *.wasm file.
	Toolchain common.ToolchainOpts
	// Run options.
	Run runOpts
}

// runOpts associates validation logic with runtime.RunOpts.
type runOpts runtime.RunOpts

func (opts *runOpts) Validate() error {
	if err := opts.validateExample(); err != nil {
		return err
	}
	if err := opts.validateExtension(); err != nil {
		return err
	}
	if err := opts.validateEnvoy(); err != nil {
		return err
	}
	return nil
}

func (opts *runOpts) validateExample() error {
	return model.ValidateExampleName(opts.Example.Name)
}

func (opts *runOpts) validateExtension() error {
	// pre-built *.wasm file
	if opts.Extension.WasmFile != "" {
		if err := osutil.IsRegularFile(opts.Extension.WasmFile); err != nil {
			return fmt.Errorf("unable to find a pre-built *.wasm file at %q: %w", opts.Extension.WasmFile, err)
		}
	}
	// custom extension config
	if opts.Extension.Config.Source != "" {
		data, err := os.ReadFile(opts.Extension.Config.Source)
		if err != nil {
			return fmt.Errorf("failed to read custom extension config from file %q: %w", opts.Extension.Config.Source, err)
		}
		opts.Extension.Config.Content = data
	}
	return nil
}

func (opts *runOpts) validateEnvoy() error {
	// Envoy version
	if opts.Envoy.Version != "" {
		if _, err := commontypes.ParseReference(opts.Envoy.Version); err != nil {
			return fmt.Errorf("envoy version is not valid: %w", err)
		}
	}
	// Envoy args
	if len(opts.Envoy.Args) > 0 {
		args, err := argutil.SplitCommandLine(opts.Envoy.Args...)
		if err != nil {
			return err
		}
		opts.Envoy.Args = args
	}
	return nil
}

func (opts *cmdOpts) GetToolchainName() string {
	return opts.Toolchain.Name
}

// ApplyTo applies toolchain-related command options to a given toolchain config.
func (opts *cmdOpts) ApplyTo(config interface{}) {
	if c, ok := config.(*builtinconfig.ToolchainConfig); ok {
		opts.Toolchain.Builtin.ApplyTo(c.GetBuildContainer())
	}
}

// NewCmd returns a command that runs the extension.
func NewCmd(o *globals.GlobalOpts) *cobra.Command {
	opts := &cmdOpts{
		Toolchain: common.NewToolchainOpts(o),
		Run:       runOpts{Example: runtime.ExampleOpts{Name: examples.Default}},
	}

	//nolint:lll
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run Envoy extension in the example setup.",
		Long: `
Run Envoy extension in the example setup.`,
		Example: fmt.Sprintf(`
  # Run Envoy extension in the "default" example setup
  getenvoy extension run

  # Run Envoy extension in the "default" example setup using a particular Envoy release provided by getenvoy.io
  getenvoy extension run --envoy-version %s

  # Run Envoy extension in the "default" example setup using Envoy with extra options
  getenvoy extension run --envoy-options '--concurrency 2 --component-log-level wasm:debug,config:trace'

  # Run Envoy extension in the "default" example setup using a pre-built *.wasm file
  getenvoy extension run --extension-file /path/to/extension.wasm

  # Run Envoy extension in the "default" example setup using a custom extension config
  getenvoy extension run --extension-config-file /path/to/config.json

  # Run Envoy extension in the "default" example setup; build the extension using Docker build container with extra options
  getenvoy extension run --toolchain-container-options '-e VAR=VALUE -v /host/path:/container/path'

  # Run Envoy extension in the "default" example setup; build the extension using Docker build container with SSH agent forwarding enabled (Docker for Mac)
  getenvoy extension run --toolchain-container-options `+
			`'--mount type=bind,src=/run/host-services/ssh-auth.sock,target=/run/host-services/ssh-auth.sock `+
			`-e SSH_AUTH_SOCK=/run/host-services/ssh-auth.sock'`, reference.Latest),
		Args: func(*cobra.Command, []string) error {
			if err := opts.Run.Validate(); err != nil {
				return err
			}
			return opts.Toolchain.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := workspaces.GetWorkspaceAt(o.ExtensionDir)
			if err != nil {
				return err
			}
			opts.Run.Workspace = workspace

			progressSink := examplecmd.NewAddExampleFeedback(uiutil.NewStyleFuncs(o.NoColors), cmd.ErrOrStderr())
			example, err := ensureDefaultExample(workspace, opts.Run.Example.Name, progressSink)
			if err != nil {
				return err
			}
			opts.Run.Example.Example = example

			if e := validateOrBuildWasm(cmd, opts, workspace); e != nil {
				return e
			}

			// We now have a valid workspace, prepare to generate the Envoy config file.
			runtimeOpts := runtime.RunOpts(opts.Run)
			if e := run.InitializeRunOpts(o, runtimeOpts.GetEnvoyReference()); e != nil {
				return e
			}
			if _, e := configdir.NewConfigDir(&runtimeOpts, o.WorkingDir); e != nil {
				return e
			}

			// Run Envoy, fetching the distribution as needed.
			if o.EnvoyPath == "" { // not overridden for tests
				envoyPath, e := envoy.FetchIfNeeded(o, runtimeOpts.GetEnvoyReference())
				if e != nil {
					return e
				}
				o.EnvoyPath = envoyPath
			}

			err = run.Run(o, cmd, append([]string{"-c", "envoy.yaml"}, runtimeOpts.Envoy.Args...))
			if err != nil {
				return fmt.Errorf("failed to run %q example: %w", opts.Run.Example.Name, err)
			}
			return nil
		},
	}
	common.AddToolchainFlags(cmd, &opts.Toolchain)
	cmd.PersistentFlags().StringVar(&opts.Run.Example.Name, "example", opts.Run.Example.Name,
		`Name of the example to run`)
	cmd.PersistentFlags().StringVar(&opts.Run.Extension.WasmFile, "extension-file", opts.Run.Extension.WasmFile,
		`Use a pre-built *.wasm file`)
	cmd.PersistentFlags().StringVar(&opts.Run.Extension.Config.Source, "extension-config-file", opts.Run.Extension.Config.Source,
		`Use a custom extension config`)
	cmd.PersistentFlags().StringVar(&opts.Run.Envoy.Version, "envoy-version", opts.Run.Envoy.Version,
		`Use a particular Envoy release provided by getenvoy.io. For a list of available releases run "getenvoy list"`)
	cmd.PersistentFlags().StringArrayVar(&opts.Run.Envoy.Args, "envoy-options", nil,
		`Run Envoy using extra cli options`)
	return cmd
}

// auto-create default example setup if necessary
func ensureDefaultExample(workspace model.Workspace, name string, progressSink scaffoldutil.ProgressSink) (model.Example, error) {
	scaffoldOpts := &examples.ScaffoldOpts{
		Workspace:    workspace,
		Name:         name,
		ProgressSink: progressSink,
	}
	if e := examples.ScaffoldIfDefault(scaffoldOpts); e != nil {
		return nil, e
	}
	// find example
	return examples.LoadExample(scaffoldOpts.Name, workspace)
}

func validateOrBuildWasm(cmd *cobra.Command, opts *cmdOpts, workspace model.Workspace) error {
	// build *.wasm file unless a user provided a pre-built one
	if opts.Run.Extension.WasmFile == "" {
		toolchain, e := common.LoadToolchain(workspace, opts)
		if e != nil {
			return e
		}
		e = build.Build(toolchain, cmdutil.StreamsOf(cmd))
		if e != nil {
			return e
		}
		opts.Run.Extension.WasmFile = toolchain.GetBuildOutputWasmFile()
	}

	// Use absolute path because Envoy doesn't run in this directory and may be configured to hot-reload the VM.
	absoluteWasmFile, err := filepath.Abs(opts.Run.Extension.WasmFile)
	if err != nil {
		return err
	}
	opts.Run.Extension.WasmFile = absoluteWasmFile
	return nil
}
