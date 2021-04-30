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
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/log"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
	e2e "github.com/tetratelabs/getenvoy/test/e2e/util"
)

var (
	// getEnvoy is the absolute path to the "getenvoy" binary used in all tests.
	getEnvoy = e2e.GetEnvoy
	// "getenvoy extension" tests default to run these
	extensionLanguages []extension.Language

	// stripAnsiEscapeRegexp is a regular expression to clean ANSI Control sequences
	// feat https://stackoverflow.com/questions/14693701/how-can-i-remove-the-ansi-escape-sequences-from-a-string-in-python#33925425
	stripAnsiEscapeRegexp = regexp.MustCompile(`(\x9B|\x1B\[)[0-?]*[ -/]*[@-~]`)
)

//nolint:golint
const (
	E2E_EXTENSION_LANGUAGE          = "E2E_EXTENSION_LANGUAGE"
	E2E_GETENVOY_BINARY             = "E2E_GETENVOY_BINARY"
	E2E_TOOLCHAIN_CONTAINER_OPTIONS = "E2E_TOOLCHAIN_CONTAINER_OPTIONS"
)

// ExtensionTestCase represents a combination of extension category and  programming language.
type extensionTestCase struct {
	extension.Category
	extension.Language
}

func (t extensionTestCase) String() string {
	return fmt.Sprintf("category=%s, language=%s", t.Category, t.Language)
}

// getExtensionTestMatrix returns the base matrix of category and language "getenvoy extension" tests run.
func getExtensionTestMatrix() []extensionTestCase {
	tuples := make([]extensionTestCase, 0)
	for _, category := range extension.Categories {
		for _, language := range extensionLanguages {
			tuples = append(tuples, extensionTestCase{category, language})
		}
	}
	return tuples
}

