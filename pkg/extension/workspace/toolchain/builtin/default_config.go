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
	"fmt"

	extensionconfig "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
	builtinconfig "github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/toolchain/builtin"
	"github.com/tetratelabs/getenvoy/pkg/version"
)

var (
	// defaultBuildImageOrg represents organization name of the default builder images in a Docker registry.
	defaultBuildImageOrg = "getenvoy"
	// defaultBuildImageTag represents a tag of the default builder images in a Docker registry.
	defaultBuildImageTag = func() string {
		if version.IsDevBuild() {
			// in case of development builds, fallback to the 'latest' version of a builder image
			return "latest"
		}
		// by default, use a builder image of the same version as getenvoy binary
		return version.Build.Version
	}()
	// defaultRustBuildImage represents a full name of the default Rust builder image in a Docker registry.
	defaultRustBuildImage   = fmt.Sprintf("%s/%s:%s", defaultBuildImageOrg, "extension-rust-builder", defaultBuildImageTag)
	defaultTinyGoBuildImage = fmt.Sprintf("%s/%s:%s", defaultBuildImageOrg, "extension-tinygo-builder", defaultBuildImageTag)
)

// defaultConfigFor returns a default toolchain config for a given extension.
func defaultConfigFor(extension *extensionconfig.Descriptor) *builtinconfig.ToolchainConfig {
	cfg := builtinconfig.NewToolchainConfig()
	cfg.Container = &builtinconfig.ContainerConfig{
		Image: defaultBuildImageFor(extension.Language),
	}
	cfg.Build = &builtinconfig.BuildConfig{
		Output: &builtinconfig.BuildOutput{
			WasmFile: defaultOutputPathFor(extension.Language),
		},
	}
	return cfg
}

func defaultBuildImageFor(language extensionconfig.Language) string {
	switch language {
	case extensionconfig.LanguageRust:
		return defaultRustBuildImage
	case extensionconfig.LanguageTinyGo:
		return defaultTinyGoBuildImage
	default:
		// must be caught by unit tests
		panic(fmt.Errorf("failed to determine default build image for unsupported programming language %q", language))
	}
}

func defaultOutputPathFor(language extensionconfig.Language) string {
	// choose location inside a build directory of that language
	switch language {
	case extensionconfig.LanguageRust:
		// keep *.wasm file inside Cargo build dir (to be cleaned up automatically)
		return "target/getenvoy/extension.wasm"
	case extensionconfig.LanguageTinyGo:
		return "build/extension.wasm"
	default:
		// must be caught by unit tests
		panic(fmt.Errorf("failed to determine default output path for unsupported programming language %q", language))
	}
}
