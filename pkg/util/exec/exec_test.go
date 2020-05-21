// +build !windows

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
	"bytes"
	"context"
	stderrors "errors"
	"os"
	"os/exec"
	"syscall"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	commonerrors "github.com/tetratelabs/getenvoy/pkg/errors"
	ioutil "github.com/tetratelabs/getenvoy/pkg/util/io"
)

var _ = Describe("Run()", func() {

	var stopCh chan os.Signal
	var setupSignalHandlerBackup func(ctx context.Context) <-chan os.Signal
	var killTimeoutBackup time.Duration

	BeforeEach(func() {
		setupSignalHandlerBackup = setupSignalHandler
		stopCh = make(chan os.Signal, 1)
		setupSignalHandler = func(ctx context.Context) <-chan os.Signal {
			return stopCh
		}
		killTimeoutBackup = killTimeout
		killTimeout = 3 * time.Second
	})

	AfterEach(func() {
		killTimeout = killTimeoutBackup
		setupSignalHandler = setupSignalHandlerBackup
	})

	var stdin *bytes.Buffer
	var stdout *bytes.Buffer
	var stderr *bytes.Buffer
	var stdio ioutil.StdStreams

	BeforeEach(func() {
		stdin = new(bytes.Buffer)
		stdout = new(bytes.Buffer)
		stderr = new(bytes.Buffer)
		stdio = ioutil.StdStreams{
			In:  stdin,
			Out: stdout,
			Err: stderr,
		}
	})

	It("should properly pipe standard I/O", func() {
		cmd := exec.Command("testdata/test_stdio.sh", "0", "stderr")
		stdin.WriteString("stdin\n")

		err := Run(cmd, stdio)

		Expect(err).ToNot(HaveOccurred())
		Expect(stdout.String()).To(Equal(`stdin`))
		Expect(stderr.String()).To(Equal(`stderr`))
	})

	It("should return a meaningful error when a command cannot start", func() {
		cmd := exec.Command("testdata/test_stdio.sh", "123")
		cmd.Dir = "testdata"

		err := Run(cmd, stdio)

		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(`failed to execute an external command "testdata/test_stdio.sh 123": ` +
			`fork/exec testdata/test_stdio.sh: no such file or directory`))

		var runErr *RunError
		Expect(stderrors.As(err, &runErr)).To(BeTrue())
		Expect(runErr.Cmd()).To(Equal("testdata/test_stdio.sh 123"))

		var pathErr *os.PathError
		Expect(stderrors.As(runErr.Cause(), &pathErr)).To(BeTrue())
	})

	It("should return a meaningful error when a command exits with a non-0 code", func() {
		cmd := exec.Command("testdata/test_stdio.sh", "123")

		err := Run(cmd, stdio)

		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(`failed to execute an external command "testdata/test_stdio.sh 123": exit status 123`))

		var runErr *RunError
		Expect(stderrors.As(err, &runErr)).To(BeTrue())
		Expect(runErr.Cmd()).To(Equal("testdata/test_stdio.sh 123"))

		var exitErr *exec.ExitError
		Expect(stderrors.As(runErr.Cause(), &exitErr)).To(BeTrue())
	})

	It("on shutdown, should gracefully terminate (SIGTERM) a command", func() {
		cmd := exec.Command("testdata/test_sigterm.sh", "10")

		errCh := make(chan error)
		go func() {
			defer close(errCh)
			if err := Run(cmd, stdio); err != nil {
				errCh <- err
			}
		}()

		stopCh <- syscall.SIGINT
		close(stopCh)

		var err error
		Eventually(errCh, "5s", "100ms").Should(Receive(&err))
		Expect(err).To(Equal(commonerrors.NewShutdownError(syscall.SIGINT)))
	})

	It("on shutdown, should forcefully kill (SIGKILL) a command if the latter didn't exit timely after receiving SIGTERM", func() {
		cmd := exec.Command("testdata/test_sigkill.sh", "10")

		errCh := make(chan error)
		go func() {
			defer close(errCh)
			if err := Run(cmd, stdio); err != nil {
				errCh <- err
			}
		}()

		Eventually(stdout.String).Should(Equal("running"))

		stopCh <- syscall.SIGINT
		close(stopCh)

		var err error
		Eventually(errCh, "5s", "100ms").Should(Receive(&err))
		Expect(err).To(Equal(commonerrors.NewShutdownError(syscall.SIGINT)))
	})
})
