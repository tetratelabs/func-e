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

package cmd_test

import (
	"bytes"
	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	commonerrors "github.com/tetratelabs/getenvoy/pkg/errors"
	. "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

var _ = Describe("Execute()", func() {

	var stdout *bytes.Buffer
	var stderr *bytes.Buffer

	BeforeEach(func() {
		stdout = new(bytes.Buffer)
		stderr = new(bytes.Buffer)
	})

	Describe("should properly format errors", func() {
		It(`should properly format "unknown command" error`, func() {
			rootCmd := &cobra.Command{
				Use: "getenvoy",
			}
			rootCmd.AddCommand(&cobra.Command{
				Use: "init",
				RunE: func(_ *cobra.Command, _ []string) error {
					return errors.New("unexpected error")
				},
			})
			rootCmd.SetOut(stdout)
			rootCmd.SetErr(stderr)
			rootCmd.SetArgs([]string{"other", "command"})

			err := Execute(rootCmd)
			Expect(err).To(HaveOccurred())

			Expect(stdout.String()).To(BeEmpty())
			Expect(stderr.String()).To(Equal(`Error: unknown command "other" for "getenvoy"

Run 'getenvoy --help' for usage.
`))
		})

		It(`should properly format "unknown flag" error`, func() {
			rootCmd := &cobra.Command{
				Use: "getenvoy",
				RunE: func(_ *cobra.Command, _ []string) error {
					return errors.New("unexpected error")
				},
			}
			rootCmd.SetOut(stdout)
			rootCmd.SetErr(stderr)
			rootCmd.SetArgs([]string{"--xyz"})

			err := Execute(rootCmd)
			Expect(err).To(HaveOccurred())

			Expect(stdout.String()).To(BeEmpty())
			Expect(stderr.String()).To(Equal(`Error: unknown flag: --xyz

Run 'getenvoy --help' for usage.
`))
		})

		Context("application-specific errors", func() {
			type testCase struct {
				err         error
				expectedOut string
			}
			DescribeTable("should properly format application-specific errors",
				func(given testCase) {
					rootCmd := &cobra.Command{
						Use: "getenvoy",
						RunE: func(_ *cobra.Command, _ []string) error {
							return given.err
						},
					}
					rootCmd.SetOut(stdout)
					rootCmd.SetErr(stderr)
					rootCmd.SetArgs([]string{})

					err := Execute(rootCmd)
					Expect(err).To(Equal(given.err))

					Expect(stdout.String()).To(BeEmpty())
					Expect(stderr.String()).To(Equal(given.expectedOut))
				},
				Entry("arbitrary error", testCase{
					err: errors.New("expected error"),
					expectedOut: `Error: expected error

Run 'getenvoy --help' for usage.
`,
				}),
				Entry("shutdown error", testCase{
					err: commonerrors.NewShutdownError(syscall.SIGINT),
					expectedOut: `NOTE: Shutting down early because a Ctrl-C ("interrupt") was received.
`,
				}),
				Entry("wrapped shutdown error", testCase{
					err: errors.Wrap(commonerrors.NewShutdownError(syscall.SIGINT), "wrapped"),
					expectedOut: `NOTE: Shutting down early because a Ctrl-C ("interrupt") was received.
`,
				}),
			)
		})
	})

	It(`should support commands that run without an error`, func() {
		rootCmd := &cobra.Command{
			Use: "getenvoy",
			Run: func(_ *cobra.Command, _ []string) {},
		}
		rootCmd.SetOut(stdout)
		rootCmd.SetErr(stderr)
		rootCmd.SetArgs([]string{})

		err := Execute(rootCmd)

		Expect(err).ToNot(HaveOccurred())
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(BeEmpty())
	})
})
