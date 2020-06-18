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

package toolchain

import (
	"github.com/pkg/errors"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain/builtin"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain/registry"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain/types"
)

const (
	// Default represents a name of the toolchain available out-of-the-box.
	Default = "default"
)

// LoadToolchain loads toolchain configuration by its name.
func LoadToolchain(name string, workspace model.Workspace) (types.ToolchainBuilder, error) {
	switch name {
	case Default:
		if err := ensureDefaultToolchainExists(workspace); err != nil {
			return nil, errors.Wrap(err, "failed to ensure the default toolchain exists")
		}
	default:
		return nil, errors.Errorf("unknown toolchain %q. At the moment, only %q toolchain is supported", name, Default)
	}
	return loadToolchain(name, workspace)
}

func ensureDefaultToolchainExists(workspace model.Workspace) error {
	exists, err := workspace.HasToolchain(Default)
	if err != nil {
		return errors.Wrapf(err, "failed to determine whether toolchain %q already exists", Default)
	}
	if exists {
		return nil
	}
	extension := workspace.GetExtensionDescriptor()
	cfg := builtin.ExampleConfig(extension)
	return workspace.SaveToolchainConfig(Default, cfg)
}

func loadToolchain(name string, workspace model.Workspace) (types.ToolchainBuilder, error) {
	exists, err := workspace.HasToolchain(name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to determine whether toolchain %q exists", name)
	}
	if !exists {
		return nil, errors.Errorf("unknown toolchain %q", name)
	}
	file, err := workspace.GetToolchainConfig(name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load configuration for toolchain %q", name)
	}
	configErr := func(err error) error {
		return errors.Wrapf(err, "toolchain %q has invalid configuration coming from %q", name, file.Source)
	}
	meta := new(config.Meta)
	err = config.Unmarshal(file.Content, meta)
	if err != nil {
		return nil, configErr(err)
	}
	factory, exists := registry.Get(meta.Kind)
	if !exists {
		return nil, configErr(errors.Errorf("unknown toolchain kind %q", meta.Kind))
	}
	builder, err := factory.LoadConfig(registry.LoadConfigArgs{
		Workspace: workspace,
		Toolchain: registry.ToolchainConfig{
			Name:   name,
			Config: file,
		},
	})
	if err != nil {
		return nil, configErr(err)
	}
	if err := builder.GetConfig().Validate(); err != nil {
		return nil, configErr(err)
	}
	return &validatingBuilder{name, builder}, nil
}

type validatingBuilder struct {
	name string
	types.ToolchainBuilder
}

func (b *validatingBuilder) Build() (types.Toolchain, error) {
	if err := b.GetConfig().Validate(); err != nil {
		return nil, errors.Wrapf(err, "toolchain %q has invalid configuration", b.name)
	}
	return b.ToolchainBuilder.Build()
}
