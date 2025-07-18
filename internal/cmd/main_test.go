// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package cmd_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/tetratelabs/func-e/internal"
	"github.com/tetratelabs/func-e/internal/test/build"
)

// fakeEnvoyBin holds a path to the compiled internal.FakeEnvoySrcPath
var fakeEnvoyBin string

func TestMain(m *testing.M) {
	var err error
	if fakeEnvoyBin, err = build.GoBuild(internal.FakeEnvoySrcPath, os.TempDir()); err != nil {
		fmt.Fprintf(os.Stderr, `failed to start cmd tests due to build error: %v\n`, err) //nolint:errcheck
		os.Exit(1)
	}
	os.Exit(m.Run())
}
