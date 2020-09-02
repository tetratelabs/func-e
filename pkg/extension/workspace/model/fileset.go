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

package model

import (
	"sort"
)

// NewFileSet returns a new mutable file set.
func NewFileSet() FileSet {
	return make(fileSet)
}

// fileSet represents a mutable set of configuration files.
type fileSet map[string]*File

// Add adds a file to the set.
func (s fileSet) Add(name string, file *File) {
	s[name] = file
}

// Has returns true if the set includes a file with a given name.
func (s fileSet) Has(name string) bool {
	_, exists := s[name]
	return exists
}

// Get returns a file by name.
func (s fileSet) Get(name string) *File {
	return s[name]
}

// GetNames returns an ordered list of file names in the set.
func (s fileSet) GetNames() []string {
	names := make([]string, 0, len(s))
	for name := range s {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
