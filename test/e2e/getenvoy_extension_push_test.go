// Copyright 2021 Tetrate
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

package e2e_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/tetratelabs/getenvoy/pkg/common"
	utilenvoy "github.com/tetratelabs/getenvoy/test/e2e/util/envoy"

	"github.com/tetratelabs/getenvoy/pkg/extension/wasmimage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	e2e "github.com/tetratelabs/getenvoy/test/e2e/util"

	workspaces "github.com/tetratelabs/getenvoy/pkg/extension/workspace"
	toolchains "github.com/tetratelabs/getenvoy/pkg/extension/workspace/toolchain"
)

const (
	localRegistryWasmImageRef = "localhost:5000/getenvoy/sample:latest"
)

var _ = Describe("getenvoy extension push", func() {
	var debugDir string

	BeforeEach(func() {
		debugDir = filepath.Join(common.DefaultHomeDir(), "debug")
	})

	var backupDebugDir string

	BeforeEach(func() {
		_, err := ioutil.ReadDir(debugDir)
		if os.IsNotExist(err) {
			return
		}
		Expect(err).NotTo(HaveOccurred())

		By("backing up GetEnvoy debug dir")
		backupDir, err := ioutil.TempDir(filepath.Dir(debugDir), "debug")
		Expect(err).NotTo(HaveOccurred())
		err = os.RemoveAll(backupDir)
		Expect(err).NotTo(HaveOccurred())

		err = os.Rename(debugDir, backupDir)
		Expect(err).NotTo(HaveOccurred())
		backupDebugDir = backupDir
	})

	AfterEach(func() {
		if backupDebugDir == "" {
			return
		}
		By("restoring GetEnvoy debug dir from backup")
		err := os.RemoveAll(debugDir)
		Expect(err).NotTo(HaveOccurred())
		err = os.Rename(backupDebugDir, debugDir)
		Expect(err).NotTo(HaveOccurred())
	})

	type testCase e2e.CategoryLanguageTuple

	testCases := func() []TableEntry {
		testCases := make([]TableEntry, 0)
		for _, combination := range e2e.GetCategoryLanguageCombinations() {
			testCases = append(testCases, Entry(combination.String(), testCase(combination)))
		}
		return testCases
	}

	AtMostOnce := func(fn func()) func() {
		var once sync.Once
		return func() {
			once.Do(fn)
		}
	}

	// TODO(musaprg): write teardown process for local registries if it's needed

	const extensionName = "my.extension"

	const terminateTimeout = 2 * time.Minute

	DescribeTable("should push a *.wasm file",
		func(given testCase) {
			By("choosing the output directory")
			outputDir := filepath.Join(tempDir, "new")

			By("running `extension init` command")
			_, _, err := GetEnvoy("extension init").
				Arg(outputDir).
				Arg("--category").Arg(given.Category.String()).
				Arg("--language").Arg(given.Language.String()).
				Arg("--name").Arg(extensionName).
				Exec()
			Expect(err).NotTo(HaveOccurred())

			By("changing to the output directory")
			err = os.Chdir(outputDir)
			Expect(err).NotTo(HaveOccurred())

			By("running `extension build` command")
			stdout, stderr, err := GetEnvoy("extension build").
				Args(e2e.Env.GetBuiltinContainerOptions()...).
				Exec()
			Expect(err).NotTo(HaveOccurred())

			By("verifying stdout/stderr")
			// apparently, use of `-t` option in `docker run` causes stderr to be incorporated into stdout
			Expect(stdout).NotTo(BeEmpty())
			Expect(stderr).To(BeEmpty())

			By("verifying *.wasm file")
			workspace, err := workspaces.GetWorkspaceAt(outputDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(workspace).NotTo(BeNil())
			toolchain, err := toolchains.LoadToolchain(toolchains.Default, workspace)
			Expect(err).NotTo(HaveOccurred())
			Expect(toolchain).NotTo(BeNil())

			By("running `extension push` command")
			_, _, err = GetEnvoy("extension push").Arg(localRegistryWasmImageRef).Exec()
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).NotTo(BeEmpty())
			Expect(stderr).To(BeEmpty())

			By("pulling pushed wasm binary")
			puller, err := wasmimage.NewPuller(false, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(puller).NotTo(BeNil())
			dstPath := filepath.Join(outputDir, "pulled_extension.wasm")
			_, err = puller.Pull(localRegistryWasmImageRef, dstPath)
			Expect(err).NotTo(HaveOccurred())

			By("running `extension run` command with pulled image")
			_, se, cancel, errs := GetEnvoy("extension run").
				Args(e2e.Env.GetBuiltinContainerOptions()...).
				Arg("--extension-file").Arg(dstPath).
				Start()

			cancelCh := make(chan struct{})
			cancelGracefully := AtMostOnce(func() {
				close(cancelCh)

				Expect(cancel()).To(Succeed())
				select {
				case e := <-errs:
					Expect(e).NotTo(HaveOccurred())
				case <-time.After(terminateTimeout):
					Fail(fmt.Sprintf("getenvoy command didn't exit gracefully within %s", terminateTimeout))
				}
			})
			// make sure to stop Envoy if test fails
			defer cancelGracefully()

			// fail the test if `getenvoy extension run` exits with an error or unexpectedly
			go func() {
				select {
				case e := <-errs:
					Expect(e).NotTo(HaveOccurred(), "getenvoy command exited unexpectedly")
				case <-cancelCh:
				}
			}()

			stderrLines := e2e.StreamLines(se).Named("stderr")

			By("waiting for Envoy Admin address to get logged")
			adminAddressPattern := regexp.MustCompile(`admin address: ([^:]+:[0-9]+)`)
			line, err := stderrLines.FirstMatch(adminAddressPattern).Wait(10 * time.Minute) // give time to compile the extension
			Expect(err).NotTo(HaveOccurred())
			adminAddress := adminAddressPattern.FindStringSubmatch(line)[1]

			By("waiting for Envoy start-up to complete")
			stderrLines.FirstMatch(regexp.MustCompile(`starting main dispatch loop`)).Wait(1 * time.Minute)

			By("verifying Envoy is ready")
			envoyClient, err := utilenvoy.NewClient(adminAddress)
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() bool {
				ready, e := envoyClient.IsReady()
				return e == nil && ready
			}, "60s", "100ms").Should(BeTrue())

			By("verifying Wasm extensions have been created")
			Eventually(func() bool {
				stats, e := envoyClient.GetStats()
				if e != nil {
					return false
				}
				// at the moment, the only available Wasm metric is the number of Wasm VMs
				concurrency := stats.GetMetric("server.concurrency")
				activeWasmVms := stats.GetMetric("wasm.envoy.wasm.runtime.v8.active")
				return concurrency != nil && activeWasmVms != nil && activeWasmVms.Value == concurrency.Value+2
			}, "60s", "100ms").Should(BeTrue())

			By("signaling Envoy to stop")
			cancelGracefully()
		},
		testCases()...,
	)
})
