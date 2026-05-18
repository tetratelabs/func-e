// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
)

// TestMain ensures the "func-e" binary is valid.
func TestMain(m *testing.M) {
	// As this is an e2e test, we execute all tests with a binary compiled earlier.
	if err := readOrBuildFuncEBin(); err != nil {
		exitOnInvalidBinary(err)
	}

	// pre-flight check the binary is usable
	if _, _, err := funcEExec(context.Background(), "--version"); err != nil {
		exitOnInvalidBinary(err)
	}
	os.Exit(m.Run())
}

func exitOnInvalidBinary(err error) {
	fmt.Fprintf(os.Stderr, `failed to start e2e tests due to an invalid "func-e" binary: %v\n`, err)
	os.Exit(1)
}
