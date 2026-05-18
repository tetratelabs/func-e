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
	e2e.TestRunAdminAddressPath(t.Context(), t, funcEFactory{})
}

func TestRun_LogWarn(t *testing.T) {
	e2e.TestRunLogWarn(t.Context(), t, funcEFactory{})
}

func TestRun_RunDir(t *testing.T) {
	// For binary e2e tests, state directory is controlled via FUNC_E_STATE_HOME env var
	// (CLI layer), not library API options like api.StateHome().
	stateDir := t.TempDir()
	t.Setenv("FUNC_E_STATE_HOME", stateDir)
	e2e.TestRunRunDir(t.Context(), t, funcEFactory{}, stateDir)
}

func TestRun_InvalidConfig(t *testing.T) {
	e2e.TestRunInvalidConfig(t.Context(), t, funcEFactory{})
}

func TestRun_StaticFile(t *testing.T) {
	e2e.TestRunStaticFile(t.Context(), t, funcEFactory{})
}

func TestRun_CtrlCs(t *testing.T) {
	e2e.TestRunCtrlCs(t.Context(), t, funcEFactory{})
}

func TestRun_LegacyHomeDir(t *testing.T) {
	e2e.TestRunLegacyHomeDir(t.Context(), t, funcEFactory{})
}

func TestRun_Dev(t *testing.T) {
	e2e.TestRunDev(t.Context(), t, funcEFactory{})
}
