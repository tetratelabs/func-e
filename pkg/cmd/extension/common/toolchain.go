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

package common

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/docker/distribution/reference"

	builtinconfig "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/toolchain/builtin"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
	toolchains "github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain/types"
	argutil "github.com/tetratelabs/getenvoy/pkg/util/args"
)

// NewToolchainOpts returns new ToolchainOpts.
func NewToolchainOpts() ToolchainOpts {
	return ToolchainOpts{
		Name: toolchains.Default,
	}
}

// ToolchainOpts represents a toolchain-related part of command options.
type ToolchainOpts struct {
	Name    string
	Builtin BuiltinToolchainOpts
}

// BuiltinToolchainOpts represents command options specific to built-in toolchain.
type BuiltinToolchainOpts struct {
	// Builder image.
	ContainerImage string
	// Docker cli options.
	ContainerOptions []string
}

// Validate returns an error if ToolchainOpts is not valid.
func (o *ToolchainOpts) Validate() error {
	if o.Name == "" {
		return errors.Errorf("toolchain name cannot be empty")
	}
	return o.Builtin.Validate()
}

// Validate returns an error if BuiltinToolchainOpts is not valid.
func (o *BuiltinToolchainOpts) Validate() error {
	if o.ContainerImage != "" {
		if _, err := reference.Parse(o.ContainerImage); err != nil {
			return errors.Wrapf(err, "%q is not a valid image name", o.ContainerImage)
		}
	}
	if len(o.ContainerOptions) > 0 {
		options, err := argutil.SplitCommandLine(o.ContainerOptions...)
		if err != nil {
			return err
		}
		o.ContainerOptions = options
	}
	return nil
}

// ApplyTo applies container-specific command options to a given built-in toolchain config.
func (o *BuiltinToolchainOpts) ApplyTo(config *builtinconfig.ContainerConfig) {
	if o.ContainerImage != "" {
		config.Image = o.ContainerImage
	}
	if len(o.ContainerOptions) > 0 {
		config.Options = append(config.Options, o.ContainerOptions...)
	}
}

// AddToolchainFlags adds toolchain-related options to a given command.
func AddToolchainFlags(cmd *cobra.Command, opts *ToolchainOpts) {
	cmd.PersistentFlags().StringVar(&opts.Name, "toolchain", opts.Name,
		`Name of the toolchain to use, e.g. "default" toolchain that is backed by a Docker build container`)
	cmd.PersistentFlags().StringVar(&opts.Builtin.ContainerImage, "toolchain-container-image", "",
		`Run build container using given image`)
	cmd.PersistentFlags().StringArrayVar(&opts.Builtin.ContainerOptions, "toolchain-container-options", nil,
		`Run build container using extra Docker cli options`)
}

// ToolchainCustomizer knows how to customize toolchain config.
type ToolchainCustomizer interface {
	GetToolchainName() string
	ApplyTo(config interface{})
}

// LoadToolchain loads a toolchain by its name and customizes its configuration according to
// command-line flags.
func LoadToolchain(workspace model.Workspace, opts ToolchainCustomizer) (types.Toolchain, error) {
	builder, err := toolchains.LoadToolchain(opts.GetToolchainName(), workspace)
	if err != nil {
		return nil, err
	}
	opts.ApplyTo(builder.GetConfig())
	toolchain, err := builder.Build()
	if err != nil {
		return nil, err
	}
	return toolchain, nil
}
