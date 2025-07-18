// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

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

// AccessLogYaml shows a listener with access logging to stdout
//
//go:embed envoy/config/testdata/access_log.yaml
var AccessLogYaml []byte

// AccessLogTypedConfigYaml is the critical configuration in AccessLogYaml
//
//go:embed envoy/config/testdata/access_log_typed_config.yaml
var AccessLogTypedConfigYaml string

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
