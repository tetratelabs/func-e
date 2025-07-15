// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

// Package morerequire includes more require functions than "github.com/stretchr/testify/require"
// Do not add dependencies on any main code as it will cause cycles.
package morerequire

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// RequireSetMtime sets the mtime of the dir given a string formatted date. Ex "2006-01-02"
func RequireSetMtime(t *testing.T, dir, date string) {
	// Make sure we do parsing using time.Local to match the time.ModTime's location used for sorting.
	td, err := time.ParseInLocation("2006-01-02", date, time.Local)
	require.NoError(t, err)
	require.NoError(t, os.Chtimes(dir, td, td))
}

// RequireChdir changes the working directory reverts it on the returned function
func RequireChdir(t *testing.T, dir string) func() {
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))

	return func() {
		require.NoError(t, os.Chdir(wd))
	}
}
