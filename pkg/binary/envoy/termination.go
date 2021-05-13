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
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/tetratelabs/getenvoy/internal/tar"
)

func (r *Runtime) handleTermination() {
	defer r.interruptEnvoy() // Ensure the SIGINT forwards to Envoy even if a pre-termination hook panics

	fmt.Fprintln(r.Out, "invoking pre-termination hooks") //nolint
	// Execute all registered preTermination functions
	for _, f := range r.preTermination {
		if err := f(); err != nil {
			fmt.Fprintln(r.Out, "failed pre-termination hook:", err) //nolint
		}
	}
}

func (r *Runtime) interruptEnvoy() {
	p := r.cmd.Process
	fmt.Fprintln(r.Out, "stopping envoy") //nolint
	_ = p.Signal(syscall.SIGINT)
}

func (r *Runtime) handlePostTermination() error {
	for _, f := range r.postTermination {
		if err := f(); err != nil {
			fmt.Fprintln(r.Out, "failed post-termination hook:", err) //nolint
		}
	}

	if r.opts.DontArchiveWorkingDir {
		return nil
	}

	// Given ~/.getenvoy/debug/1620955405964267000
	dirName := filepath.Dir(r.GetWorkingDir())              // ~/.getenvoy/debug
	baseName := filepath.Base(r.GetWorkingDir())            // 1620955405964267000
	targzName := filepath.Join(dirName, baseName+".tar.gz") // ~/.getenvoy/debug/1620955405964267000.tar.gz

	targz, err := os.Create(targzName)
	if err != nil {
		return err
	}
	defer targz.Close() //nolint
	zw := gzip.NewWriter(targz)
	defer zw.Close() //nolint

	if err = tar.Tar(zw, os.DirFS(dirName), baseName); err != nil {
		return fmt.Errorf("unable to archive run directory %v: %w", r.GetWorkingDir(), err)
	}
	return os.RemoveAll(r.GetWorkingDir())
}

// RegisterPreTermination registers the passed functions to be run after Envoy has started
// and just before GetEnvoy instructs Envoy to terminate
func (r *Runtime) RegisterPreTermination(f ...func() error) {
	r.preTermination = append(r.preTermination, f...)
}

// RegisterPostTermination registers the passed functions to be run after Envoy has terminated
// and just before GetEnvoy archives the run directory.
func (r *Runtime) RegisterPostTermination(f ...func() error) {
	r.postTermination = append(r.postTermination, f...)
}
