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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/pkg/extension/workspace/config/extension"
)

func TestGenerateExtensionName(t *testing.T) {
	tests := []struct {
		category     extension.Category
		ExtensionDir string
		expected     string
	}{
		{
			category:     extension.EnvoyHTTPFilter,
			ExtensionDir: `My-Ext3nsi0n.com`,
			expected:     `mycompany.filters.http.my_ext3nsi0n_com`,
		},
		{
			category:     extension.EnvoyNetworkFilter,
			ExtensionDir: `My-Ext3nsi0n.com`,
			expected:     `mycompany.filters.network.my_ext3nsi0n_com`,
		},
		{
			category:     extension.EnvoyAccessLogger,
			ExtensionDir: `My-Ext3nsi0n.com`,
			expected:     `mycompany.access_loggers.my_ext3nsi0n_com`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.category.String(), func(t *testing.T) {
			actual := GenerateExtensionName(test.category, test.ExtensionDir)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestGenerateExtensionNamePanicsOnUnknownCategory(t *testing.T) {
	require.PanicsWithError(t, `unknown extension category "my-category"`, func() {
		_ = GenerateExtensionName(`my-category`, `ExtensionDir`)
	})
}
