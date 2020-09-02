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

package init

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Params", func() {
	Describe("OutputDir", func() {
		It("should reject output path that is a file", func() {
			By("creating a file")
			tmpFile, err := ioutil.TempFile("", "file")
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				Expect(tmpFile.Close()).To(Succeed())
				Expect(os.Remove(tmpFile.Name())).To(Succeed())
			}()

			By("verifying file path")
			err = newParams().OutputDir.Validator(tmpFile.Name())
			Expect(err).To(MatchError(fmt.Sprintf(`output path is not a directory: %s`, tmpFile.Name())))
		})

		It("should reject output path that is under a file", func() {
			By("creating a file")
			tmpFile, err := ioutil.TempFile("", "file")
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				Expect(tmpFile.Close()).To(Succeed())
				Expect(os.Remove(tmpFile.Name())).To(Succeed())
			}()

			By("verifying path under a file")
			invalidPath := filepath.Join(tmpFile.Name(), "new_dir")
			err = newParams().OutputDir.Validator(invalidPath)
			Expect(err).To(MatchError(fmt.Sprintf(`stat %s: not a directory`, invalidPath)))
		})

		It("should reject output path that is a non-empty existing dir", func() {
			By("creating a dir")
			tmpDir, err := ioutil.TempDir("", "dir")
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				Expect(os.Remove(tmpDir)).To(Succeed())
			}()
			By("creating another dir inside")
			innerDir, err := ioutil.TempDir(tmpDir, "dir")
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				Expect(os.Remove(innerDir)).To(Succeed())
			}()

			By("verifying non-empty existing dir")
			err = newParams().OutputDir.Validator(tmpDir)
			Expect(err).To(MatchError(fmt.Sprintf(`output directory must be empty or new: %s`, tmpDir)))
		})

		It("should accept output path that is a non-existing dir", func() {
			By("creating a dir")
			tmpDir, err := ioutil.TempDir("", "dir")
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				Expect(os.Remove(tmpDir)).To(Succeed())
			}()

			By("verifying non-existing dir")
			validPath := filepath.Join(tmpDir, "child_dir", "grand_child_dir")
			err = newParams().OutputDir.Validator(validPath)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should accept output path that is an empty existing dir", func() {
			By("creating a dir")
			tmpDir, err := ioutil.TempDir("", "dir")
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				Expect(os.Remove(tmpDir)).To(Succeed())
			}()

			By("verifying an empty existing dir")
			err = newParams().OutputDir.Validator(tmpDir)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
