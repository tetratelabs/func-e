// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/globals"
)

func TestHttpGet_AddsDefaultHeaders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range map[string]string{"User-Agent": "func-e/dev"} {
			require.Equal(t, v, r.Header.Get(k))
		}
	}))
	defer ts.Close()

	res, err := httpGet(context.Background(), ts.URL, globals.DefaultPlatform, "dev")
	require.NoError(t, err)

	defer res.Body.Close() //nolint:errcheck
	require.Equal(t, 200, res.StatusCode)
}
