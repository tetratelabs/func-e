// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package func_e_test

import (
	"net/http"
	"net/url"
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/require"

	func_e "github.com/tetratelabs/func-e"
	"github.com/tetratelabs/func-e/api"
	"github.com/tetratelabs/func-e/internal/test/httptest"
)

// TestRun shows func-e works inside synctest.Test without real network I/O.
func TestRun(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Pipe-backed server keeps all I/O in-process, compatible with synctest.
		var actualURL *url.URL
		ts := httptest.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actualURL = r.URL
			w.WriteHeader(http.StatusNotFound)
		}))

		// Route func-e's HTTP traffic through the test server via its transport.
		versionsURL := ts.URL + "/envoy-versions.json"
		err := func_e.Run(t.Context(), []string{"--config-yaml", "foo"},
			api.HomeDir(t.TempDir()), api.EnvoyVersionsURL(versionsURL), api.HTTPTransport(ts.Client().Transport))
		require.Error(t, err)
		require.Equal(t, "/envoy-versions.json", actualURL.String())
	})
}
