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

package types

import (
	ioutil "github.com/tetratelabs/getenvoy/pkg/util/io"
)

// BuildContext represents a context of the `build` tool.
type BuildContext struct {
	IO ioutil.StdStreams
}

// BuildTool knows how to build extension created by getenvoy toolkit.
type BuildTool interface {
	Build(BuildContext) error
}

// TestContext represents a context of the `test` tool.
type TestContext struct {
	IO ioutil.StdStreams
}

// TestTool knows how to test extension created by getenvoy toolkit.
type TestTool interface {
	Test(TestContext) error
}

// CleanContext represents a context of the `clean` tool.
type CleanContext struct {
	IO ioutil.StdStreams
}

// CleanTool knows how to clean build directory of an extension created by getenvoy toolkit.
type CleanTool interface {
	Clean(CleanContext) error
}

// Toolchain represents a set of external tools used by getenvoy toolkit.
type Toolchain interface {
	BuildTool
	TestTool
	CleanTool
}

// ToolchainConfig represents a generic toolchain config.
type ToolchainConfig interface {
	Validate() error
}

// ToolchainBuilder represents a Toolchain builder.
type ToolchainBuilder interface {
	// GetConfig returns toolchain config that can be modified prior to building
	// the final toolchain.
	GetConfig() ToolchainConfig
	Build() (Toolchain, error)
}
