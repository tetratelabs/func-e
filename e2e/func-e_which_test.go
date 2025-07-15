// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/version"
)

// TestFuncEWhich ensures the command can show the current version in use. This can't use version.LastKnownEnvoy without
// explicitly downloading it first, because the latest version on Linux and macOS can be ahead of that due to routine
// lagging on Homebrew maintenance (OS/x), or lag in someone re-releasing on archive-envoy after Homebrew is updated.
func TestFuncEWhich(t *testing.T) { // not parallel as it can end up downloading concurrently
	// Explicitly issue "use" for the last known version to ensure when latest is ahead of this, the test doesn't fail.
	_, _, err := funcEExec("use", version.LastKnownEnvoy.String())
	require.NoError(t, err)

	stdout, stderr, err := funcEExec("which")
	relativeEnvoyBin := filepath.Join("versions", version.LastKnownEnvoy.String(), "bin", "envoy"+moreos.Exe)
	require.Contains(t, stdout, moreos.Sprintf("%s\n", relativeEnvoyBin))
	require.Empty(t, stderr)
	require.NoError(t, err)
}
