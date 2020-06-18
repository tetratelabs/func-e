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
)

const (
	// Default represents a name of the example available out-of-the-box.
	Default = "default"
)

// LoadExample loads an example by its name.
func LoadExample(name string, workspace model.Workspace) (model.Example, error) {
	switch name {
	case Default:
		if err := ensureDefaultExampleExists(workspace); err != nil {
			return nil, errors.Wrapf(err, "failed to ensure the %q example always exists", Default)
		}
	default:
		return nil, errors.Errorf("unknown example %q. At the moment, only %q example is supported", name, Default)
	}
	return loadExample(name, workspace)
}

func ensureDefaultExampleExists(workspace model.Workspace) error {
	exists, err := workspace.HasExample(Default)
	if err != nil {
		return errors.Wrapf(err, "failed to determine whether example %q already exists", Default)
	}
	if exists {
		return nil
	}
	descriptor := workspace.GetExtensionDescriptor()
	factory, err := registry.Get(descriptor.Category, registry.DefaultExample)
	if err != nil {
		// must be caught by unit tests
		panic(errors.Errorf("there is no %q example for extension category %q", registry.DefaultExample, descriptor.Category))
	}
	example, err := factory.NewExample(descriptor)
	if err != nil {
		return errors.Wrapf(err, "failed to generate %q example", Default)
	}
	return workspace.SaveExample(Default, example)
}

func loadExample(name string, workspace model.Workspace) (model.Example, error) {
	exists, err := workspace.HasExample(name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to determine whether example %q exists", name)
	}
	if !exists {
		return nil, errors.Errorf("there is no example %q", name)
	}
	example, err := workspace.GetExample(name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load example %q", name)
	}
	return example, nil
}
