// Copyright 2025 Tetrate
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

package cmd_test

import (
	"os"
	"testing"

	"github.com/tetratelabs/func-e/internal"
	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/test/build"
)

// fakeEnvoyBin holds a path to the compiled internal.FakeEnvoySrcPath
var fakeEnvoyBin string

func TestMain(m *testing.M) {
	var err error
	if fakeEnvoyBin, err = build.GoBuild(internal.FakeEnvoySrcPath, os.TempDir()); err != nil {
		moreos.Fprintf(os.Stderr, `failed to start cmd tests due to build error: %v\n`, err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
