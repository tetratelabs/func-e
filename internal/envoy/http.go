// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"context"
	"errors"
	"net"
	"net/http"
	neturl "net/url"
	"time"
)

const userAgentHeader = "User-Agent"

// httpGet GETs rawURL with a User-Agent header and one retry on transient network error.
func httpGet(ctx context.Context, client *http.Client, rawURL, ua string) (*http.Response, error) {
	// #nosec -> url can be anywhere by design
	get := func() (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
		if err != nil {
			return nil, err
		}
		req.Header.Add(userAgentHeader, ua)
		return client.Do(req)
	}

	resp, err := get()

	// Return unless this hit a transient network error worth retrying.
	if resp != nil || err == nil || ctx.Err() != nil || !isNetError(err) {
		return resp, err
	}

	// Wait up to 1s before retrying, or bail if the context is canceled.
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(time.Second):
	}

	return get()
}

// isNetError unwraps url.Error so transient dial/TLS failures are retried
// while HTTP-level errors (4xx, 5xx) are not.
func isNetError(err error) bool {
	if urlErr, ok := errors.AsType[*neturl.Error](err); ok {
		err = urlErr.Err
	}

	netErr, ok := errors.AsType[net.Error](err)
	return ok && netErr != nil
}
