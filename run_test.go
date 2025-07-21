// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package func_e_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"

	func_e "github.com/tetratelabs/func-e"
	"github.com/tetratelabs/func-e/api"
	"github.com/tetratelabs/func-e/internal"
	"github.com/tetratelabs/func-e/internal/run"
	"github.com/tetratelabs/func-e/internal/test/build"
)

// TestFuncERun_InvalidConfig takes care to not duplicate test/e2e/testrun.go,
// but still give some coverage.
func TestFuncERun_InvalidConfig(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := func_e.Run(t.Context(), []string{}, api.HomeDir(t.TempDir()), run.EnvoyPath(fakeEnvoyBin),
		api.Out(&stdout),
		api.EnvoyOut(&stdout),
		api.EnvoyErr(&stderr))

	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode())

	require.Contains(t, stdout.String(), fakeEnvoyBin)
	require.Contains(t, stderr.String(), "At least one of --config-path or --config-yaml")
}

// fakeEnvoyBin holds a path to the compiled internal.FakeEnvoySrcPath
var fakeEnvoyBin string

func TestMain(m *testing.M) {
	var err error
	if fakeEnvoyBin, err = build.GoBuild(internal.FakeEnvoySrcPath, os.TempDir()); err != nil {
		fmt.Fprintf(os.Stderr, `failed to start api tests due to build error: %v\n`, err) //nolint:errcheck
		os.Exit(1)
	}
	os.Exit(m.Run())
}
