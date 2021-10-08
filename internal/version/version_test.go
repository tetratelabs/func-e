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

package version

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersion_IsDebug(t *testing.T) {
	tests := []struct {
		Input    Version
		Expected bool
	}{
		{
			Input:    "1.19",
			Expected: false,
		},
		{
			Input:    "1.19.1",
			Expected: false,
		},
		{
			Input:    "1.19_debug",
			Expected: true,
		},
		{
			Input:    "1.19.1_debug",
			Expected: true,
		},
	}

	for _, tc := range tests {
		actual := tc.Input.IsDebug()
		require.Equal(t, tc.Expected, actual)
	}
}

func TestVersion_MinorPrefix(t *testing.T) {
	tests := []struct {
		Input           Version
		WithTrailingDot bool
		Expected        string
	}{
		{
			Input:           "1.19",
			WithTrailingDot: true,
			Expected:        "1.19.",
		},
		{
			Input:           "1.19",
			WithTrailingDot: false,
			Expected:        "1.19",
		},
		{
			Input:           "1.19_debug",
			WithTrailingDot: true,
			Expected:        "1.19.",
		},
		{
			Input:           "1.19_debug",
			WithTrailingDot: false,
			Expected:        "1.19",
		},
		{
			Input:           "1.19.1",
			WithTrailingDot: true,
			Expected:        "1.19.",
		},
		{
			Input:           "1.19.1",
			WithTrailingDot: false,
			Expected:        "1.19",
		},
		{
			Input:           "1.19.1_debug",
			WithTrailingDot: true,
			Expected:        "1.19.",
		},
		{
			Input:           "1.19.1_debug",
			WithTrailingDot: false,
			Expected:        "1.19",
		},
	}

	for _, tc := range tests {
		actual := tc.Input.MinorPrefix(tc.WithTrailingDot)
		require.Equal(t, tc.Expected, actual)
	}
}
