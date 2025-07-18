// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"testing"

	"github.com/tetratelabs/func-e/internal/test/e2e"
)

func TestRun(t *testing.T) {
	e2e.TestRun(context.Background(), t, funcEFactory{})
}

func TestRun_MinimalListener(t *testing.T) {
	e2e.TestRun_MinimalListener(context.Background(), t, funcEFactory{})
}

func TestRun_InvalidConfig(t *testing.T) {
	e2e.TestRun_InvalidConfig(context.Background(), t, funcEFactory{})
}

func TestRun_StaticFile(t *testing.T) {
	e2e.TestRun_StaticFile(context.Background(), t, funcEFactory{})
}

func TestRun_CtrlCs(t *testing.T) {
	e2e.TestRun_CtrlCs(context.Background(), t, funcEFactory{})
}

func TestRun_Kill9(t *testing.T) {
	e2e.TestRun_Kill9(context.Background(), t, funcEFactory{})
}
