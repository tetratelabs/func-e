// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package func_e

import (
	"context"

	"github.com/tetratelabs/func-e/api"
	"github.com/tetratelabs/func-e/internal/run"
)

// Run is the default implementation of api.RunFunc.
func Run(ctx context.Context, args []string, options ...api.RunOption) error {
	return run.Run(ctx, args, options...)
}
