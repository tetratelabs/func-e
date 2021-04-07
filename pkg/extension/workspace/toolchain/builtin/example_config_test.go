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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
)

func TestExampleConfigValidate(t *testing.T) {
	type testCase struct {
		name      string
		extension *extension.Descriptor
		expected  string
	}

	tests := make([]testCase, len(extension.Languages))
	for i, lang := range extension.Languages {
		expected, err := ioutil.ReadFile(fmt.Sprintf("testdata/example_config/%s.toolchain.yaml", lang))
		if err != nil {
			panic(fmt.Errorf("missing example config for language %s: %w", lang, err))
		}
		tests[i] = testCase{lang.String(), &extension.Descriptor{Language: lang}, string(expected)}
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			cfg := ExampleConfig(test.extension)
			require.Equal(t, test.expected, string(cfg))
		})
	}
}

func TestExampleConfigPanicsOnUnknownLanguage(t *testing.T) {
	require.PanicsWithError(t, `failed to determine default build image for unsupported programming language ""`, func() {
		_ = ExampleConfig(&extension.Descriptor{Language: ""})
	})
}
