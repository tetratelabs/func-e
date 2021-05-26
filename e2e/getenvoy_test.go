// Copyright 2020 Tetrate
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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/version"
)

func TestGetEnvoyVersion(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := getEnvoy("--version").exec()

	require.Regexp(t, `^getenvoy version ([^\s]+)\n$`, stdout)
	require.Equal(t, ``, stderr)
	require.NoError(t, err)
}

func TestGetEnvoyVersions(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := getEnvoy("versions").exec()

	require.Regexp(t, "^VERSION\tRELEASE_DATE\n", stdout)
	require.Regexp(t, fmt.Sprintf("%s\t202[1-9]-[01][0-9]-[0-3][0-9]\n", version.LastKnownEnvoy), stdout)
	require.Empty(t, stderr)
	require.NoError(t, err)
}

func TestGetEnvoyString(t *testing.T) {
	t.Parallel()

	g := getEnvoy("--version")

	require.Regexp(t, `.*getenvoy --version$`, g.String())
}
