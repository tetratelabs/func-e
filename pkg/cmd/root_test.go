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
	"io"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"

	"github.com/mitchellh/go-homedir"

	. "github.com/tetratelabs/getenvoy/pkg/cmd"
	"github.com/tetratelabs/getenvoy/pkg/common"
	"github.com/tetratelabs/getenvoy/pkg/manifest"

	cmdutil "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

var _ = Describe("getenvoy", func() {

	var backupEnviron []string

	BeforeEach(func() {
		backupEnviron = os.Environ()
	})

	AfterEach(func() {
		for _, pair := range backupEnviron {
			parts := strings.SplitN(pair, "=", 2)
			key, value := parts[0], parts[1]
			os.Setenv(key, value)
		}
	})

	var stdout *bytes.Buffer
	var stderr *bytes.Buffer

	BeforeEach(func() {
		stdout = new(bytes.Buffer)
		stderr = new(bytes.Buffer)
	})

	newRootCmd := func(stdout, stderr io.Writer) *cobra.Command {
		c := NewRoot()
		c.SetOut(stdout)
		c.SetErr(stderr)

		// add a fake sub-command for unit test
		c.AddCommand(&cobra.Command{
			Use: "fake-command",
			RunE: func(_ *cobra.Command, _ []string) error {
				return nil
			},
		})
		return c
	}

	defaultHomeDir := func() string {
		home, err := homedir.Dir()
		Expect(err).NotTo(HaveOccurred())
		return filepath.Join(home, ".getenvoy")
	}

	It("should not have any required arguments", func() {
		By("running command")
		c := newRootCmd(stdout, stderr)
		c.SetArgs([]string{"fake-command"})
		err := cmdutil.Execute(c)
		Expect(err).NotTo(HaveOccurred())

		By("verifying command output")
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(BeEmpty())

		By("verifying global state")
		Expect(common.HomeDir).To(Equal(defaultHomeDir()))
		Expect(manifest.GetURL()).To(Equal(`https://tetrate.bintray.com/getenvoy/manifest.json`))
	})

	It("should support 'GETENVOY_HOME' environment variable", func() {
		expected := "/path/to/getenvoy/home" //nolint:goconst

		By("running command")
		os.Setenv("GETENVOY_HOME", expected)
		c := newRootCmd(stdout, stderr)
		c.SetArgs([]string{"fake-command"})
		err := cmdutil.Execute(c)
		Expect(err).NotTo(HaveOccurred())

		By("verifying global state")
		Expect(common.HomeDir).To(Equal(expected))
	})

	It("should support '--home-dir' command line option", func() {
		expected := "/path/to/getenvoy/home" //nolint:goconst

		By("running command")
		c := newRootCmd(stdout, stderr)
		c.SetArgs([]string{"--home-dir", expected, "fake-command"})
		err := cmdutil.Execute(c)
		Expect(err).NotTo(HaveOccurred())

		By("verifying global state")
		Expect(common.HomeDir).To(Equal(expected))
	})

	It("should prioritize '--home-dir' command line option over 'GETENVOY_HOME' environment variable", func() {
		unexpected := "/path/that/should/be/ignored"
		expected := "/path/to/getenvoy/home" //nolint:goconst

		By("running command")
		os.Setenv("GETENVOY_HOME", unexpected)
		c := newRootCmd(stdout, stderr)
		c.SetArgs([]string{"--home-dir", expected, "fake-command"})
		err := cmdutil.Execute(c)
		Expect(err).NotTo(HaveOccurred())

		By("verifying global state")
		Expect(common.HomeDir).To(Equal(expected))
	})

	It("should reject empty '--home-dir'", func() {
		unexpected := "/path/that/should/be/ignored"

		By("running command")
		os.Setenv("GETENVOY_HOME", unexpected)
		c := newRootCmd(stdout, stderr)
		c.SetArgs([]string{"--home-dir=", "fake-command"})
		err := cmdutil.Execute(c)
		Expect(err).To(HaveOccurred())

		By("verifying command output")
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(Equal(`Error: GetEnvoy home directory cannot be empty

Run 'getenvoy fake-command --help' for usage.
`))
	})

	It("should support 'GETENVOY_MANIFEST_URL' environment variable", func() {
		expected := "http://host/path/to/manifest"

		By("running command")
		os.Setenv("GETENVOY_MANIFEST_URL", expected)
		c := newRootCmd(stdout, stderr)
		c.SetArgs([]string{"fake-command"})
		err := cmdutil.Execute(c)
		Expect(err).NotTo(HaveOccurred())

		By("verifying global state")
		Expect(manifest.GetURL()).To(Equal(expected))
	})

	It("should support '--manifest' command line option", func() {
		expected := "http://host/path/to/manifest"

		By("running command")
		c := newRootCmd(stdout, stderr)
		c.SetArgs([]string{"--manifest", expected, "fake-command"})
		err := cmdutil.Execute(c)
		Expect(err).NotTo(HaveOccurred())

		By("verifying global state")
		Expect(manifest.GetURL()).To(Equal(expected))
	})

	It("should prioritize '--manifest' command line option over 'GETENVOY_MANIFEST_URL' environment variable", func() {
		unexpected := "https://host/path/that/should/be/ignored" //nolint:goconst
		expected := "https://host/path/to/manifest"

		By("running command")
		os.Setenv("GETENVOY_MANIFEST_URL", unexpected)
		c := newRootCmd(stdout, stderr)
		c.SetArgs([]string{"--manifest", expected, "fake-command"})
		err := cmdutil.Execute(c)
		Expect(err).NotTo(HaveOccurred())

		By("verifying global state")
		Expect(manifest.GetURL()).To(Equal(expected))
	})

	It("should reject empty '--manifest' command line option", func() {
		unexpected := "https://host/path/that/should/be/ignored" //nolint:goconst

		By("running command")
		os.Setenv("GETENVOY_MANIFEST_URL", unexpected)
		c := newRootCmd(stdout, stderr)
		c.SetArgs([]string{"--manifest=", "fake-command"})
		err := cmdutil.Execute(c)
		Expect(err).To(HaveOccurred())

		By("verifying command output")
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(Equal(`Error: GetEnvoy manifest URL cannot be empty

Run 'getenvoy fake-command --help' for usage.
`))
	})

	It("should reject invalid '--manifest' command line option", func() {
		unexpected := "https://host/path/that/should/be/ignored" //nolint:goconst
		invalid := "/not/a/url"

		By("running command")
		os.Setenv("GETENVOY_MANIFEST_URL", unexpected)
		c := newRootCmd(stdout, stderr)
		c.SetArgs([]string{"--manifest", invalid, "fake-command"})
		err := cmdutil.Execute(c)
		Expect(err).To(HaveOccurred())

		By("verifying command output")
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(Equal(`Error: "/not/a/url" is not a valid manifest URL

Run 'getenvoy fake-command --help' for usage.
`))
	})
})
