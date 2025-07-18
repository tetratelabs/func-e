// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/tetratelabs/func-e/internal/version"
)

// httpGet adds the userAgent header to the request, so that we can tell what is a dev build vs release.
func httpGet(ctx context.Context, url string, p version.Platform, v string) (*http.Response, error) {
	// #nosec -> url can be anywhere by design
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent(p, v))
	return http.DefaultClient.Do(req)
}

// userAgent returns the 'User-Agent' header value used in HTTP requests. This is useful in log, metrics, analytics, or
// request filtering. As this is a CLI, the best 'User-Agent' is the binary version including platform.
//
// The returned value limits cardinality to formal release * platform or one value for all non-releases.
//
// Note: Analytics may not work out-of-box. For example, Netlify does not support server-side analytics on 'User-Agent',
// and even its 'Referer' analytics are limited to requests to HTML resources.
func userAgent(p version.Platform, v string) string {
	if !strings.HasPrefix(v, "v") || strings.Contains(v, "SNAPSHOT") {
		return "func-e/dev"
	}
	return fmt.Sprintf("func-e/%s (%s)", v, p)
}
