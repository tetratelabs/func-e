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

	internalapi "github.com/tetratelabs/func-e/internal/api"
)

const userAgentHeader = "User-Agent"

// httpGet GETs rawURL with a User-Agent header and one retry on transient
// network errors.
func httpGet(ctx context.Context, clientFn internalapi.HTTPClientFunc, rawURL, ua string) (*http.Response, error) {
	client := clientFn()

	for attempt := 0; ; attempt++ {
		// #nosec -> url can be anywhere by design
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
		if err != nil {
			return nil, err
		}
		req.Header.Add(userAgentHeader, ua)

		resp, err := client.Do(req)

		if resp != nil || err == nil || ctx.Err() != nil || !isNetError(err) || attempt > 0 {
			return resp, err
		}

		// Budget the retry delay from the remaining deadline to avoid wasting
		// it on sleep (capped at 1s when there's no deadline).
		delay := time.Second
		if deadline, ok := ctx.Deadline(); ok {
			if half := time.Until(deadline) / 2; half < delay {
				delay = half
			}
		}
		if delay <= 0 {
			return resp, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}
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