// TestMain ensures the "getenvoy" binary and "--language" parameter to "get envoy init" are valid, as these are
// constant for all tests that us them.
//
// Note: "getenvoy extension build" and commands that imply it, can be extremely slow due to implicit responsibilities
// such as downloading modules or compilation. Commands like this use Docker, so changes to the Dockerfile or contents
// like "commands.sh" will effect performance.
//
// Note: Pay close attention to values of util.E2E_TOOLCHAIN_CONTAINER_OPTIONS as these can change assumptions.
// CI may override this to set HOME or CARGO_HOME (rust) used by "getenvoy" and effect its execution.
func TestMain(m *testing.M) {
	// As this is an e2e test, we execute all tests with a binary compiled earlier.
	path, err := getEnvoyPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, `failed to start e2e tests due to an invalid "getenvoy" binary: %v`, err)
		os.Exit(1)
	}
	e2e.GetEnvoyPath = path
	extensionLanguages, err = getExtensionLanguages()
	if err != nil {
		fmt.Fprintf(os.Stderr, `failed to start e2e tests due to an invalid extension language": %v`, err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// getExtensionLanguage reads E2E_EXTENSION_LANGUAGE or defaults to extension.LanguageTinyGo because it builds an order
// of magnitude faster extension.LanguageRust. All languages test when "E2E_EXTENSION_LANGUAGE=all".
func getExtensionLanguages() ([]extension.Language, error) {
	fromEnv := os.Getenv(E2E_EXTENSION_LANGUAGE)
	if fromEnv == "all" {
		return extension.Languages, nil
	} else if fromEnv == "" {
		fromEnv = extension.LanguageTinyGo.String() // default
	}
	parsed, err := extension.ParseLanguage(fromEnv)
	if err != nil {
		return nil, fmt.Errorf("%s is not a valid extension language. Correct environment variable %s", fromEnv, E2E_GETENVOY_BINARY)
	}
	return []extension.Language{parsed}, nil
}

// getEnvoyPath reads E2E_GETENVOY_BINARY or defaults to "$PWD/build/bin/$GOOS/$GOARCH/getenvoy"
// An error is returned if the value isn't an executable file.
func getEnvoyPath() (string, error) {
	path := os.Getenv(E2E_GETENVOY_BINARY)
	if path == "" {
		// Assemble the default created by "make bin"
		relativePath := filepath.Join("..", "..", "build", "bin", runtime.GOOS, runtime.GOARCH, "getenvoy")
		abs, err := filepath.Abs(relativePath)
		if err != nil {
			return "", fmt.Errorf("%s didn't resolve to a valid path. Correct environment variable %s", path, E2E_GETENVOY_BINARY)
		}
		path = abs
	}
	stat, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return "", fmt.Errorf("%s doesn't exist. Correct environment variable %s", path, E2E_GETENVOY_BINARY)
	}
	if stat.IsDir() {
		return "", fmt.Errorf("%s is not a file. Correct environment variable %s", path, E2E_GETENVOY_BINARY)
	}
	// While "make bin" should result in correct permissions, double-check as some tools lose them, such as
	// https://github.com/actions/upload-artifact#maintaining-file-permissions-and-case-sensitive-files
	if stat.Mode()&0111 == 0 {
		return "", fmt.Errorf("%s is not executable. Correct environment variable %s", path, E2E_GETENVOY_BINARY)
	}
	return path, nil
}

func getToolchainContainerOptions() []string {
	value := os.Getenv(E2E_TOOLCHAIN_CONTAINER_OPTIONS)
	if value == "" {
		return nil
	}
	return []string{"--toolchain-container-options", value}
}

// Command gives us an interface needed for testing getEnvoy
type Command interface {
	Exec() (string, string, error)
}

// requireExecNoStdout invokes the command and returns its stderr if successful and stdout is empty.
func requireExecNoStdout(t *testing.T, c Command) string {
	stdout, stderr := requireExec(t, c)
	require.Empty(t, stdout, `expected no stdout running [%v]`, c)
	require.NotEmpty(t, stderr, `expected stderr running [%v]`, c)
	return stderr
}

// requireExecNoStderr invokes the command and returns its stdout if successful and stderr is empty.
func requireExecNoStderr(t *testing.T, c Command) string {
	stdout, stderr := requireExec(t, c)
	require.NotEmpty(t, stdout, `expected stdout running [%v]`, c)
	require.Empty(t, stderr, `expected no stderr running [%v]`, c)
	return stdout
}

// requireExec invokes the command and returns its (stdout, stderr) if successful.
func requireExec(t *testing.T, c Command) (string, string) {
	log.Infof(`running [%v]`, c)
	stdout, stderr, err := c.Exec()

	require.NoError(t, err, `error running [%v]`, c)
	return stdout, stderr
}

// requireExtensionInit is useful for tests that depend on "getenvoy extension init" as a prerequisite.
func requireExtensionInit(t *testing.T, extensionDir string, category extension.Category, language extension.Language, name string) {
	c := getEnvoy("extension init").
		Arg(extensionDir).
		Arg("--category").Arg(string(category)).
		Arg("--language").Arg(string(language)).
		Arg("--name").Arg(name)
	// stderr returned is not tested because doing so is redundant to TestGetEnvoyExtensionInit.
	_ = requireExecNoStdout(t, c)
}

// requireExtensionInit is useful for tests that depend on "getenvoy extension build" as a prerequisite.
// The result of calling this is the bytes representing the built wasm
func requireExtensionBuild(t *testing.T, language extension.Language, workingDir string) []byte {
	c := getEnvoy("extension build").Args(getToolchainContainerOptions()...).WorkingDir(workingDir)
	// stderr returned is not tested because doing so is redundant to TestGetEnvoyExtensionInit.
	_ = requireExecNoStderr(t, c)

	extensionWasmFile := filepath.Join(workingDir, extensionWasmPath(language))
	require.FileExists(t, extensionWasmFile, `extension wasm file %s missing after running [%v]`, extensionWasmFile, c)

	wasmBytes, err := os.ReadFile(extensionWasmFile)
	require.NoError(t, err, `error reading %s after running [%v]: %s`, extensionWasmFile, c)
	require.NotEmpty(t, wasmBytes, `%s empty after running [%v]`, extensionWasmFile, c)
	return wasmBytes
}

// requireExtensionClean is useful for tests that depend on "getenvoy extension clean" on completion.
// (stdout, stderr) returned are not tested because they can both be empty.
func requireExtensionClean(t *testing.T, workingDir string) {
	c := getEnvoy("extension clean").WorkingDir(workingDir)
	_, _ = requireExec(t, c)
}

// extensionWasmPath returns the language-specific location of the extension.wasm.
func extensionWasmPath(language extension.Language) string {
	switch language {
	case extension.LanguageRust:
		return filepath.Join("target", "getenvoy", "extension.wasm")
	case extension.LanguageTinyGo:
		return filepath.Join("build", "extension.wasm")
	}
	panic("unsupported extension language " + language)
}

// extensionWasmPath returns the language-specific basename of the extension config file.
func extensionConfigFileName(language extension.Language) string {
	switch language {
	case extension.LanguageTinyGo:
		return "extension.txt"
	default:
		return "extension.json"
	}
}
