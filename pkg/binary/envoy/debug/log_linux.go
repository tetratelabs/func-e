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

	"github.com/tetratelabs/getenvoy/pkg/binary"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/log"
)

// EnableEnvoyLogCollection is a preset option that registers collection of Envoy access logs and stderr
func EnableEnvoyLogCollection(r *envoy.Runtime) {
	if err := os.MkdirAll(filepath.Join(r.DebugStore(), "logs"), os.ModePerm); err != nil {
		log.Errorf("unable to create directory to write logs to, no logs will be captured: %v", err)
		return
	}
	r.RegisterPreStart(captureStdout)
	r.RegisterPreStart(captureStderr)
}

func captureStdout(r binary.Runner) error {
	f, err := createLogFile(filepath.Join(r.DebugStore(), "logs", "access.log"))
	if err != nil {
		return err
	}
	r.SetStdout(io.MultiWriter(os.Stdout, f))
	go capture(r, f)
	return nil
}

func captureStderr(r binary.Runner) error {
	f, err := createLogFile(filepath.Join(r.DebugStore(), "logs", "error.log"))
	if err != nil {
		return err
	}
	r.SetStderr(io.MultiWriter(os.Stderr, f))
	go capture(r, f)
	return nil
}

func createLogFile(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("unable to open file to write logs to %v: %v", path, err)
	}
	return f, nil
}

func capture(r binary.Runner, file io.Closer) {
	r.RegisterWait(1)
	r.Wait(binary.StatusTerminated)
	if err := file.Close(); err != nil {
		log.Errorf("error closing access log file: %v", err)
	}
	r.RegisterDone()
}
