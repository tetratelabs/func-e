// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/admin"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/test"
	"github.com/tetratelabs/func-e/internal/test/httptest"
	"github.com/tetratelabs/func-e/internal/version"
)

func TestNewGetVersions(t *testing.T) {
	baseURL := "http://" + admin.ServerAddr
	handler := test.NewEnvoyVersionsHandler(t, baseURL, version.LastKnownEnvoy)
	gv := NewGetVersions(httptest.HTTPClient(handler), baseURL+"/envoy-versions.json", globals.DefaultDevUserAgent)

	evs, err := gv(t.Context())
	require.NoError(t, err)
	require.Contains(t, evs.Versions, version.LastKnownEnvoy)
}
