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

package e2e_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/log"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
	e2e "github.com/tetratelabs/getenvoy/test/e2e/util"
)

var (
	// GetEnvoy is a convenient alias.
	GetEnvoy = e2e.GetEnvoy
)

// stripAnsiEscapeRegexp is a regular expression to clean ANSI Control sequences
// feat https://stackoverflow.com/questions/14693701/how-can-i-remove-the-ansi-escape-sequences-from-a-string-in-python#33925425
var stripAnsiEscapeRegexp = regexp.MustCompile(`(\x9B|\x1B\[)[0-?]*[ -/]*[@-~]`)

func requireEnvoyBinaryPath(t *testing.T) {
	path, err := e2e.Env.GetEnvoyBinary()
	require.NoError(t, err, `error reading path to getenvoy binary`)
	e2e.GetEnvoyBinaryPath = path
}

// requireNewTempDir creates a new directory. The function returned cleans it up.
func requireNewTempDir(t *testing.T) (string, func()) {
	d, err := ioutil.TempDir("", "")
	if err != nil {
		require.NoError(t, err, `ioutil.TempDir("", "") erred`)
	}
	dir := requireAbsDir(t, d)
	return dir, func() {
		e := os.RemoveAll(dir)
		require.NoError(t, e, `error removing directory: %v`, dir)
	}
}

// RequireChDir will os.Chdir into the indicated dir, panicing on any problem.
// The function returned reverts to the original.
func requireChDir(t *testing.T, d string) func() {
	// Save previous working directory to that it can be reverted later.
	previous, err := os.Getwd()
	require.NoError(t, err, `error determining current directory`)

	// Now, actually change to the directory.
	err = os.Chdir(d)
	require.NoError(t, err, `error changing to directory: %v`, d)
	return func() {
		e := os.Chdir(previous)
		require.NoError(t, e, `error changing to directory: %v`, previous)
	}
}

// requireAbsDir runs filepath.Abs and ensures there are no errors and the input is a directory.
func requireAbsDir(t *testing.T, d string) string {
	dir, err := filepath.Abs(d)
	require.NoError(t, err, `error determining absolute directory: %v`, d)
	require.DirExists(t, dir, `directory doesn't exist': %v`, dir)
	dir, err = filepath.EvalSymlinks(dir)
	require.NoError(t, err, `filepath.EvalSymlinks(%s) erred`, dir)
	require.NotEmpty(t, dir, `filepath.EvalSymlinks(%s) returned ""`)
	return dir
}

// Command gives us an interface needed for testing GetEnvoy
type Command interface {
	Exec() (string, string, error)
}

// requireExecNoStdout invokes the command and returns its stderr if successful and stdout is empty.
func requireExecNoStdout(t *testing.T, cmd Command) string {
	stdout, stderr := requireExec(t, cmd)
	require.Empty(t, stdout, `expected no stdout running [%v]`, cmd)
	require.NotEmpty(t, stderr, `expected stderr running [%v]`, cmd)
	return stderr
}

// requireExecNoStderr invokes the command and returns its stdout if successful and stderr is empty.
func requireExecNoStderr(t *testing.T, cmd Command) string {
	stdout, stderr := requireExec(t, cmd)
	require.NotEmpty(t, stdout, `expected stdout running [%v]`, cmd)
	require.Empty(t, stderr, `expected no stderr running [%v]`, cmd)
	return stdout
}

// requireExec invokes the command and returns its (stdout, stderr) if successful.
func requireExec(t *testing.T, cmd Command) (string, string) {
	log.Infof(`running [%v]`, cmd)
	stdout, stderr, err := cmd.Exec()

	require.NoError(t, err, `error running [%v]`, cmd)
	return stdout, stderr
}

// requireExtensionInit is useful for tests that depend on "getenvoy extension init" as a prerequisite.
func requireExtensionInit(t *testing.T, workDir string, category extension.Category, language extension.Language, name string) {
	cmd := GetEnvoy("extension init").
		Arg(workDir).
		Arg("--category").Arg(string(category)).
		Arg("--language").Arg(string(language)).
		Arg("--name").Arg(name)
	// stderr returned is not tested because doing so is redundant to TestGetEnvoyExtensionInit.
	_ = requireExecNoStdout(t, cmd)
}

// extensionWasmPath returns the language-specific location of the extension.wasm.
func extensionWasmPath(language extension.Language) string {
	switch language {
	case extension.LanguageRust:
		return filepath.Join("target", "getenvoy", "extension.wasm")
	case extension.LanguageTinyGo:
		return filepath.Join("build", "extension.wasm")
	}
	panic("unsupported language " + language)
}

// requireExtensionInit is useful for tests that depend on "getenvoy extension build" as a prerequisite.
// The result of calling this is the bytes representing the built wasm
func requireExtensionBuild(t *testing.T, language extension.Language, workDir string) []byte {
	cmd := GetEnvoy("extension build").Args(e2e.Env.GetBuiltinContainerOptions()...)
	// stderr returned is not tested because doing so is redundant to TestGetEnvoyExtensionInit.
	_ = requireExecNoStderr(t, cmd)

	extensionWasmFile := filepath.Join(workDir, extensionWasmPath(language))
	require.FileExists(t, extensionWasmFile, `extension wasm file %s missing after running [%v]`, extensionWasmFile, cmd)

	wasmBytes, err := ioutil.ReadFile(extensionWasmFile)
	require.NoError(t, err, `error reading %s after running [%v]: %s`, extensionWasmFile, cmd)
	require.NotEmpty(t, wasmBytes, `%s empty after running [%v]`, extensionWasmFile, cmd)
	return wasmBytes
}

// requireExtensionClean is useful for tests that depend on "getenvoy extension clean" on completion.
// (stdout, stderr) returned are not tested because they can both be empty.
func requireExtensionClean(t *testing.T, workDir string) {
	err := os.Chdir(workDir)
	require.NoError(t, err, `error changing to directory: %v`, workDir)

	cmd := GetEnvoy("extension clean")
	_, _ = requireExec(t, cmd)
}
