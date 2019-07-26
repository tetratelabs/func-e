// Copyright 2019 Tetrate
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

package getenvoy

import (
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/tetratelabs/getenvoy/pkg/binary"
)

// New creates a new GetEnvoy binary.Runtime with the local file storage set to the home directory
func New() (binary.Runtime, error) {
	usrDir, err := homedir.Dir()
	return &Runtime{
		local: filepath.Join(usrDir, ".getenvoy", "builds"),
	}, err
}

// Runtime implements the GetEnvoy binary.Runtime
type Runtime struct {
	local string
}
