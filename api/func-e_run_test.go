// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"testing"

	"github.com/tetratelabs/func-e/internal/test/e2e"
)

func TestRun(t *testing.T) {
	e2e.TestRun(context.Background(), t, fakeFuncEFactory{})
}

func TestRun_RunDirectory(t *testing.T) {
	e2e.TestRun_RunDirectory(context.Background(), t, fakeFuncEFactory{})
}

func TestRun_InvalidConfig(t *testing.T) {
	e2e.TestRun_InvalidConfig(context.Background(), t, fakeFuncEFactory{})
}

func TestRun_StaticFile(t *testing.T) {
	e2e.TestRun_StaticFile(context.Background(), t, fakeFuncEFactory{})
}

func TestRun_CtrlCs(t *testing.T) {
	// This doesn't call ctrl-c, rather cancels the context multiple times
	e2e.TestRun_CtrlCs(context.Background(), t, fakeFuncEFactory{})
}
