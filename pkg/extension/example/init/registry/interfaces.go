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

package registry

import (
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/model"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
)

const (
	// DefaultExample represents a name of the example that is always available.
	DefaultExample = "default"
)

// Entry represents a registry entry.
type Entry struct {
	Category   extension.Category
	Name       string
	NewExample NewExampleFunc
}

// NewExampleFunc represents a function responsible for generating an example
// for a given extension.
type NewExampleFunc func(descriptor *extension.Descriptor) (model.Example, error)
