// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"bufio"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/version"
)

func TestFuncEVersions_NothingYet(t *testing.T) {
	dataHome := t.TempDir()

	stdout, stderr, err := funcEExec(t.Context(), "--data-home", dataHome, "versions")

	require.NoError(t, err)
	require.Empty(t, stdout)
	require.Empty(t, stderr)
}

func TestFuncEVersions(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := funcEExec(t.Context(), "versions")

	// Depending on ~/func-e/version, what's selected may not be the latest version or even installed at all.
	require.Regexp(t, "[ *] [1-9][0-9]*\\.[0-9]+\\.[0-9]+(_debug)? 202[1-9]-[01][0-9]-[0-3][0-9].*\n", stdout)
	require.Empty(t, stderr)
	require.NoError(t, err)
}

func TestFuncEVersions_All(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := funcEExec(t.Context(), "versions", "-a")

	require.Regexp(t, fmt.Sprintf("[ *] %s 202[1-9]-[01][0-9]-[0-3][0-9].*\n", version.LastKnownEnvoy), stdout)
	require.Empty(t, stderr)
	require.NoError(t, err)
}

func TestFuncEVersions_AllIncludesInstalled(t *testing.T) {
	t.Parallel()

	// Cheap test that one includes the other. It doesn't actually parse the output, but the above tests prove the
	// latest version is in each deviation.
	allVersions, _, err := funcEExec(t.Context(), "versions", "-a")
	require.NoError(t, err)
	installedVersions, _, err := funcEExec(t.Context(), "versions")
	require.NoError(t, err)

	require.Greater(t, countLines(allVersions), countLines(installedVersions), "expected more versions available than installed")
}

func countLines(stdout string) (count int) {
	s := bufio.NewScanner(strings.NewReader(stdout))
	for s.Scan() {
		count++
	}
	return count
}
