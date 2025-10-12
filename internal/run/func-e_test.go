// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package run

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/tetratelabs/func-e/api"
	internaladmin "github.com/tetratelabs/func-e/internal/admin"
	internalapi "github.com/tetratelabs/func-e/internal/api"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/runtime"
	"github.com/tetratelabs/func-e/internal/test/e2e"
)

// fakeFuncEFactory implements e2e.FuncEFactory for API tests using fake envoy
type fakeFuncEFactory struct{}

func (f fakeFuncEFactory) New(ctx context.Context, t *testing.T, stdout, stderr io.Writer) (e2e.FuncE, error) {
	var opts []api.RunOption

	// Read from environment variables to support both legacy and separate directory modes
	// This mirrors CLI behavior where env vars control directory structure
	homeDir := os.Getenv("FUNC_E_HOME")
	dataHome := os.Getenv("FUNC_E_DATA_HOME")
	stateHome := os.Getenv("FUNC_E_STATE_HOME")
	runtimeDir := os.Getenv("FUNC_E_RUNTIME_DIR")

	switch {
	case homeDir != "":
		// Legacy mode via FUNC_E_HOME
		opts = []api.RunOption{api.HomeDir(homeDir)} //nolint:staticcheck // intentional use of deprecated API for legacy mode testing
	case dataHome != "" || stateHome != "" || runtimeDir != "":
		// Separate directories mode - apply only what's set via env vars
		// Helper function to get directory or create temp dir
		getDir := func(envValue string) string {
			if envValue != "" {
				return envValue
			}
			return t.TempDir()
		}

		opts = append(opts,
			api.DataHome(getDir(dataHome)),
			api.StateHome(getDir(stateHome)),
			api.RuntimeDir(getDir(runtimeDir)))
	default:
		// Default: use separate temp directories
		opts = []api.RunOption{
			api.DataHome(t.TempDir()),
			api.StateHome(t.TempDir()),
			api.RuntimeDir(t.TempDir()),
		}
	}

	opts = append(opts,
		EnvoyPath(fakeEnvoyBin),
		api.Out(stdout),
		api.EnvoyOut(stdout),
		api.EnvoyErr(stderr))

	o, err := initOpts(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &fakeFuncE{o: o}, nil
}

// fakeFuncE implements e2e.FuncE for API tests using fake envoy
type fakeFuncE struct {
	o          *globals.GlobalOpts
	cancelFunc context.CancelFunc
	envoyPid   int
}

// EnvoyPid implements the same method as documented on e2e.FuncE
func (f *fakeFuncE) EnvoyPid() int {
	return f.envoyPid
}

// Interrupt cancels the context created in Run as we don't want to actually interrupt the calling test!
func (f *fakeFuncE) Interrupt(context.Context) error {
	if f.cancelFunc != nil {
		f.cancelFunc()
		// Don't set to nil in case interrupt is called multiple times (ctrl+c twice)
	}
	return nil
}

// OnStart creates an admin client using the run directory and waits for Envoy to be ready.
func (f *fakeFuncE) OnStart(ctx context.Context) (internalapi.AdminClient, error) {
	// Poll for the admin address path from the Envoy process command line
	envoyPid, adminAddressPath, err := internaladmin.PollEnvoyPidAndAdminAddressPath(ctx, os.Getpid())
	if err != nil {
		return nil, err
	}
	f.envoyPid = envoyPid
	adminClient, err := internaladmin.NewAdminClient(ctx, adminAddressPath)
	if err == nil {
		err = adminClient.AwaitReady(ctx, 100*time.Millisecond)
	}
	return adminClient, err
}

// Run invokes the underlying api.Run function, which has been configured to use a fake Envoy binary.
func (f *fakeFuncE) Run(ctx context.Context, args []string) error {
	// Since we aren't launching a real process, we proxy interrupt with context cancellation.
	ctx, cancel := context.WithCancel(ctx)
	f.cancelFunc = cancel
	return runtime.Run(ctx, f.o, args)
}
