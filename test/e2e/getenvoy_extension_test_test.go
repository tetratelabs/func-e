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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
)

// TestGetEnvoyExtensionTest runs the equivalent of "getenvoy extension test" for a matrix of extension.Categories and
// extension.Languages. "getenvoy extension init" is a prerequisite, so run first.
//
// "getenvoy extension test" uses Docker. See TestMain for general notes on about the test runtime.
func TestGetEnvoyExtensionTest(t *testing.T) {
	const extensionName = "getenvoy_extension_test"

	for _, test := range getExtensionTestMatrix() {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.String(), func(t *testing.T) {
			workDir, removeWorkDir := requireNewTempDir(t)
			defer removeWorkDir()

			revertChDir := requireChDir(t, workDir)
			defer revertChDir()

			// test requires "get envoy extension init" to have succeeded
			requireExtensionInit(t, workDir, test.Category, test.Language, extensionName)
			defer requireExtensionClean(t, workDir)

			cmd := getEnvoy("extension test").Args(getToolchainContainerOptions()...)
			// "getenvoy extension test" only returns stdout because `docker run -t` redirects stderr to stdout.
			stdout := requireExecNoStderr(t, cmd)

			// Verify the tests ran
			switch test.Language {
			case extension.LanguageRust:
				// `cargo` colorizes output. After stripping ANSI codes, ensure the output is successful.
				stdout = stripAnsiEscapeRegexp.ReplaceAllString(stdout, "")
				require.Regexp(t, `(?s)^.*test result: ok.*$`, stdout, `invalid stdout running [%v]`, cmd)

			case extension.LanguageTinyGo:
				// We expect the test output to include the extension name.
				stdoutRegexp := fmt.Sprintf(`(?s)^.*ok  	%s.*$`, extensionName)
				require.Regexp(t, stdoutRegexp, stdout, `invalid stdout running [%v]`, cmd)
			}
		})
	}
}
