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
		input    version.MinorVersion
		versions map[version.PatchVersion]version.Release
		want     version.PatchVersion
	}

	tests := []testCase{
		{
			name:  "zero",
			input: version.MinorVersion("1.20"),
			versions: map[version.PatchVersion]version.Release{
				version.PatchVersion("1.20.0_debug"): {}, // mixed is unlikely, but possible
				version.PatchVersion("1.20.0"):       {},
			},
			want: version.PatchVersion("1.20.0"),
		},
		{
			name:  "upgradable",
			input: version.MinorVersion("1.18"),
			versions: map[version.PatchVersion]version.Release{
				version.PatchVersion("1.18.3"):       {},
				version.PatchVersion("1.18.14"):      {},
				version.PatchVersion("1.18.4"):       {},
				version.PatchVersion("1.18.4_debug"): {},
			},
			want: version.PatchVersion("1.18.14"),
		},
		{
			name:  "notfound",
			input: version.MinorVersion("1.1"),
			versions: map[version.PatchVersion]version.Release{
				version.PatchVersion("1.20.0"):    {},
				version.PatchVersion("1.1_debug"): {},
			},
		},
		{
			name:  "debug",
			input: version.MinorVersion("1.19_debug"),
			versions: map[version.PatchVersion]version.Release{
				version.PatchVersion("1.19.10_debug"): {},
				version.PatchVersion("1.19.2_debug"):  {},
				version.PatchVersion("1.19.1"):        {},
			},
			want: version.PatchVersion("1.19.10_debug"),
		},
	}

	ctx := context.Background()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tester := newFuncEVersionsTester(tc.versions)
			actual, err := tester.feV.FindLatestPatch(ctx, tc.input)
			if tc.want == "" {
				require.Errorf(t, err, "couldn't find the latest patch for version %s", tc.input)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, actual)
			}
		})
	}
}

type funcEVersionsTester struct {
	feV funcEVersions
}

func newFuncEVersionsTester(versions map[version.PatchVersion]version.Release) funcEVersionsTester {
	return funcEVersionsTester{
		feV: funcEVersions{
			// Override Envoy versions getter for testing purpose only.
			getFunc: func(context.Context) (version.ReleaseVersions, error) {
				return version.ReleaseVersions{Versions: versions}, nil
			},
		},
	}
}
