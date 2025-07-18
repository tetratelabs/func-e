// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import "syscall"

// processGroupAttr returns nil on non-Linux as they lack Pdeathsig.
func processGroupAttr() *syscall.SysProcAttr {
	return nil
}
