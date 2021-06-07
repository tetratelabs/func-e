// Copyright 2019 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package envoy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/tetratelabs/getenvoy/internal/tar"
)

// RegisterShutdownHook registers the passed functions to be run after Envoy has started
// and just before GetEnvoy instructs Envoy to exit
func (r *Runtime) RegisterShutdownHook(f ...func(context.Context) error) {
	r.shutdownHooks = append(r.shutdownHooks, f...)
}

func (r *Runtime) handleShutdown(ctx context.Context) {
	defer r.interruptEnvoy() // Ensure the SIGINT forwards to Envoy even if a shutdown hook panics

	deadline := time.Now().Add(shutdownTimeout)
	timeout, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	fmt.Fprintf(r.Out, "invoking shutdown hooks with deadline %s\n", deadline.Format(dateFormat)) //nolint

	// Run each hook, logging each error
	for _, f := range r.shutdownHooks {
		f := f // pin! see https://github.com/kyoh86/scopelint for why
		if err := f(timeout); err != nil {
			fmt.Fprintln(r.Out, "failed shutdown hook:", err) //nolint
		}
	}
}

func (r *Runtime) interruptEnvoy() {
	p := r.cmd.Process
	fmt.Fprintf(r.Out, "sending interrupt to envoy (pid=%d)\n", p.Pid) //nolint
	_ = p.Signal(syscall.SIGINT)
}

func (r *Runtime) archiveRunDir() error {
	if r.opts.DontArchiveRunDir {
		return nil
	}

	// Given ~/.getenvoy/debug/1620955405964267000
	dirName := filepath.Dir(r.GetRunDir())                  // ~/.getenvoy/runs
	baseName := filepath.Base(r.GetRunDir())                // 1620955405964267000
	targzName := filepath.Join(dirName, baseName+".tar.gz") // ~/.getenvoy/runs/1620955405964267000.tar.gz

	if err := tar.TarGz(targzName, r.GetRunDir()); err != nil {
		return fmt.Errorf("unable to archive run directory %v: %w", r.GetRunDir(), err)
	}
	return os.RemoveAll(r.GetRunDir())
}
