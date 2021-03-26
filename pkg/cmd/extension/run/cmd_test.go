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

package run_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/otiai10/copy"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/tetratelabs/getenvoy/pkg/cmd"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
	testcontext "github.com/tetratelabs/getenvoy/pkg/test/cmd/extension"
	manifesttest "github.com/tetratelabs/getenvoy/pkg/test/manifest"
	"github.com/tetratelabs/getenvoy/pkg/types"
	cmdutil "github.com/tetratelabs/getenvoy/pkg/util/cmd"
)

//nolint:lll
var _ = Describe("getenvoy extension run", func() {

	var cwdBackup string

	BeforeEach(func() {
		cwd, err := os.Getwd()
		Expect(err).ToNot(HaveOccurred())
		cwdBackup = cwd
	})

	AfterEach(func() {
		if cwdBackup != "" {
			Expect(os.Chdir(cwdBackup)).To(Succeed())
		}
	})

	var dockerDir string

	BeforeEach(func() {
		dir, err := filepath.Abs("../../../extension/workspace/toolchain/builtin/testdata/toolchain")
		Expect(err).ToNot(HaveOccurred())
		dockerDir = dir
	})

	var pathBackup string

	BeforeEach(func() {
		pathBackup = os.Getenv("PATH")
	})

	AfterEach(func() {
		os.Setenv("PATH", pathBackup)
	})

	BeforeEach(func() {
		// override PATH to overshadow `docker` executable during the test
		path := strings.Join([]string{dockerDir, pathBackup}, string(filepath.ListSeparator))
		os.Setenv("PATH", path)
	})

	var getenvoyHomeBackup string

	BeforeEach(func() {
		getenvoyHomeBackup = os.Getenv("GETENVOY_HOME")
	})

	AfterEach(func() {
		os.Setenv("GETENVOY_HOME", getenvoyHomeBackup)
	})

	var getenvoyHomeDir string

	BeforeEach(func() {
		tempDir, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		getenvoyHomeDir = tempDir

		// override GETENVOY_HOME directory during the test
		os.Setenv("GETENVOY_HOME", getenvoyHomeDir)
	})

	AfterEach(func() {
		if getenvoyHomeDir != "" {
			Expect(os.RemoveAll(getenvoyHomeDir)).To(Succeed())
		}
	})

	var envoySubstituteArchiveDir string

	BeforeEach(func() {
		envoySubstituteArchiveDir = filepath.Join(cwdBackup, "../../../extension/workspace/example/runtime/getenvoy/testdata/envoy")
	})

	var manifestURLBackup string

	BeforeEach(func() {
		manifestURLBackup = manifest.GetURL()
	})

	AfterEach(func() {
		Expect(manifest.SetURL(manifestURLBackup)).To(Succeed())
	})

	var manifestServer manifesttest.Server

	BeforeEach(func() {
		testManifest, err := manifesttest.NewSimpleManifest("standard:1.17.0", "wasm:1.15", "wasm:stable")
		Expect(err).NotTo(HaveOccurred())

		manifestServer = manifesttest.NewServer(&manifesttest.ServerOpts{
			Manifest: testManifest,
			GetArtifactDir: func(uri string) (string, error) {
				ref, e := types.ParseReference(uri)
				if e != nil {
					return "", e
				}
				if ref.Flavor == "wasm" {
					return envoySubstituteArchiveDir, nil
				}
				if ref.Flavor == "standard" {
					ver, e := semver.NewVersion(ref.Version)
					if e == nil && ver.Major() >= 1 && ver.Minor() >= 17 {
						return envoySubstituteArchiveDir, nil
					}
				}
				return "", errors.Errorf("unexpected version of Envoy %q", uri)
			},
			OnError: func(err error) {
				Expect(err).NotTo(HaveOccurred())
			},
		})

		// override location of the GetEnvoy manifest
		err = manifest.SetURL(manifestServer.GetManifestURL())
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if manifestServer != nil {
			manifestServer.Close()
		}
	})

	var platform string

	BeforeEach(func() {
		key, err := manifest.NewKey("standard:1.17.0")
		Expect(err).NotTo(HaveOccurred())
		platform = strings.ToLower(key.Platform)
	})

	// envoyCaptureDir represents a directory used by the Envoy substitute script
	// to store captured info.
	var envoyCaptureDir string

	BeforeEach(func() {
		dir, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		envoyCaptureDir = dir
	})

	AfterEach(func() {
		if envoyCaptureDir != "" {
			Expect(os.RemoveAll(envoyCaptureDir)).To(Succeed())
		}
	})

	BeforeEach(func() {
		// set environment variables to give `envoy` substitute script a hint
		// where to put captured info
		os.Setenv("TEST_ENVOY_CAPTURE_CMDLINE_FILE", filepath.Join(envoyCaptureDir, "cmdline"))
		os.Setenv("TEST_ENVOY_CAPTURE_CWD_FILE", filepath.Join(envoyCaptureDir, "cwd"))
		os.Setenv("TEST_ENVOY_CAPTURE_CWD_DIR", filepath.Join(envoyCaptureDir, "cwd.d"))
	})

	envoyCaptured := struct {
		cmdline        func() string
		cwd            func() string
		readFile       func(string) []byte
		readFileToJSON func(string) map[string]interface{}
	}{
		cmdline: func() string {
			data, err := ioutil.ReadFile(os.Getenv("TEST_ENVOY_CAPTURE_CMDLINE_FILE"))
			Expect(err).NotTo(HaveOccurred())
			return string(data)
		},
		cwd: func() string {
			data, err := ioutil.ReadFile(os.Getenv("TEST_ENVOY_CAPTURE_CWD_FILE"))
			Expect(err).NotTo(HaveOccurred())
			return strings.TrimSpace(string(data))
		},
		readFile: func(name string) []byte {
			data, err := ioutil.ReadFile(filepath.Join(os.Getenv("TEST_ENVOY_CAPTURE_CWD_DIR"), name))
			Expect(err).NotTo(HaveOccurred())
			return data
		},
		readFileToJSON: func(name string) map[string]interface{} {
			data, err := ioutil.ReadFile(filepath.Join(os.Getenv("TEST_ENVOY_CAPTURE_CWD_DIR"), name))
			Expect(err).NotTo(HaveOccurred())
			data, err = yaml.YAMLToJSON(data)
			Expect(err).NotTo(HaveOccurred())
			obj := make(map[string]interface{})
			err = json.Unmarshal(data, &obj)
			Expect(err).ToNot(HaveOccurred())
			return obj
		},
	}

	testcontext.SetDefaultUser() // UID:GID == 1001:1002

	var stdout *bytes.Buffer
	var stderr *bytes.Buffer

	BeforeEach(func() {
		stdout = new(bytes.Buffer)
		stderr = new(bytes.Buffer)
	})

	var c *cobra.Command

	BeforeEach(func() {
		c = cmd.NewRoot()
		c.SetOut(stdout)
		c.SetErr(stderr)
	})

	It("should validate value of --toolchain-container-image flag", func() {
		By("running command")
		c.SetArgs([]string{"extension", "run", "--toolchain-container-image", "?invalid value?"})
		err := cmdutil.Execute(c)
		Expect(err).To(HaveOccurred())

		By("verifying command output")
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(Equal(`Error: "?invalid value?" is not a valid image name: invalid reference format

Run 'getenvoy extension run --help' for usage.
`))
	})

	It("should validate value of --toolchain-container-options flag", func() {
		By("running command")
		c.SetArgs([]string{"extension", "run", "--toolchain-container-options", "imbalanced ' quotes"})
		err := cmdutil.Execute(c)
		Expect(err).To(HaveOccurred())

		By("verifying command output")
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(Equal(`Error: "imbalanced ' quotes" is not a valid command line string

Run 'getenvoy extension run --help' for usage.
`))
	})

	It("should validate value of --envoy-version flag", func() {
		By("running command")
		c.SetArgs([]string{"extension", "run", "--envoy-version", "???"})
		err := cmdutil.Execute(c)
		Expect(err).To(HaveOccurred())

		By("verifying command output")
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(Equal(`Error: Envoy version is not valid: "???" is not a valid GetEnvoy reference. Expected format: <flavor>:<version>[/<platform>]

Run 'getenvoy extension run --help' for usage.
`))
	})

	It("should not allow --envoy-version and --envoy-path flags at the same time", func() {
		By("running command")
		c.SetArgs([]string{"extension", "run", "--envoy-version", "standard:1.17.0", "--envoy-path", "envoy"})
		err := cmdutil.Execute(c)
		Expect(err).To(HaveOccurred())

		By("verifying command output")
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(Equal(`Error: only one of flags '--envoy-version' and '--envoy-path' can be used at a time

Run 'getenvoy extension run --help' for usage.
`))
	})

	It("should validate value of --envoy-path flag (path doesn't exist)", func() {
		By("creating a path for test")
		tempDir, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		defer func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		}()
		filePath := filepath.Join(tempDir, "non-existing-dir", "non-existing-file")

		By("running command")
		c.SetArgs([]string{"extension", "run", "--envoy-path", filePath})
		err = cmdutil.Execute(c)
		Expect(err).To(HaveOccurred())

		By("verifying command output")
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(Equal(fmt.Sprintf(`Error: unable to find custom Envoy binary at %[1]q: stat %[1]s: no such file or directory

Run 'getenvoy extension run --help' for usage.
`, filePath)))
	})

	It("should validate value of --envoy-path flag (path is a dir)", func() {
		By("creating a path for test")
		tempDir, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		defer func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		}()
		dirPath := tempDir

		By("running command")
		c.SetArgs([]string{"extension", "run", "--envoy-path", dirPath})
		err = cmdutil.Execute(c)
		Expect(err).To(HaveOccurred())

		By("verifying command output")
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(Equal(fmt.Sprintf(`Error: unable to find custom Envoy binary at %q: there is a directory at a given path instead of a regular file

Run 'getenvoy extension run --help' for usage.
`, dirPath)))
	})

	It("should validate value of --envoy-path flag (file is not executable)", func() {
		By("creating a path for test")
		tempDir, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		defer func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		}()

		By("creating a non-executable file")
		filePath := filepath.Join(tempDir, "envoy")
		err = ioutil.WriteFile(filePath, []byte(`#!/bin/sh`), 0600)
		Expect(err).NotTo(HaveOccurred())

		By("running command")
		c.SetArgs([]string{"extension", "run", "--envoy-path", filePath})
		err = cmdutil.Execute(c)
		Expect(err).To(HaveOccurred())

		By("verifying command output")
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(Equal(fmt.Sprintf(`Error: unable to find custom Envoy binary at %q: file is not executable

Run 'getenvoy extension run --help' for usage.
`, filePath)))
	})

	It("should validate value of --envoy-options flag", func() {
		By("running command")
		c.SetArgs([]string{"extension", "run", "--envoy-options", "imbalanced ' quotes"})
		err := cmdutil.Execute(c)
		Expect(err).To(HaveOccurred())

		By("verifying command output")
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(Equal(`Error: "imbalanced ' quotes" is not a valid command line string

Run 'getenvoy extension run --help' for usage.
`))
	})

	It("should validate value of --extension-file flag (path doesn't exist)", func() {
		By("creating a path for test")
		tempDir, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		defer func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		}()
		filePath := filepath.Join(tempDir, "non-existing-dir", "non-existing-file")

		By("running command")
		c.SetArgs([]string{"extension", "run", "--extension-file", filePath})
		err = cmdutil.Execute(c)
		Expect(err).To(HaveOccurred())

		By("verifying command output")
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(Equal(fmt.Sprintf(`Error: unable to find a pre-built *.wasm file at %[1]q: stat %[1]s: no such file or directory

Run 'getenvoy extension run --help' for usage.
`, filePath)))
	})

	It("should validate value of --extension-file flag (path is a dir)", func() {
		By("creating a path for test")
		tempDir, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		defer func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		}()
		dirPath := tempDir

		By("running command")
		c.SetArgs([]string{"extension", "run", "--extension-file", dirPath})
		err = cmdutil.Execute(c)
		Expect(err).To(HaveOccurred())

		By("verifying command output")
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(Equal(fmt.Sprintf(`Error: unable to find a pre-built *.wasm file at %q: there is a directory at a given path instead of a regular file

Run 'getenvoy extension run --help' for usage.
`, dirPath)))
	})

	It("should validate value of --extension-config-file flag (path doesn't exist)", func() {
		By("creating a path for test")
		tempDir, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		defer func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		}()
		filePath := filepath.Join(tempDir, "non-existing-dir", "non-existing-file")

		By("running command")
		c.SetArgs([]string{"extension", "run", "--extension-config-file", filePath})
		err = cmdutil.Execute(c)
		Expect(err).To(HaveOccurred())

		By("verifying command output")
		Expect(stdout.String()).To(BeEmpty())
		Expect(stderr.String()).To(Equal(fmt.Sprintf(`Error: failed to read custom extension config from file %[1]q: open %[1]s: no such file or directory

Run 'getenvoy extension run --help' for usage.
`, filePath)))
	})

	chdir := func(path string) string {
		dir, err := filepath.Abs(path)
		Expect(err).ToNot(HaveOccurred())

		dir, err = filepath.EvalSymlinks(dir)
		Expect(err).ToNot(HaveOccurred())

		err = os.Chdir(dir)
		Expect(err).ToNot(HaveOccurred())

		return dir
	}

	//nolint:lll
	Context("inside a workspace directory", func() {
		It("should succeed", func() {
			By("changing to a workspace dir")
			workspaceDir := chdir("testdata/workspace")

			By("running command")
			c.SetArgs([]string{"extension", "run"})
			err := cmdutil.Execute(c)
			Expect(err).ToNot(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(Equal(fmt.Sprintf(`%s/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init getenvoy/extension-rust-builder:latest build --output-file target/getenvoy/extension.wasm
%s/builds/standard/1.17.0/%s/bin/envoy -c %s/envoy.tmpl.yaml
`, dockerDir, workspaceDir, getenvoyHomeDir, platform, envoyCaptured.cwd())))
			Expect(stderr.String()).To(Equal("docker stderr\nenvoy stderr\n"))

			By("verifying Envoy config")
			placeholders := envoyCaptured.readFileToJSON("placeholders.tmpl.yaml")
			Expect(placeholders["extension.name"]).To(Equal(`mycompany.filters.http.custom_metrics`))
			Expect(placeholders["extension.code"]).To(Equal(map[string]interface{}{
				"local": map[string]interface{}{
					"filename": filepath.Join(workspaceDir, "target/getenvoy/extension.wasm"),
				},
			}))
			Expect(placeholders["extension.config"]).To(Equal(map[string]interface{}{
				"@type": "type.googleapis.com/google.protobuf.StringValue",
				"value": `{"key":"value"}`,
			}))
		})

		It("should allow to override build image and add Docker cli options", func() {
			By("changing to a workspace dir")
			workspaceDir := chdir("testdata/workspace")

			By("running command")
			c.SetArgs([]string{"extension", "run",
				"--toolchain-container-image", "build/image",
				"--toolchain-container-options", `-e 'VAR=VALUE' -v "/host:/container"`,
			})
			err := cmdutil.Execute(c)
			Expect(err).ToNot(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(Equal(fmt.Sprintf(`%s/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init -e VAR=VALUE -v /host:/container build/image build --output-file target/getenvoy/extension.wasm
%s/builds/standard/1.17.0/%s/bin/envoy -c %s/envoy.tmpl.yaml
`, dockerDir, workspaceDir, getenvoyHomeDir, platform, envoyCaptured.cwd())))
			Expect(stderr.String()).To(Equal("docker stderr\nenvoy stderr\n"))
		})

		It("should properly handle Docker build failing", func() {
			By("changing to a workspace dir")
			workspaceDir := chdir("testdata/workspace")

			By("running command")
			c.SetArgs([]string{"extension", "run",
				"--toolchain-container-image", "build/image",
				"--toolchain-container-options", `-e EXIT_CODE=3`,
			})
			err := cmdutil.Execute(c)
			Expect(err).To(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(Equal(fmt.Sprintf("%s/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init -e EXIT_CODE=3 build/image build --output-file target/getenvoy/extension.wasm\n", dockerDir, workspaceDir)))
			Expect(stderr.String()).To(Equal(fmt.Sprintf(`docker stderr
Error: failed to build Envoy extension using "default" toolchain: failed to execute an external command "%s/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init -e EXIT_CODE=3 build/image build --output-file target/getenvoy/extension.wasm": exit status 3

Run 'getenvoy extension run --help' for usage.
`, dockerDir, workspaceDir)))
		})

		It("should allow to override Envoy version via --envoy-version flag", func() {
			By("changing to a workspace dir")
			workspaceDir := chdir("testdata/workspace")

			By("running command")
			c.SetArgs([]string{"extension", "run", "--envoy-version", "wasm:stable"})
			err := cmdutil.Execute(c)
			Expect(err).NotTo(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(Equal(fmt.Sprintf(`%s/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init getenvoy/extension-rust-builder:latest build --output-file target/getenvoy/extension.wasm
%s/builds/wasm/stable/%s/bin/envoy -c %s/envoy.tmpl.yaml
`, dockerDir, workspaceDir, getenvoyHomeDir, platform, envoyCaptured.cwd())))
			Expect(stderr.String()).To(Equal("docker stderr\nenvoy stderr\n"))
		})

		It("should properly handle unknown Envoy version", func() {
			By("changing to a workspace dir")
			workspaceDir := chdir("testdata/workspace")

			By("running command")
			c.SetArgs([]string{"extension", "run", "--envoy-version", "wasm:unknown"})
			err := cmdutil.Execute(c)
			Expect(err).To(HaveOccurred())

			key, err := manifest.NewKey("wasm:unknown")
			Expect(err).NotTo(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(Equal(fmt.Sprintf("%s/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init getenvoy/extension-rust-builder:latest build --output-file target/getenvoy/extension.wasm\n", dockerDir, workspaceDir)))
			Expect(stderr.String()).To(Equal(fmt.Sprintf(`docker stderr
Error: failed to run "default" example: unable to find matching GetEnvoy build for reference %q

Run 'getenvoy extension run --help' for usage.
`, key)))
		})

		It("should allow to provide a custom Envoy binary via --envoy-path flag", func() {
			By("changing to a workspace dir")
			workspaceDir := chdir("testdata/workspace")

			By("running command")
			c.SetArgs([]string{"extension", "run", "--envoy-path", filepath.Join(envoySubstituteArchiveDir, "bin/envoy")})
			err := cmdutil.Execute(c)
			Expect(err).NotTo(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(Equal(fmt.Sprintf(`%s/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init getenvoy/extension-rust-builder:latest build --output-file target/getenvoy/extension.wasm
%s -c %s/envoy.tmpl.yaml
`, dockerDir, workspaceDir, filepath.Join(envoySubstituteArchiveDir, "bin/envoy"), envoyCaptured.cwd())))
			Expect(stderr.String()).To(Equal("docker stderr\nenvoy stderr\n"))
		})

		It("should allow to provide extra options for Envoy via --envoy-options flag", func() {
			By("changing to a workspace dir")
			workspaceDir := chdir("testdata/workspace")

			By("running command")
			c.SetArgs([]string{"extension", "run", "--envoy-options", "'--concurrency 2 --component-log-level wasm:debug,config:trace'"})
			err := cmdutil.Execute(c)
			Expect(err).NotTo(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(Equal(fmt.Sprintf(`%s/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init getenvoy/extension-rust-builder:latest build --output-file target/getenvoy/extension.wasm
%s/builds/standard/1.17.0/%s/bin/envoy -c %s/envoy.tmpl.yaml --concurrency 2 --component-log-level wasm:debug,config:trace
`, dockerDir, workspaceDir, getenvoyHomeDir, platform, envoyCaptured.cwd())))
			Expect(stderr.String()).To(Equal("docker stderr\nenvoy stderr\n"))
		})

		It("should allow to provide a pre-build *.wasm files via --extension-file flag", func() {
			By("changing to a workspace dir")
			_ = chdir("testdata/workspace")

			By("simulating a pre-built *.wasm file")
			tempDir, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
			defer func() {
				Expect(os.RemoveAll(tempDir)).To(Succeed())
			}()
			wasmFile := filepath.Join(tempDir, "extension.wasm")
			err = ioutil.WriteFile(wasmFile, []byte{}, 0600)
			Expect(err).NotTo(HaveOccurred())

			By("running command")
			c.SetArgs([]string{"extension", "run", "--extension-file", wasmFile})
			err = cmdutil.Execute(c)
			Expect(err).NotTo(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(Equal(fmt.Sprintf("%s/builds/standard/1.17.0/%s/bin/envoy -c %s/envoy.tmpl.yaml\n", getenvoyHomeDir, platform, envoyCaptured.cwd())))
			Expect(stderr.String()).To(Equal("envoy stderr\n"))

			By("verifying Envoy config")
			placeholders := envoyCaptured.readFileToJSON("placeholders.tmpl.yaml")
			Expect(placeholders["extension.code"]).To(Equal(map[string]interface{}{
				"local": map[string]interface{}{
					"filename": wasmFile,
				},
			}))
		})

		It("should allow to provide a custom extension config via --extension-config-file flag", func() {
			By("changing to a workspace dir")
			workspaceDir := chdir("testdata/workspace")

			By("simulating a custom extension config")
			tempDir, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
			defer func() {
				Expect(os.RemoveAll(tempDir)).To(Succeed())
			}()
			configFile := filepath.Join(tempDir, "config.json")
			err = ioutil.WriteFile(configFile, []byte(`{"key2":"value2"}`), 0600)
			Expect(err).NotTo(HaveOccurred())

			By("running command")
			c.SetArgs([]string{"extension", "run", "--extension-config-file", configFile})
			err = cmdutil.Execute(c)
			Expect(err).NotTo(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(Equal(fmt.Sprintf(`%s/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init getenvoy/extension-rust-builder:latest build --output-file target/getenvoy/extension.wasm
%s/builds/standard/1.17.0/%s/bin/envoy -c %s/envoy.tmpl.yaml
`, dockerDir, workspaceDir, getenvoyHomeDir, platform, envoyCaptured.cwd())))
			Expect(stderr.String()).To(Equal("docker stderr\nenvoy stderr\n"))

			By("verifying Envoy config")
			placeholders := envoyCaptured.readFileToJSON("placeholders.tmpl.yaml")
			Expect(placeholders["extension.config"]).To(Equal(map[string]interface{}{
				"@type": "type.googleapis.com/google.protobuf.StringValue",
				"value": `{"key2":"value2"}`,
			}))
		})

		It("should create default example if missing", func() {
			By("simulating a workspace without 'default' example")
			tempDir, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
			defer func() {
				Expect(os.RemoveAll(tempDir)).To(Succeed())
			}()
			err = copy.Copy("../build/testdata/workspace", tempDir)
			Expect(err).NotTo(HaveOccurred())

			By("changing to a workspace dir")
			workspaceDir := chdir(tempDir)

			By("running command")
			c.SetArgs([]string{"extension", "run"})
			err = cmdutil.Execute(c)
			Expect(err).ToNot(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(Equal(fmt.Sprintf(`%s/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init getenvoy/extension-rust-builder:latest build --output-file target/getenvoy/extension.wasm
%s/builds/standard/1.17.0/%s/bin/envoy -c %s/envoy.tmpl.yaml
`, dockerDir, workspaceDir, getenvoyHomeDir, platform, envoyCaptured.cwd())))
			Expect(stderr.String()).To(Equal(`Scaffolding a new example setup:
* .getenvoy/extension/examples/default/README.md
* .getenvoy/extension/examples/default/envoy.tmpl.yaml
* .getenvoy/extension/examples/default/example.yaml
* .getenvoy/extension/examples/default/extension.json
Done!
docker stderr
envoy stderr
`))

			By("verifying Envoy config")
			bootstrap := envoyCaptured.readFileToJSON("envoy.tmpl.yaml")
			Expect(bootstrap).NotTo(BeEmpty())
		})

		It("should create default example if missing for TinyGo", func() {
			By("simulating a workspace without 'default' example")
			tempDir, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
			defer func() {
				Expect(os.RemoveAll(tempDir)).To(Succeed())
			}()
			err = copy.Copy("testdata/workspace_tinygo", tempDir)
			Expect(err).NotTo(HaveOccurred())

			By("changing to a workspace dir")
			workspaceDir := chdir(tempDir)

			By("running command")
			c.SetArgs([]string{"extension", "run"})
			err = cmdutil.Execute(c)
			Expect(err).ToNot(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(Equal(fmt.Sprintf(`%s/docker run -u 1001:1002 --rm -t -v %s:/source -w /source --init getenvoy/extension-tinygo-builder:latest build --output-file build/extension.wasm
%s/builds/standard/1.17.0/%s/bin/envoy -c %s/envoy.tmpl.yaml
`, dockerDir, workspaceDir, getenvoyHomeDir, platform, envoyCaptured.cwd())))
			Expect(stderr.String()).To(Equal(`Scaffolding a new example setup:
* .getenvoy/extension/examples/default/README.md
* .getenvoy/extension/examples/default/envoy.tmpl.yaml
* .getenvoy/extension/examples/default/example.yaml
* .getenvoy/extension/examples/default/extension.txt
Done!
docker stderr
envoy stderr
`))

			By("verifying Envoy config")
			bootstrap := envoyCaptured.readFileToJSON("envoy.tmpl.yaml")
			Expect(bootstrap).NotTo(BeEmpty())
		})
	})

	Context("outside of a workspace directory", func() {
		It("should fail", func() {
			By("changing to a non-workspace dir")
			dir := chdir("testdata")

			By("running command")
			c.SetArgs([]string{"extension", "run"})
			err := cmdutil.Execute(c)
			Expect(err).To(HaveOccurred())

			By("verifying command output")
			Expect(stdout.String()).To(BeEmpty())
			Expect(stderr.String()).To(Equal(fmt.Sprintf(`Error: there is no extension directory at or above: %s

Run 'getenvoy extension run --help' for usage.
`, dir)))
		})
	})
})
