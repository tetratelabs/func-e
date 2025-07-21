// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import "syscall"

// processGroupAttr sets SysProcAttr.Pdeathsig to syscall.SIGKILL, to avoid
// orphaning envoy, if func-e is kill -9'd. We don't test this because it isn't
// deterministic, and outside our control.
func processGroupAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}
}
