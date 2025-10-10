// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package run

import (
	"context"
	"os"

	"github.com/tetratelabs/func-e/api"
	internalapi "github.com/tetratelabs/func-e/internal/api"
	"github.com/tetratelabs/func-e/internal/globals"
	internalmiddleware "github.com/tetratelabs/func-e/internal/middleware"
	"github.com/tetratelabs/func-e/internal/opts"
	"github.com/tetratelabs/func-e/internal/version"
)

// EnvoyPath overrides the path to the Envoy binary. Used for testing with a fake binary.
func EnvoyPath(envoyPath string) api.RunOption {
	return func(o *opts.RunOpts) {
		o.EnvoyPath = envoyPath
	}
}

// Run implements api.RunFunc
func Run(ctx context.Context, args []string, options ...api.RunOption) error {
	// Check if middleware is set in context
	baseRun := api.RunFunc(runImpl)
	if middlewareVal := ctx.Value(internalmiddleware.MiddlewareKey{}); middlewareVal != nil {
		// Type assert to function that matches our middleware signature
		if middleware, ok := middlewareVal.(func(api.RunFunc) api.RunFunc); ok {
			baseRun = middleware(baseRun)
		}
	}

	return baseRun(ctx, args, options...)
}

// runImpl is the default implementation of api.RunFunc
func runImpl(ctx context.Context, args []string, options ...api.RunOption) error {
	o, err := initOpts(ctx, options...)
	if err != nil {
		return err
	}
	return internalapi.Run(ctx, o, args)
}

func initOpts(ctx context.Context, options ...api.RunOption) (*globals.GlobalOpts, error) {
	ro := &opts.RunOpts{
		Out:      os.Stdout,
		EnvoyOut: os.Stdout,
		EnvoyErr: os.Stderr,
	}
	for _, option := range options {
		option(ro)
	}

	o := &globals.GlobalOpts{
		EnvoyVersion: version.PatchVersion(ro.EnvoyVersion),
		Out:          ro.Out,
		RunOpts: globals.RunOpts{
			EnvoyPath:   ro.EnvoyPath,
			EnvoyOut:    ro.EnvoyOut,
			EnvoyErr:    ro.EnvoyErr,
			StartupHook: ro.StartupHook,
		},
	}
	if err := internalapi.InitializeGlobalOpts(o, ro.EnvoyVersionsURL, ro.HomeDir, ""); err != nil {
		return nil, err
	}

	if err := internalapi.EnsureEnvoyVersion(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}
