// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/moreos"
)

// TestFuncEWhich shows the path to the Envoy binary
func TestFuncEWhich(t *testing.T) {
	o := setupTest(t)

	c, stdout, stderr := newApp(o)

	require.NoError(t, c.Run([]string{"func-e", "which"}))
	envoyPath := filepath.Join(o.HomeDir, "versions", o.EnvoyVersion.String(), "bin", "envoy"+moreos.Exe)
	require.Equal(t, moreos.Sprintf("%s\n", envoyPath), stdout.String())
	require.Empty(t, stderr)
}
