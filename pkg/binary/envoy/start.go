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

package envoy

import (
	"github.com/tetratelabs/log"

	"github.com/tetratelabs/getenvoy/pkg/binary"
)

func (r *Runtime) handlePreStart() {
	// Execute all registered preStart functions
	for _, f := range r.preStart {
		if err := f(r); err != nil {
			log.Error(err.Error())
		}
	}
}

// RegisterPreStart registers the passed functions to be run before Envoy has started
func (r *Runtime) RegisterPreStart(f ...func(binary.Runner) error) {
	r.preStart = append(r.preStart, f...)
}
