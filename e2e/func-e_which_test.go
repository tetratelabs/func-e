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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestFuncEWhich(t *testing.T) { // not parallel as it can end up downloading concurrently
	stdout, stderr, err := funcEExec("which")
	relativeEnvoyBin := filepath.Join("versions", version.LastKnownEnvoy.String(), "bin", "envoy"+moreos.Exe)
	require.Contains(t, stdout, moreos.Sprintf("%s\n", relativeEnvoyBin))
	require.Empty(t, stderr)
	require.NoError(t, err)
}
