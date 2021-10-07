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

package globals_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/globals"
)

func TestEnvoyVersionPattern_Valid(t *testing.T) {
	versions := []string{
		"1.1.1", "1.18.1", "1.18.1_debug",
	}

	for _, v := range versions {
		require.True(t, globals.EnvoyVersionPattern.MatchString(v))
	}
}

func TestEnvoyVersionPattern_Invalid(t *testing.T) {
	versions := []string{
		"a.b.c", "1", "1.1", "1.1_debug",
	}

	for _, v := range versions {
		require.False(t, globals.EnvoyVersionPattern.MatchString(v))
	}
}

func TestEnvoyStrictMinorVersionPattern_Valid(t *testing.T) {
	versions := []string{
		"1.1", "1.18", "1.18_debug",
	}

	for _, v := range versions {
		require.True(t, globals.EnvoyStrictMinorVersionPattern.MatchString(v))
	}
}

func TestEnvoyStrictMinorVersionPattern_Invalid(t *testing.T) {
	versions := []string{
		"a.b",
		"1.18-debug",
		"1", "1.", ".1",
		"1.1.1", "1.1.1_debug",
	}

	for _, v := range versions {
		require.False(t, globals.EnvoyStrictMinorVersionPattern.MatchString(v))
	}
}

func TestEnvoyMinorVersionPattern_CapturePatchComponent(t *testing.T) {
	type testCase struct {
		version                string
		capturedPatchComponent string
	}

	tests := []testCase{
		{
			version:                "1.1",
			capturedPatchComponent: "",
		},
		{
			version:                "1.1.1",
			capturedPatchComponent: ".1",
		},
		{
			version:                "1.1.1_debug",
			capturedPatchComponent: ".1",
		},
	}

	for _, tc := range tests {
		var matched [][]string
		if matched = globals.EnvoyMinorVersionPattern.FindAllStringSubmatch(tc.version, -1); matched != nil {
			for _, sub := range matched {
				require.Equal(t, sub[1], tc.capturedPatchComponent)
			}
			continue
		}
	}
}

func TestEnvoyLatestPatchVersionPattern_CaptureLatestPatchVersion(t *testing.T) {
	tests := []struct {
		Input    string
		Expected string
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
			Expected: "1.18",
		},
	}

	for _, tc := range tests {
		actual := globals.EnvoyLatestPatchVersionPattern.FindString(tc.Input)
		require.Equal(t, tc.Expected, actual)
	}
}
