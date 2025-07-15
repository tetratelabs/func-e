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

package internal

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

// StaticFileYaml shows Envoy reading a file referenced from the current directory
//
//go:embed envoy/config/testdata/static_file.yaml
var StaticFileYaml []byte

// StaticFileTypedConfigYaml is the critical configuration in StaticFileYaml
//
//go:embed envoy/config/testdata/static_file_typed_config.yaml
var StaticFileTypedConfigYaml string

// MinimalYaml shows the smallest possible listener without admin server
//
//go:embed envoy/config/testdata/minimal.yaml
var MinimalYaml []byte

// MinimalTypedConfigYaml is the critical configuration in MinimalYaml
//
//go:embed envoy/config/testdata/minimal_typed_config.yaml
var MinimalTypedConfigYaml string

// FakeEnvoySrcPath is the absolute path to the fake Envoy source file used in tests.
var FakeEnvoySrcPath = absolutePath("testdata", "fake_envoy", "main.go")

// absolutePath returns the absolute path to a file relative to this source file.
// It panics if the file doesn't exist.
func absolutePath(parts ...string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("could not determine current file path")
	}
	dir := filepath.Dir(file)

	path := filepath.Join(append([]string{dir}, parts...)...)

	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		panic(fmt.Sprintf("required file not found: %s", path))
	}

	return path
}
