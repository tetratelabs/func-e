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

package builtin

import (
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config"
	builtinconfig "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/toolchain/builtin"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain/registry"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain/types"
)

func init() {
	registry.Register(registry.Entry{
		Kind: builtinconfig.Kind,
		LoadConfig: func(args registry.LoadConfigArgs) (types.ToolchainBuilder, error) {
			cfg := builtinconfig.NewToolchainConfig()
			if err := config.Unmarshal(args.Toolchain.Config.Content, cfg); err != nil {
				return nil, err
			}
			extension := args.Workspace.GetExtensionDescriptor()
			defaultCfg := defaultConfigFor(extension)
			cfg.DefaultTo(defaultCfg)
			return &builder{args.Toolchain.Name, cfg, args.Workspace}, nil
		},
	})
}

type builder struct {
	name      string
	cfg       *builtinconfig.ToolchainConfig
	workspace model.Workspace
}

func (b *builder) GetConfig() types.ToolchainConfig {
	return b.cfg
}

func (b *builder) Build() (types.Toolchain, error) {
	return NewToolchain(b.name, b.cfg, b.workspace), nil
}
