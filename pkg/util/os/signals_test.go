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

package os

import (
	"context"
	"os"
	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetupSignalHandler()", func() {
	It("should register for SIGTERM and SIGINT", func() {
		Expect(shutdownSignals).To(ConsistOf(syscall.SIGINT, syscall.SIGTERM))
	})

	Context("with a given set of shutdown signals", func() {
		var relevantSignal = syscall.SIGUSR1
		var irrelevantSignal = syscall.SIGUSR2

		var shutdownSignalsBackup []os.Signal

		BeforeEach(func() {
			shutdownSignalsBackup = shutdownSignals
			shutdownSignals = []os.Signal{relevantSignal}
		})

		AfterEach(func() {
			shutdownSignals = shutdownSignalsBackup
		})

		It("should not close the returned channel on irrelevant signals", func() {
			By("doing set up")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			stopCh := SetupSignalHandler(ctx)

			By("sending an irrelevant signal to the process")
			Expect(syscall.Kill(syscall.Getpid(), irrelevantSignal)).To(Succeed())
			Consistently(stopCh).ShouldNot(Receive())
			Expect(stopCh).NotTo(BeClosed())
		})

		It("should close the returned channel on the first relevant signal", func() {
			By("doing set up")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			stopCh := SetupSignalHandler(ctx)

			By("sending a relevant signal to the process")
			Expect(syscall.Kill(syscall.Getpid(), relevantSignal)).To(Succeed())
			Eventually(stopCh).Should(Receive(Equal(relevantSignal)))
			Eventually(stopCh).Should(BeClosed())
		})

		Context("with a stub instead of terminate logic", func() {

			var terminateBackup func()
			var terminateCh chan struct{}

			BeforeEach(func() {
				terminateBackup = terminate
				terminateCh = make(chan struct{})
				terminate = func() {
					close(terminateCh)
				}
			})

			AfterEach(func() {
				terminate = terminateBackup
			})

			It("should terminate the process on the second relevant signal", func() {
				By("doing set up")
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				stopCh := SetupSignalHandler(ctx)

				By("sending first relevant signal to the process")
				Expect(syscall.Kill(syscall.Getpid(), relevantSignal)).To(Succeed())
				Eventually(stopCh).Should(Receive(Equal(relevantSignal)))
				Eventually(stopCh).Should(BeClosed())
				Consistently(terminateCh).ShouldNot(BeClosed())

				By("sending second relevant signal to the process")
				Expect(syscall.Kill(syscall.Getpid(), relevantSignal)).To(Succeed())
				Expect(stopCh).To(BeClosed())
				Eventually(terminateCh).Should(BeClosed())
			})

			It("should not terminate the process if the context becomes done before the first signal", func() {
				By("doing set up")
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				stopCh := SetupSignalHandler(ctx)

				By("canceling the context before the first signal")
				cancel()
				// give SetupSignalHandler some time to notice that the context is done
				Consistently(stopCh).ShouldNot(BeClosed())

				By("sending first relevant signal to the process")
				Expect(syscall.Kill(syscall.Getpid(), relevantSignal)).To(Succeed())
				Consistently(stopCh).ShouldNot(Receive())
				Expect(stopCh).NotTo(BeClosed())
				Consistently(terminateCh).ShouldNot(BeClosed())

				By("sending second relevant signal to the process")
				Expect(syscall.Kill(syscall.Getpid(), relevantSignal)).To(Succeed())
				Consistently(stopCh).ShouldNot(Receive())
				Expect(stopCh).NotTo(BeClosed())
				Consistently(terminateCh).ShouldNot(BeClosed())
			})

			It("should not terminate the process if the context becomes done before the second signal", func() {
				By("doing set up")
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				stopCh := SetupSignalHandler(ctx)

				By("sending first relevant signal to the process")
				Expect(syscall.Kill(syscall.Getpid(), relevantSignal)).To(Succeed())
				Eventually(stopCh).Should(Receive(Equal(relevantSignal)))
				Eventually(stopCh).Should(BeClosed())
				Consistently(terminateCh).ShouldNot(BeClosed())

				By("canceling the context before the second signal")
				cancel()
				// give SetupSignalHandler some time to notice that the context is done
				Consistently(terminateCh).ShouldNot(BeClosed())

				By("sending second relevant signal to the process")
				Expect(syscall.Kill(syscall.Getpid(), relevantSignal)).To(Succeed())
				Expect(stopCh).To(BeClosed())
				Consistently(terminateCh).ShouldNot(BeClosed())
			})
		})
	})
})
