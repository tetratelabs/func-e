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

package e2e_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	e2e "github.com/tetratelabs/getenvoy/test/e2e/util"
)

func TestEndToEnd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "e2e Suite")
}

var _ = BeforeSuite(func() {
	path, err := e2e.Env.GetEnvoyBinary()
	Expect(err).NotTo(HaveOccurred())
	e2e.GetEnvoyBinaryPath = path
})

var (
	// GetEnvoy is a convenient alias.
	GetEnvoy = e2e.GetEnvoy
)

// tempDir represents a unique temporary directory made available for every test case.
var tempDir string

var _ = BeforeEach(func() {
	dir, err := ioutil.TempDir("", "")
	Expect(err).NotTo(HaveOccurred())
	dir, err = filepath.EvalSymlinks(dir)
	Expect(err).NotTo(HaveOccurred())
	tempDir = dir
})

var _ = AfterEach(func() {
	if tempDir != "" {
		Expect(os.RemoveAll(tempDir)).To(Succeed())
	}
})
