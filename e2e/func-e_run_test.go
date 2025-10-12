// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"testing"

	"github.com/tetratelabs/func-e/internal/test/e2e"
)

func TestRun(t *testing.T) {
	e2e.TestRun(t.Context(), t, funcEFactory{})
}

func TestRun_AdminAddressPath(t *testing.T) {
	e2e.TestRun_AdminAddressPath(t.Context(), t, funcEFactory{})
}

func TestRun_LogWarn(t *testing.T) {
	e2e.TestRun_LogWarn(t.Context(), t, funcEFactory{})
}

func TestRun_RunDirectory(t *testing.T) {
	e2e.TestRun_RunDirectory(t.Context(), t, funcEFactory{})
}

func TestRun_InvalidConfig(t *testing.T) {
	e2e.TestRun_InvalidConfig(t.Context(), t, funcEFactory{})
}

func TestRun_StaticFile(t *testing.T) {
	e2e.TestRun_StaticFile(t.Context(), t, funcEFactory{})
}

func TestRun_CtrlCs(t *testing.T) {
	e2e.TestRun_CtrlCs(t.Context(), t, funcEFactory{})
}
