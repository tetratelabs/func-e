// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package run

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/tetratelabs/func-e/api"
	internalapi "github.com/tetratelabs/func-e/internal/api"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/test/e2e"
)

// fakeFuncEFactory implements runtest.FuncEFactory for API tests using fake envoy
type fakeFuncEFactory struct{}

func (fakeFuncEFactory) New(ctx context.Context, t *testing.T, stdout, stderr io.Writer) (e2e.FuncE, error) {
	o, err := initOpts(ctx, api.HomeDir(t.TempDir()),
		EnvoyPath(fakeEnvoyBin),
		api.Out(stdout),
		api.EnvoyOut(stdout),
		api.EnvoyErr(stderr))
	if err != nil {
		return nil, err
	}
	return &fakeFuncE{o: o}, nil
}

// fakeFuncE implements runtest.FuncE for API tests using fake envoy
type fakeFuncE struct {
	o          *globals.GlobalOpts
	cancelFunc context.CancelFunc
}

// Interrupt cancels the context created in Run as we don't want to actually interrupt the calling test!
func (f *fakeFuncE) Interrupt(context.Context) error {
	if f.cancelFunc != nil {
		f.cancelFunc()
		// Don't set to nil in case interrupt is called multiple times (ctrl+c twice)
	}
	return nil
}

// OnStart uses the cached runDir to read the envoy PID from the file created by envoy/run.go
func (f *fakeFuncE) OnStart(context.Context) (runDir string, envoyPid int32, err error) {
	envoyPidFile := filepath.Join(f.o.RunDir, "envoy.pid")
	pidBytes, err := os.ReadFile(envoyPidFile)
	if err != nil {
		return f.o.RunDir, 0, err
	}

	pid, err := strconv.Atoi(string(pidBytes))
	if err != nil {
		return f.o.RunDir, 0, fmt.Errorf("failed to parse Envoy PID from %s: %w", envoyPidFile, err)
	}

	return f.o.RunDir, int32(pid), nil
}

// Run invokes the underlying api.Run function, which has been configured to use a fake Envoy binary.
func (f *fakeFuncE) Run(ctx context.Context, args []string) error {
	// Since we aren't launching a real process, we proxy interrupt with context cancellation.
	ctx, cancel := context.WithCancel(ctx)
	f.cancelFunc = cancel
	return internalapi.Run(ctx, f.o, args)
}
