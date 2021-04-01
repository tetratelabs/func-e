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

package cmd

import (
	"os/user"

	. "github.com/onsi/ginkgo" //nolint:golint

	builtintoolchain "github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain/builtin"
)

// GetCurrentUserFunc returns the current user.
type GetCurrentUserFunc func() (*user.User, error)

// SetDefaultUser configures built-in toolchain to resolve the current user to a predefined value.
func SetDefaultUser() {
	SetUser(func() (*user.User, error) {
		return &user.User{Uid: "1001", Gid: "1002"}, nil
	})
}

// SetUser configures built-in toolchain to resolve the current user using a given function.
func SetUser(fn GetCurrentUserFunc) {
	var getCurrentUserBackup GetCurrentUserFunc

	BeforeEach(func() {
		getCurrentUserBackup = builtintoolchain.GetCurrentUser
	})

	AfterEach(func() {
		builtintoolchain.GetCurrentUser = getCurrentUserBackup
	})

	BeforeEach(func() {
		builtintoolchain.GetCurrentUser = fn
	})
}
