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

package cmd_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/version"
)

func TestGetEnvoyVersions(t *testing.T) {
	o, cleanup := setupTest(t)
	defer cleanup()

	// Run "getenvoy versions"
	c, stdout, stderr := newApp(o)
	o.Out = stdout
	err := c.Run([]string{"getenvoy", "versions"})

	// Verify the command invoked, passing the correct default commandline
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(`VERSION	RELEASE_DATE
%s	2020-12-31
`, version.LastKnownEnvoy), stdout.String())
	require.Empty(t, stderr)
}
