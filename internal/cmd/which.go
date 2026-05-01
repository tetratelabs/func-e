// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/tetratelabs/func-e/internal/envoy"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/runtime"
)

type cmdWhich struct{}

func (c *cmdWhich) Run(ctx context.Context, o *globals.GlobalOpts, w io.Writer) error {
	o.Quiet = true
	if err := runtime.EnsureEnvoyVersion(ctx, o); err != nil {
		return err
	}
	if err := o.Mkdirs(); err != nil {
		return err
	}
	ev, err := envoy.InstallIfNeeded(ctx, o)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%s\n", ev)
	return nil
}
