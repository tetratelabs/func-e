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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
	"github.com/tetratelabs/getenvoy/pkg/test/morerequire"
)

func TestInterpolateExtensionName(t *testing.T) {
	e := extension.Descriptor{
		Name: "my_company.my_extension",
	}
	actual, err := interpolate(&e)("src/factory.rs", []byte(`
impl<'a> ExtensionFactory for SampleHttpFilterFactory<'a> {
    type Extension = SampleHttpFilter<'a>;

    /// The reference name for Sample HTTP Filter.
    ///
    /// This name appears in "Envoy" configuration as a value of "root_id" field
    /// (also known as "group_name").
	const NAME: &'static str = "{{ .Extension.Name }}";
}
`))

	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(`
impl<'a> ExtensionFactory for SampleHttpFilterFactory<'a> {
    type Extension = SampleHttpFilter<'a>;

    /// The reference name for Sample HTTP Filter.
    ///
    /// This name appears in "Envoy" configuration as a value of "root_id" field
    /// (also known as "group_name").
	const NAME: &'static str = "%s";
}
`, e.Name), string(actual))
}

func TestScaffold(t *testing.T) {
	tests := []struct {
		name      string
		extension *extension.Descriptor
		file      string
		expected  string
	}{
		{
			name: "rust/filters/http",
			extension: &extension.Descriptor{
				Name:     "my_company.my_extension",
				Category: extension.EnvoyHTTPFilter,
				Language: extension.LanguageRust,
			},
			file:     "src/factory.rs",
			expected: `"my_company.my_extension"`,
		},
		{
			name: "rust/filters/network",
			extension: &extension.Descriptor{
				Name:     "my_company.my_extension",
				Category: extension.EnvoyNetworkFilter,
				Language: extension.LanguageRust,
			},
			file:     "src/factory.rs",
			expected: `"my_company.my_extension"`,
		},
		{
			name: "rust/access_logger",
			extension: &extension.Descriptor{
				Name:     "my_company.my_extension",
				Category: extension.EnvoyAccessLogger,
				Language: extension.LanguageRust,
			},
			file:     "src/logger.rs",
			expected: `"my_company.my_extension"`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			extensionDir, revertExtensionDir := morerequire.RequireNewTempDir(t)
			defer revertExtensionDir()

			opts := &ScaffoldOpts{
				Extension:    test.extension,
				TemplateName: "default",
				ExtensionDir: extensionDir,
			}

			err := Scaffold(opts)
			require.NoError(t, err)

			actual, err := os.ReadFile(filepath.Join(extensionDir, test.file))
			require.NoError(t, err)
			require.Contains(t, string(actual), test.expected)
		})
	}
}
