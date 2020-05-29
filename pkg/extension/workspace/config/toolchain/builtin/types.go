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
)

const (
	// Kind identifies configuration of the built-in toolchain.
	Kind = "BuiltinToolchain"
)

// ToolchainConfig represents configuration of the built-in toolchain.
type ToolchainConfig struct {
	config.Meta `json:",inline"`
	// Configuration of the default Docker build container.
	Container *ContainerConfig `json:"container,omitempty"`
	// Configuration of `build` tool.
	Build *BuildConfig `json:"build,omitempty"`
	// Configuration of `test` tool.
	Test *TestConfig `json:"test,omitempty"`
	// Configuration of `clean` tool.
	Clean *CleanConfig `json:"clean,omitempty"`
}

// BuildConfig represents configuration of `build` tool.
type BuildConfig struct {
	// Configuration of a Docker build container.
	Container *ContainerConfig `json:"container,omitempty"`
}

// TestConfig represents configuration of `test` tool.
type TestConfig struct {
	// Configuration of a Docker build container.
	Container *ContainerConfig `json:"container,omitempty"`
}

// CleanConfig represents configuration of `clean` tool.
type CleanConfig struct {
	// Configuration of a Docker build container.
	Container *ContainerConfig `json:"container,omitempty"`
}

// ContainerConfig represents configuration of a Docker build container.
type ContainerConfig struct {
	// Builder image.
	Image string `json:"image,omitempty"`
	// Docker cli options.
	Options []string `json:"options,omitempty"`
}
