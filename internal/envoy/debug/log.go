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

package debug

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	envoy2 "github.com/tetratelabs/getenvoy/internal/envoy"
)

// enableEnvoyLogCollection is a preset option that registers collection of Envoy access logs and stderr
func enableEnvoyLogCollection(r *envoy2.Runtime) error {
	logsDir := filepath.Join(r.GetWorkingDir(), "logs")
	if err := os.MkdirAll(logsDir, 0750); err != nil {
		return fmt.Errorf("unable to create directory %q, so no logs will be captured: %w", logsDir, err)
	}
	e := envoyLogCollection{r, logsDir}
	r.RegisterPreStart(e.captureStdout)
	r.RegisterPreStart(e.captureStderr)
	return nil
}

type envoyLogCollection struct {
	r       *envoy2.Runtime
	logsDir string
}

func (e *envoyLogCollection) captureStdout() error {
	f, err := createLogFile(filepath.Join(e.logsDir, "access.log"))
	if err != nil {
		return err
	}
	e.r.RegisterPostTermination(f.Close)
	e.r.SetStdout(func(w io.Writer) io.Writer {
		if w == nil {
			return f
		}
		return io.MultiWriter(w, f)
	})
	return nil
}

func (e *envoyLogCollection) captureStderr() error {
	f, err := createLogFile(filepath.Join(e.logsDir, "error.log"))
	if err != nil {
		return err
	}
	e.r.RegisterPostTermination(f.Close)
	e.r.SetStderr(func(w io.Writer) io.Writer {
		if w == nil {
			return f
		}
		return io.MultiWriter(w, f)
	})
	return nil
}

func createLogFile(path string) (*os.File, error) {
	// #nosec -> logs can be written anywhere
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("unable to open file to write logs to %v: %w", path, err)
	}
	return f, nil
}
