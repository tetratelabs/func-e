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

package builtin

import (
	"fmt"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config"
	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
)

var _ = Describe("defaultConfigFor()", func() {
	type testCase struct {
		extension *extension.Descriptor
		expected  string
	}
	DescribeTable("should generate proper default config for every supported programming language",
		func(given testCase) {
			cfg := defaultConfigFor(given.extension)
			Expect(cfg.Validate()).To(Succeed())

			actual, err := config.Marshal(cfg)
			Expect(err).ToNot(HaveOccurred())

			Expect(actual).To(MatchYAML(given.expected))
		},
		func() []TableEntry {
			entries := make([]TableEntry, len(extension.Languages))
			for i, lang := range extension.Languages {
				expected, err := ioutil.ReadFile(fmt.Sprintf("testdata/default_config/%s.toolchain.yaml", lang))
				if err != nil {
					panic(errors.Wrapf(err, "missing default config for language %q", lang))
				}
				entries[i] = Entry(lang.String(), testCase{
					extension: &extension.Descriptor{
						Language: lang,
					},
					expected: string(expected),
				})
			}
			return entries
		}()...,
	)

	It("should panic if the programming language is unknown", func() {
		descriptor := &extension.Descriptor{
			Language: "",
		}

		Expect(func() { defaultConfigFor(descriptor) }).
			To(PanicWith(MatchError(`failed to determine default build image for unsupported programming language ""`)))
	})
})
