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

package test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	// fakeFuncESrcPath is a path to the test source file used to simulate func-e which runs envoy as its child.
	fakeFuncESrcPath = filepath.Join(funcEGoModDir, "internal", "test", "testdata", "fake_func-e", "fake_func-e.go")
	// fakeFuncEBin is the compiled code of fakeFuncESrcPath which will be runtime.GOOS dependent.
	fakeFuncEBin   []byte
	builtFakeFuncE sync.Once
)

// RequireFakeFuncE writes fakeFuncEBin to the given path.
func RequireFakeFuncE(t *testing.T, path string) {
	builtFakeFuncE.Do(func() {
		fakeFuncEBin = requireBuildFakeFuncE(t)
	})
	require.NoError(t, os.WriteFile(path, fakeFuncEBin, 0700)) //nolint:gosec
}

// requireBuildFakeEnvoy builds a fake envoy binary and returns its contents.
func requireBuildFakeFuncE(t *testing.T) []byte {
	return requireBuildFakeBinary(t, "func-e", fakeBinarySrc{path: fakeFuncESrcPath})
}
