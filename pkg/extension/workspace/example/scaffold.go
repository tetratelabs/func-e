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

package example

import (
	"github.com/pkg/errors"

	"github.com/tetratelabs/getenvoy/pkg/extension/example/init/registry"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"

	scaffoldutil "github.com/tetratelabs/getenvoy/pkg/util/scaffold"
)

// ScaffoldOpts represents configuration options supported by Scaffold().
type ScaffoldOpts struct {
	Workspace    model.Workspace
	Name         string
	ProgressSink scaffoldutil.ProgressSink
}

// ScaffoldIfDefault generates the default example setup.
func ScaffoldIfDefault(opts *ScaffoldOpts) error {
	if opts.Name != Default {
		return nil
	}
	exists, err := opts.Workspace.HasExample(Default)
	if err != nil {
		return errors.Wrapf(err, "failed to determine whether example %q already exists", Default)
	}
	if exists {
		return nil
	}
	return Scaffold(opts)
}

// Scaffold generates a new example setup.
func Scaffold(opts *ScaffoldOpts) error {
	descriptor := opts.Workspace.GetExtensionDescriptor()
	factory, err := registry.Get(descriptor, registry.DefaultExample)
	if err != nil {
		// must be caught by unit tests
		panic(errors.Errorf("there is no %q example for extension category %q", registry.DefaultExample, descriptor.Category))
	}
	example, err := factory.NewExample(descriptor)
	if err != nil {
		return errors.Wrapf(err, "failed to generate %q example", opts.Name)
	}
	return opts.Workspace.SaveExample(opts.Name, example, model.ProgressSink{ProgressSink: opts.ProgressSink})
}
