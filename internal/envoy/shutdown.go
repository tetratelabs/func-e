// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tetratelabs/func-e/internal/tar"
)

// RegisterShutdownHook registers the passed functions to be run after Envoy has started
// and just before func-e instructs Envoy to exit
func (r *Runtime) RegisterShutdownHook(f func(context.Context) error) {
	r.shutdownHooks = append(r.shutdownHooks, f)
}

func (r *Runtime) handleShutdown() {
	// Ensure the SIGINT forwards to Envoy even if a shutdown hook panics
	defer func() {
		r.interruptEnvoy()
		if r.cmd != nil && r.cmd.Process != nil {
			_ = ensureProcessDone(r.cmd.Process)
		}
	}()

	deadline := time.Now().Add(shutdownTimeout)
	timeout, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	fmt.Fprintf(r.Out, "invoking shutdown hooks with deadline %s\n", deadline.Format(dateFormat)) //nolint:errcheck

	// Run each hook in parallel, logging each error
	var wg sync.WaitGroup
	wg.Add(len(r.shutdownHooks))
	for _, f := range r.shutdownHooks {
		go func(f func(context.Context) error) {
			defer wg.Done()
			if err := f(timeout); err != nil {
				fmt.Fprintf(r.Out, "failed shutdown hook: %s\n", err) //nolint:errcheck
			}
		}(f)
	}
	wg.Wait()
}

func (r *Runtime) interruptEnvoy() {
	p := r.cmd.Process
	r.logf("sending interrupt to envoy (pid=%d)", p.Pid)
	r.maybeWarn(interrupt(p))
}

func (r *Runtime) archiveRunDir() error {
	// Ensure logs are closed before we try to archive them.
	if r.OutFile != nil {
		r.OutFile.Close() //nolint
	}
	if r.ErrFile != nil {
		r.ErrFile.Close() //nolint
	}
	if r.o.DontArchiveRunDir {
		return nil
	}

	// Given ~/.func-e/debug/1620955405964267000
	dirName := filepath.Dir(r.GetRunDir())                  // ~/.func-e/runs
	baseName := filepath.Base(r.GetRunDir())                // 1620955405964267000
	targzName := filepath.Join(dirName, baseName+".tar.gz") // ~/.func-e/runs/1620955405964267000.tar.gz

	if err := tar.TarGz(targzName, r.GetRunDir()); err != nil {
		return fmt.Errorf("unable to archive run directory %v: %w", r.GetRunDir(), err)
	}
	return os.RemoveAll(r.GetRunDir())
}
