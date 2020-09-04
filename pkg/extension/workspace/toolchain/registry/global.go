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

var (
	// global represents a global registry of supported toolchains.
	global = make(registry)
)

// Register registers given toolchain in a global registry.
func Register(entry Entry) {
	global.Register(entry)
}

// Get returns a toolchain registered in a global registry for a given kind.
func Get(kind string) (Entry, bool) {
	return global.Get(kind)
}
