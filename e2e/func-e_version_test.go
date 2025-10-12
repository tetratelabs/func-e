// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFuncEVersion(t *testing.T) {
	t.Parallel()

	stdout, stderr, err := funcEExec(t.Context(), "--version")

	require.Regexp(t, `^func-e version ([^\s]+)\r?\n$`, stdout)
	require.Empty(t, stderr)
	require.NoError(t, err)
}
