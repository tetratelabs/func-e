// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/globals"
)

const siteManpageFile = "../../packaging/nfpm/func-e.8"

func TestManPageMatchesCommands(t *testing.T) {
	app := NewApp(&globals.GlobalOpts{})

	expected, err := app.ToMan()
	require.NoError(t, err)

	actual, err := os.ReadFile(siteManpageFile)
	require.NoError(t, err)
	require.Equal(t, expected, string(actual))
}
