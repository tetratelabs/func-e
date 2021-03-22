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
	"sync"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
)

var (
	// global represents a global registry of example templates.
	global     registry
	globalOnce sync.Once
)

func globalRegistry() registry {
	// do lazy init to avoid zip decompression unless absolutely necessary
	globalOnce.Do(func() {
		global = newDefaultRegistry()
	})
	return global
}

// Get returns an example template registered in a global registry for a given
// extension category and name.
func Get(descriptor *extension.Descriptor, name string) (*Entry, error) {
	return globalRegistry().Get(descriptor, name)
}
