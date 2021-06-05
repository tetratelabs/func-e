// Copyright 2021 Tetrate
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
	"syscall"
)

func sysProcAttr() *syscall.SysProcAttr {
	// TODO: We don't use syscall.CREATE_NEW_PROCESS_GROUP because we want external killing of getenvoy to kill the
	// spawned process. If we put this in a new process group, that wouldn't happen. Either way this is as yet untested.
	return &syscall.SysProcAttr{}
}
