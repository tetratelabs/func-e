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
	"fmt"
	"net/http"
	"strings"
)

// httpGet adds the userAgent header to the request, so that we can tell what is a dev build vs release.
func httpGet(ctx context.Context, url, platform, version string) (*http.Response, error) {
	// #nosec -> url can be anywhere by design
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent(platform, version))
	return http.DefaultClient.Do(req)
}

// userAgent returns the 'User-Agent' header value used in HTTP requests. This is useful in log, metrics, analytics, or
// request filtering. As this is a CLI, the best 'User-Agent' is the binary version including platform.
//
// The returned value limits cardinality to formal release * platform or one value for all non-releases.
//
// Note: Analytics may not work out-of-box. For example, Netlify does not support server-side analytics on 'User-Agent',
// and even its 'Referer' analytics are limited to requests to HTML resources.
func userAgent(platform, version string) string {
	if !strings.HasPrefix(version, "v") || strings.Contains(version, "SNAPSHOT") {
		return "getenvoy/dev"
	}
	return fmt.Sprintf("getenvoy/%s (%s)", version, platform)
}
