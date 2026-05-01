// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"

	"github.com/tetratelabs/func-e/internal/envoy"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/runtime"
	"github.com/tetratelabs/func-e/internal/version"
)

type cmdUse struct {
	Version string `arg:"" name:"version" help:"Envoy version to use (e.g. 1.38.0 or 1.38)"`
}

func (c *cmdUse) Run(ctx context.Context, o *globals.GlobalOpts) error {
	v, err := version.NewVersion("[version] argument", c.Version)
	if err != nil {
		return NewValidationError(err.Error())
	}

	if err := o.Mkdirs(); err != nil {
		return err
	}
	if o.EnvoyVersion, err = runtime.EnsurePatchVersion(ctx, o, v); err != nil {
		return err
	}
	if _, err = envoy.InstallIfNeeded(ctx, o); err != nil {
		return err
	}
	return envoy.WriteCurrentVersion(v, o.ConfigHome, o.EnvoyVersionFile())
}
