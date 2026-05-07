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
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

// TestRun shows you can use func-e inside synctest.Test, by overriding the
// api.HTTPTransport threaded through func-e. More advanced test cases override
// api.RunFunc to handle app behavior all without breaking the synctest bubble!
func TestRun(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var actualURL *url.URL
		transportFn := func() http.RoundTripper {
			return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				actualURL = req.URL
				return &http.Response{StatusCode: http.StatusNotFound, Body: http.NoBody}, nil
			})
		}
		err := func_e.Run(t.Context(), []string{}, api.HomeDir(t.TempDir()), api.HTTPTransport(transportFn))
		require.Error(t, err)
		require.Equal(t, "https://archive.tetratelabs.io/envoy/envoy-versions.json", actualURL.String())
	})
}
