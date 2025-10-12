// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package run

import (
	"testing"

	"github.com/tetratelabs/func-e/internal/test/e2e"
)

func TestRun(t *testing.T) {
	e2e.TestRun(t.Context(), t, fakeFuncEFactory{})
}

func TestRun_AdminAddressPath(t *testing.T) {
	e2e.TestRun_AdminAddressPath(t.Context(), t, fakeFuncEFactory{})
}

func TestRun_LogWarn(t *testing.T) {
	e2e.TestRun_LogWarn(t.Context(), t, fakeFuncEFactory{})
}

func TestRun_RunDir(t *testing.T) {
	// Test that FUNC_E_STATE_HOME env var works correctly to control
	// where runtime-generated files (logs, config_dump.json) are written.
	stateDir := t.TempDir()
	t.Setenv("FUNC_E_STATE_HOME", stateDir)
	e2e.TestRun_RunDir(t.Context(), t, fakeFuncEFactory{}, stateDir)
}

func TestRun_InvalidConfig(t *testing.T) {
	e2e.TestRun_InvalidConfig(t.Context(), t, fakeFuncEFactory{})
}

func TestRun_StaticFile(t *testing.T) {
	e2e.TestRun_StaticFile(t.Context(), t, fakeFuncEFactory{})
}

func TestRun_CtrlCs(t *testing.T) {
	// This doesn't call ctrl-c, rather cancels the context multiple times
	e2e.TestRun_CtrlCs(t.Context(), t, fakeFuncEFactory{})
}

func TestRun_LegacyHomeDir(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("FUNC_E_HOME", homeDir)
	e2e.TestRun_LegacyHomeDir(t.Context(), t, fakeFuncEFactory{})
}
