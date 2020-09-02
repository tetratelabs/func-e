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

package runtime

import (
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
	ioutil "github.com/tetratelabs/getenvoy/pkg/util/io"
)

// RunOpts represents arguments to the `run` tool.
type RunOpts struct {
	Workspace model.Workspace
	Example   ExampleOpts
	Extension ExtensionOpts
	Envoy     EnvoyOpts
}

// ExampleOpts represents example-related part of arguments to the `run` tool.
type ExampleOpts struct {
	Name string
	model.Example
}

// ExtensionOpts represents extension-related part of arguments to the `run` tool.
type ExtensionOpts struct {
	WasmFile string
	Config   model.File
}

// EnvoyOpts represents Envoy-related part of arguments to the `run` tool.
type EnvoyOpts struct {
	Version string
	Path    string
	Args    []string
}

// RunContext represents a context of the `run` tool.
type RunContext struct {
	Opts RunOpts
	IO   ioutil.StdStreams
}

// Runtime knows how to run an example.
type Runtime interface {
	Run(*RunContext) error
}
