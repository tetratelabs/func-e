// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"context"

	"github.com/tetratelabs/func-e/api"
	"github.com/tetratelabs/func-e/internal/admin"
	internalapi "github.com/tetratelabs/func-e/internal/api"
)

// AdminClient provides methods to interact with Envoy's admin API.
//
// This type alias exposes the internal AdminClient interface for experimental use.
type AdminClient = internalapi.AdminClient

// NewAdminClient returns an AdminClient if `funcEPid` has a child envoy process.
func NewAdminClient(ctx context.Context, funcEPid int) (AdminClient, error) {
	// Poll for the admin address path from the Envoy process command line
	_, adminAddressPath, err := admin.PollEnvoyPidAndAdminAddressPath(ctx, funcEPid)
	if err != nil {
		return nil, err
	}
	return admin.NewAdminClient(ctx, adminAddressPath)
}

// StartupHook runs once the Envoy admin server is ready. Configure this
// via the WithStartupHook api.RunOption.
//
// The hook receives the AdminClient and runID. The runID is unique to this run
// and can be used to construct file paths as needed.
//
// Note: Startup hooks are considered mandatory and will stop the run with
// error if failed. If your hook is optional, rescue panics and log your own
// errors.
type StartupHook = internalapi.StartupHook

// WithStartupHook returns a RunOption that sets a startup hook.
//
// This is an experimental API that should only be used by CLI entrypoints.
// See package documentation for usage constraints.
//
// If provided, this hook will REPLACE the default config dump hook.
// If you want to preserve default behavior, do not use this option.
func WithStartupHook(hook StartupHook) api.RunOption {
	return func(o *internalapi.RunOpts) {
		o.StartupHook = hook
	}
}
