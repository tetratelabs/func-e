// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package run

import (
	"context"
	"os"

	"github.com/tetratelabs/func-e/api"
	internalapi "github.com/tetratelabs/func-e/internal/api"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/runtime"
	"github.com/tetratelabs/func-e/internal/version"
)

// EnvoyPath overrides the path to the Envoy binary. Used for testing with a fake binary.
func EnvoyPath(envoyPath string) api.RunOption {
	return func(o *internalapi.RunOpts) {
		o.EnvoyPath = envoyPath
	}
}

// Run implements api.RunFunc
func Run(ctx context.Context, args []string, options ...api.RunOption) error {
	// Check if middleware is set in context
	baseRun := api.RunFunc(runImpl)
	if middlewareVal := ctx.Value(internalapi.RunMiddlewareKey{}); middlewareVal != nil {
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
	return runtime.Run(ctx, o, args)
}

func initOpts(ctx context.Context, options ...api.RunOption) (*globals.GlobalOpts, error) {
	ro := &internalapi.RunOpts{
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
		ConfigHome:   ro.ConfigHome,
		DataHome:     ro.DataHome,
		StateHome:    ro.StateHome,
		RuntimeDir:   ro.RuntimeDir,
		RunOpts: globals.RunOpts{
			EnvoyPath:   ro.EnvoyPath,
			EnvoyOut:    ro.EnvoyOut,
			EnvoyErr:    ro.EnvoyErr,
			StartupHook: ro.StartupHook,
			// TempDir is set later in initializeRunOpts via EnvoyRuntimeDir(runID)
		},
	}
	// Note: api.HomeDir() sets ConfigHome, DataHome, StateHome, RuntimeDir to same value (legacy mode)
	// Legacy detection happens in InitializeGlobalOpts when all four match
	homeDir := ""
	if ro.ConfigHome != "" && ro.ConfigHome == ro.DataHome && ro.DataHome == ro.StateHome && ro.StateHome == ro.RuntimeDir {
		homeDir = ro.ConfigHome // Legacy mode
	}
	if err := runtime.InitializeGlobalOpts(o, ro.EnvoyVersionsURL, homeDir, ro.ConfigHome, ro.DataHome, ro.StateHome, ro.RuntimeDir, "", ro.RunID); err != nil {
		return nil, err
	}

	if err := runtime.EnsureEnvoyVersion(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}
