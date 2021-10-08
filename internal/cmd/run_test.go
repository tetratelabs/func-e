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

package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/version"
)

func TestExtractLatestPatchFormat(t *testing.T) {
	tests := []struct {
		Input    version.Version
		Expected version.Version
	}{
		{
			Input:    "1.1",
			Expected: "1.1",
		},
		{
			Input:    "1.18",
			Expected: "1.18",
		},
		{
			Input:    "1.18_debug",
			Expected: "1.18_debug",
		},
		{
			Input:    "1.1.1",
			Expected: "1.1",
		},
		{
			Input:    "1.18.1",
			Expected: "1.18",
		},
		{
			Input:    "1.18.1_debug",
			Expected: "1.18_debug",
		},
	}

	for _, tc := range tests {
		actual := extractLatestPatchFormat(tc.Input)
		require.Equal(t, tc.Expected, actual)
	}
}
