// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestFuncEWhich shows the path to the Envoy binary
func TestFuncEWhich(t *testing.T) {
	o := setupTest(t)

	c, stdout, stderr := newApp(o)

	require.NoError(t, c.RunContext(t.Context(), []string{"func-e", "which"}))
	envoyPath := filepath.Join(o.DataHome, "envoy-versions", o.EnvoyVersion.String(), "bin", "envoy")
	require.Equal(t, envoyPath+"\n", stdout.String())
	require.Empty(t, stderr)
}
