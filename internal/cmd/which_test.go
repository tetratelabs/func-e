// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	rootcmd "github.com/tetratelabs/func-e/internal/cmd"
)

// TestFuncEWhich shows the path to the Envoy binary
func TestFuncEWhich(t *testing.T) {
	o := setupTest(t)

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	require.NoError(t, rootcmd.DoMain(t.Context(), stdout, stderr, []string{"which"}, o, "test"))
	envoyPath := filepath.Join(o.DataHome, "envoy-versions", o.EnvoyVersion.String(), "bin", "envoy")
	require.Equal(t, envoyPath+"\n", stdout.String())
	require.Empty(t, stderr.String())
}
