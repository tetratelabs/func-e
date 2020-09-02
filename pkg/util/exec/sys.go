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

package exec

import "syscall"

// parentDeathAttrFn represents an OS-specific field on syscall.SysProcAttr.
type parentDeathAttrFn func(attr *syscall.SysProcAttr, signal syscall.Signal)

func (fn parentDeathAttrFn) Supported() bool {
	return fn != nil
}

func (fn parentDeathAttrFn) Set(attr *syscall.SysProcAttr, signal syscall.Signal) {
	if fn.Supported() {
		fn(attr, signal)
	}
}
