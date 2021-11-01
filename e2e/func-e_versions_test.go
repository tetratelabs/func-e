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
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestFuncEVersions_NothingYet(t *testing.T) {
	homeDir := t.TempDir()

	stdout, stderr, err := funcEExec("--home-dir", homeDir, "versions")

	require.NoError(t, err)
	require.Empty(t, stdout)
	require.Empty(t, stderr)
}

func TestFuncEVersions(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := funcEExec("versions")

	// Depending on ~/func-e/version, what's selected may not be the latest version or even installed at all.
	require.Regexp(t, moreos.Sprintf("[ *] [1-9][0-9]*\\.[0-9]+\\.[0-9]+(_debug)? 202[1-9]-[01][0-9]-[0-3][0-9].*\n"), stdout)
	require.Empty(t, stderr)
	require.NoError(t, err)
}

func TestFuncEVersions_All(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := funcEExec("versions", "-a")

	require.Regexp(t, moreos.Sprintf("[ *] %s 202[1-9]-[01][0-9]-[0-3][0-9].*\n", version.LastKnownEnvoy), stdout)
	require.Empty(t, stderr)
	require.NoError(t, err)
}

func TestFuncEVersions_AllIncludesInstalled(t *testing.T) {
	t.Parallel()

	// Cheap test that one includes the other. It doesn't actually parse the output, but the above tests prove the
	// latest version is in each deviation.
	allVersions, _, err := funcEExec("versions", "-a")
	require.NoError(t, err)
	installedVersions, _, err := funcEExec("versions")
	require.NoError(t, err)

	require.Greater(t, countLines(allVersions), countLines(installedVersions), "expected more versions available than installed")
}

func countLines(stdout string) (count int) {
	s := bufio.NewScanner(strings.NewReader(stdout))
	for s.Scan() {
		count++
	}
	return
}
