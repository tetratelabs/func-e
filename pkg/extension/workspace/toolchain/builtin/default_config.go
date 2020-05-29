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
	"github.com/pkg/errors"

	extensionconfig "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
	builtinconfig "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/toolchain/builtin"
	dockerutil "github.com/tetratelabs/getenvoy/pkg/util/docker"
	"github.com/tetratelabs/getenvoy/pkg/version"
)

var (
	// defaultBuildImageOrg represents organization name of the default builder images in a Docker registry.
	defaultBuildImageOrg = "tetratelabs"
	// defaultBuildImageTag represents a tag of the default builder images in a Docker registry.
	defaultBuildImageTag = version.Build.Version
	// defaultRustBuildImage represents a full name of the default Rust builder image in a Docker registry.
	defaultRustBuildImage = dockerutil.ImageName{Org: defaultBuildImageOrg, Name: "getenvoy-extension-rust-builder", Tag: defaultBuildImageTag}.String()
)

// defaultConfigFor returns a default toolchain config for a given extension.
func defaultConfigFor(extension *extensionconfig.Descriptor) *builtinconfig.ToolchainConfig {
	cfg := builtinconfig.NewToolchainConfig()
	cfg.Container = &builtinconfig.ContainerConfig{
		Image: defaultBuildImageFor(extension.Language),
	}
	return cfg
}

func defaultBuildImageFor(language extensionconfig.Language) string {
	switch language {
	case extensionconfig.LanguageRust:
		return defaultRustBuildImage
	default:
		// must be caught by unit tests
		panic(errors.Errorf("failed to determine default build image for unsupported programming language %q", language))
	}
}
