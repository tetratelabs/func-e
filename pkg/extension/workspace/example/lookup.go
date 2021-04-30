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
	"fmt"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"
)

const (
	// Default represents a name of the example available out-of-the-box.
	Default = "default"
)

// LoadExample loads an example by its name.
func LoadExample(name string, workspace model.Workspace) (model.Example, error) {
	exists, err := workspace.HasExample(name)
	if err != nil {
		return nil, fmt.Errorf("failed to determine whether example %q exists: %w", name, err)
	}
	if !exists {
		return nil, fmt.Errorf("there is no example %q", name)
	}
	example, err := workspace.GetExample(name)
	if err != nil {
		return nil, fmt.Errorf("failed to load example %q: %w", name, err)
	}
	return example, nil
}
