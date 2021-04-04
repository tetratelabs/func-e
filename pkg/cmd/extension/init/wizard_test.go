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
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	. "github.com/tetratelabs/getenvoy/pkg/test/morerequire"
	uiutil "github.com/tetratelabs/getenvoy/pkg/util/ui"
)

func TestWizardArgs(t *testing.T) {
	tempDir, revertTempDir := RequireNewTempDir(t)
	defer revertTempDir()

	type testCase struct {
		name     string
		noColors bool
		expected string
	}
	tests := []testCase{
		{
			name:     "colors disabled",
			noColors: true,
			expected: fmt.Sprintf(`What kind of extension would you like to create?
* Category HTTP Filter
* Language Rust
* Output directory %s
* Extension name mycompany.filters.http.custom_metrics
Great! Let me help you with that!

`, tempDir),
		},
		{
			name:     "colors enabled",
			noColors: false,
			expected: "\x1b[4mWhat kind of extension would you like to create?\x1b[0m\n" +
				"\x1b[32m✔\x1b[0m \x1b[3mCategory\x1b[0m \x1b[2mHTTP Filter\x1b[0m\n" +
				"\x1b[32m✔\x1b[0m \x1b[3mLanguage\x1b[0m \x1b[2mRust\x1b[0m\n" +
				fmt.Sprintf("\x1b[32m✔\x1b[0m \x1b[3mOutput directory\x1b[0m \x1b[2m%s\x1b[0m\n", tempDir) +
				"\x1b[32m✔\x1b[0m \x1b[3mExtension name\x1b[0m \x1b[2mmycompany.filters.http.custom_metrics\x1b[0m\n" +
				"Great! Let me help you with that!\n" +
				"\n",
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			uiutil.StylesEnabled = !test.noColors

			c := new(cobra.Command)
			out := new(bytes.Buffer)
			c.SetOut(out)

			params := newParams()
			params.Category.Value = "envoy.filters.http"
			params.Language.Value = "rust"
			params.OutputDir.Value = tempDir
			params.Name.Value = "mycompany.filters.http.custom_metrics"

			err := newWizard(c).Fill(params)

			require.NoError(t, err)
			require.Equal(t, test.expected, out.String(), `unexpected output running [%v]`, c)
		})
	}
}
