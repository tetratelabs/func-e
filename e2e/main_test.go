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

package e2e

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tetratelabs/func-e/internal/moreos"
)

//nolint:golint
const (
	// funcEPathEnvKey holds the path to funcEBin. Defaults to the project root (`$PWD/..`).
	funcEPathEnvKey        = "E2E_FUNC_E_PATH"
	envoyVersionsURLEnvKey = "ENVOY_VERSIONS_URL"
	envoyVersionsJSON      = "envoy-versions.json"
	runTimeout             = 2 * time.Minute
)

var (
	funcEBin            string // funcEBin holds a path to a 'func-e' binary under test.
	expectedMockHeaders = map[string]string{"User-Agent": "func-e/dev"}
)

// TestMain ensures the "func-e" binary is valid.
func TestMain(m *testing.M) {
	// As this is an e2e test, we execute all tests with a binary compiled earlier.
	if err := readFuncEBin(); err != nil {
		exitOnInvalidBinary(err)
	}

	versionLine, _, err := funcEExec("--version")
	if err != nil {
		exitOnInvalidBinary(err)
	}

	// Allow local file override when a SNAPSHOT version
	if _, err := os.Stat(envoyVersionsJSON); err == nil && strings.Contains(versionLine, "SNAPSHOT") {
		s, err := mockEnvoyVersionsServer() // no defer s.Close() because os.Exit() subverts it
		if err != nil {
			moreos.Fprintf(os.Stderr, "failed to serve %s: %v\n", envoyVersionsJSON, err)
			os.Exit(1)
		}
		os.Setenv(envoyVersionsURLEnvKey, s.URL)
	}
	os.Exit(m.Run())
}

func exitOnInvalidBinary(err error) {
	moreos.Fprintf(os.Stderr, `failed to start e2e tests due to an invalid "func-e" binary: %v\n`, err)
	os.Exit(1)
}

// mockEnvoyVersionsServer ensures envoyVersionsURLEnvKey is set appropriately, so that non-release versions can see
// changes to local envoyVersionsJSON.
func mockEnvoyVersionsServer() (*httptest.Server, error) {
	f, err := os.Open(envoyVersionsJSON)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ensure e2e tests won't eventually interfere with analytics when run against a release version
		for k, v := range expectedMockHeaders {
			h := r.Header.Get(k)
			if h != v {
				w.WriteHeader(500)
				_, _ = w.Write([]byte(moreos.Sprintf("invalid %q: %s != %s\n", k, h, v)))
				return
			}
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write(b)
	}))
	return ts, nil
}

// readFuncEBin reads E2E_FUNC_E_PATH or defaults to the project root (`$PWD/..`) to find func-e.
// An error is returned if the value isn't an executable file.
func readFuncEBin() error {
	path := os.Getenv(funcEPathEnvKey)
	if path != "" {
		p, err := abs(path)
		if err != nil {
			return err
		}
		path = filepath.Clean(p)
		stat, err := os.Stat(path)
		if err != nil && os.IsNotExist(err) {
			return fmt.Errorf("%s doesn't exist. Correct environment variable %s", path, funcEPathEnvKey)
		}
		if !stat.IsDir() {
			return fmt.Errorf("%s is not a directory. Correct environment variable %s", path, funcEPathEnvKey)
		}
	} else {
		// We need to make the path relative to the project root because "e2e" tests run in the "e2e" directory.
		abs, err := filepath.Abs("..")
		if err != nil {
			return err
		}
		path = abs
	}

	// Now, check the binary at the path
	funcEBin = filepath.Join(path, "func-e"+moreos.Exe)
	stat, err := os.Stat(funcEBin)
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("%s doesn't exist.  Run `go build .` or `make bin`", funcEBin)
	}

	// While "make bin" should result in correct permissions, double-check as some tools lose them, such as
	// https://github.com/actions/upload-artifact#maintaining-file-permissions-and-case-sensitive-files
	if !moreos.IsExecutable(stat) {
		return fmt.Errorf("%s is not executable. Run `go build .` or `make bin`", funcEBin)
	}
	fmt.Fprintln(os.Stderr, "using", funcEBin)
	return nil
}

// abs is like filepath.Abs except the correct relative dir is '..'
func abs(path string) (string, error) {
	if !filepath.IsAbs(path) {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("could not get current directory")
		}
		path = filepath.Join(wd, "..", path)
	}
	return path, nil
}

type funcE struct {
	cmd      *exec.Cmd
	runDir   string
	envoyPid int32
}

func newFuncE(ctx context.Context, args ...string) *funcE {
	cmd := exec.CommandContext(ctx, funcEBin, args...)
	cmd.SysProcAttr = moreos.ProcessGroupAttr()
	return &funcE{cmd: cmd}
}

func (b *funcE) String() string {
	return strings.Join(b.cmd.Args, " ")
}

func funcEExec(args ...string) (string, string, error) {
	g := newFuncE(context.Background(), args...)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	g.cmd.Stdout = io.MultiWriter(os.Stdout, stdout) // we want to see full `func-e` output in the test log
	g.cmd.Stderr = io.MultiWriter(os.Stderr, stderr)
	err := g.cmd.Run()
	return stdout.String(), stderr.String(), err
}
