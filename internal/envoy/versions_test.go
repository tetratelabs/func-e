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

package envoy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/version"
)

func TestFuncEVersions_FindLatestPatch(t *testing.T) {
	type testCase struct {
		name     string
		input    version.Version
		versions map[version.Version]version.Release
		want     version.Version
	}

	tests := []testCase{
		{
			name:  "zero",
			input: "1.20",
			versions: map[version.Version]version.Release{
				"1.20.0_debug": {},
				"1.20.0":       {},
			},
			want: "1.20.0",
		},
		{
			name:  "upgradable",
			input: "1.18",
			versions: map[version.Version]version.Release{
				"1.18.3":       {},
				"1.18.14":     {},
				"1.18.4":       {},
				"1.18.4_debug": {},
			},
			want: "1.18.4",
		},
		{
			name:  "notfound",
			input: "1.1",
			versions: map[version.Version]version.Release{
				"1.20.0":    {},
				"1.1_debug": {},
			},
			want: "",
		},
		{
			name:  "debug",
			input: "1.19_debug",
			versions: map[version.Version]version.Release{
				"1.19.10_debug": {},
				"1.19.2_debug": {},
				"1.19.1":       {},
			},
			want: "1.19.1_debug",
		},
	}

	ctx := context.Background()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tester := newFuncEVersionsTester(tc.versions)
			have, err := tester.feV.FindLatestPatch(ctx, tc.input)
			if tc.want == "" {
				require.Errorf(t, err, "couldn't find latest version for %s", tc.input)
			} else {
				require.NoError(t, err)
				require.Equal(t, have, tc.want)
			}
		})
	}
}

type funcEVersionsTester struct {
	feV funcEVersions
}

func newFuncEVersionsTester(versions map[version.Version]version.Release) funcEVersionsTester {
	return funcEVersionsTester{
		feV: funcEVersions{
			// Override Envoy versions getter for testing purpose only.
			getFunc: func(context.Context) (version.ReleaseVersions, error) {
				return version.ReleaseVersions{Versions: versions}, nil
			},
		},
	}
}
