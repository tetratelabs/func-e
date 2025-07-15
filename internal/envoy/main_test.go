// Copyright 2025 Tetrate
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"os"
	"testing"

	"github.com/tetratelabs/func-e/internal"
	"github.com/tetratelabs/func-e/internal/moreos"
	"github.com/tetratelabs/func-e/internal/test/build"
)

// fakeEnvoyBin holds a path to the compiled internal.FakeEnvoySrcPath
var fakeEnvoyBin string

func TestMain(m *testing.M) {
	var err error
	if fakeEnvoyBin, err = build.GoBuild(internal.FakeEnvoySrcPath, os.TempDir()); err != nil {
		moreos.Fprintf(os.Stderr, `failed to start envoy tests due to build error: %v\n`, err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
