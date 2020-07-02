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

package common

import (
	"path/filepath"

	"github.com/tetratelabs/log"

	"github.com/mitchellh/go-homedir"
)

var (
	// HomeDir holds the location of GetEnvoy home directory - place for
	// downloaded artifacts, caches, etc.
	HomeDir = DefaultHomeDir()
)

// DefaultHomeDir returns the default GetEnvoy home directory.
func DefaultHomeDir() string {
	home, err := homedir.Dir()
	dir := filepath.Join(home, ".getenvoy")
	if err != nil {
		log.Errorf("unable to determine the user home directory: %v", err)
		log.Warnf("default GetEnvoy home directory will have a non-standard value %q", dir)
	}
	return dir
}
