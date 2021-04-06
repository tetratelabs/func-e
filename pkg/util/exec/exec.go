// Copyright 2020 Tetrate
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

package exec

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/tetratelabs/log"

	commonerrors "github.com/tetratelabs/getenvoy/pkg/errors"
	ioutil "github.com/tetratelabs/getenvoy/pkg/util/io"
	osutil "github.com/tetratelabs/getenvoy/pkg/util/os"
)

var (
	// killTimeout represents a delay between SIGTERM and SIGKILL signals
	// that are sent to an external command when getenvoy process itself
	// gets terminated.
	killTimeout = 60 * time.Second
)

var (
	setupSignalHandler = osutil.SetupSignalHandler
)

func newRunError(cmd fmt.Stringer, cause error) *RunError {
	return &RunError{
		cmd:   cmd.String(),
		cause: cause,
	}
}

// RunError represents an error to run an external command.
type RunError struct {
	cmd   string
	cause error
}

// Cmd returns a failed command.
func (e *RunError) Cmd() string {
	return e.cmd
}

// Cause returns a cause.
func (e *RunError) Cause() error {
	return e.cause
}

func (e *RunError) Error() string {
	return fmt.Sprintf("failed to execute an external command %q: %v", e.Cmd(), e.Cause())
}

// Run executes a given command.
func Run(cmd *exec.Cmd, streams ioutil.StdStreams) error {
	log.Debugf("running: %s", cmd)

	// configure standard I/O of the external process
	cmd.Stdin = streams.In
	cmd.Stdout = streams.Out
	cmd.Stderr = streams.Err

	// configure system attributes of the external process
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = new(syscall.SysProcAttr)
	}
	// give OS a hint to automatically kill a child process whenever its parent dies.
	// notice that when it comes to a `docker run` command, SIGTERM is a better fit than SIGKILL
	parentDeathAttr.Set(cmd.SysProcAttr, syscall.SIGTERM)

	// start the specified command in a separate step to ensure that
	// cmd.Process field is always set by the time we need it
	if err := cmd.Start(); err != nil {
		return newRunError(cmd, err)
	}

	// use a dedicated goroutine to wait for the command to complete
	errCh := make(chan error)
	go func() {
		defer close(errCh)
		if err := cmd.Wait(); err != nil {
			errCh <- err
		}
	}()

	// setup a cancelable handler for stop signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stopCh := setupSignalHandler(ctx)

	// wait for the external command to complete or the current process to get stopped
	select {
	case err := <-errCh:
		if err != nil {
			return newRunError(cmd, err)
		}
		return nil
	case sig := <-stopCh:
		terminate(cmd)
		return commonerrors.NewShutdownError(sig)
	}
}

func terminate(cmd *exec.Cmd) {
	if isProcessRunning(cmd.Process.Pid) {
		// first, give the external process a chance to exit gracefully,
		// which is a must in case of `docker run` command.
		if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
			log.Warnf("failed to send SIGTERM to the external process %q: %v", cmd, err)
		}
	}

	<-time.After(killTimeout)
	if isProcessRunning(cmd.Process.Pid) {
		log.Warnf("external process didn't exit gracefully within %s: %q", killTimeout, cmd)
		if e := cmd.Process.Kill(); e != nil {
			log.Warnf("failed to send SIGKILL to the external process %q: %v", cmd, e)
		}
	}
	_ = cmd.Process.Release()
}

// Same as cgi.isProcessRunning.
// See https://github.com/golang/go/issues/34396 for tracking long term solution.
func isProcessRunning(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}
