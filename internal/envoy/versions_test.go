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

	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/test"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestNewGetVersions(t *testing.T) {
	versionsServer := test.RequireEnvoyVersionsTestServer(t, version.LastKnownEnvoy)
	gv := NewGetVersions(versionsServer.URL+"/envoy-versions.json", globals.DefaultPlatform, "dev")

	evs, err := gv(context.Background())
	require.NoError(t, err)
	require.Contains(t, evs.Versions, version.LastKnownEnvoy)
}
