// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/admin"
	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/test/httptest"
)

func TestHttpGet_AddsUserAgent(t *testing.T) {
	actualUserAgent := ""
	client := httptest.HTTPClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualUserAgent = r.Header.Get(userAgentHeader)
		w.WriteHeader(http.StatusOK)
	}))

	res, err := httpGet(t.Context(), client, "http://"+admin.ServerAddr+"/", globals.DefaultDevUserAgent)
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(t, globals.DefaultDevUserAgent, actualUserAgent)
}

func TestHttpGet_RetryDecisions(t *testing.T) {
	tests := []struct {
		name             string
		ctx              func(*testing.T) context.Context
		dialErr          error
		expectedErr      string
		expectedDials    int
		expectedRequests int
		expectedElapsed  time.Duration
	}{
		{
			name: "retries net error",
			ctx: func(t *testing.T) context.Context {
				t.Helper()
				return t.Context()
			},
			dialErr:          netError{err: errors.New("connection refused")},
			expectedDials:    2,
			expectedRequests: 1,
			expectedElapsed:  time.Second,
		},
		{
			name: "long deadline caps retry delay at 1s",
			ctx: func(t *testing.T) context.Context {
				t.Helper()
				ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
				t.Cleanup(cancel)
				return ctx
			},
			dialErr:          netError{err: errors.New("connection refused")},
			expectedDials:    2,
			expectedRequests: 1,
			expectedElapsed:  time.Second,
		},
		{
			name: "cancel during retry sleep",
			ctx: func(t *testing.T) context.Context {
				t.Helper()
				ctx, cancel := context.WithCancel(t.Context())
				go func() {
					time.Sleep(500 * time.Millisecond)
					cancel()
				}()
				return ctx
			},
			dialErr:          netError{err: errors.New("connection refused")},
			expectedErr:      context.Canceled.Error(),
			expectedDials:    1,
			expectedRequests: 0,
			expectedElapsed:  500 * time.Millisecond,
		},
		{
			name: "does not retry non-net error",
			ctx: func(t *testing.T) context.Context {
				t.Helper()
				return t.Context()
			},
			dialErr:          errors.New("connection refused"),
			expectedErr:      `Get "$URL": connection refused`,
			expectedDials:    1,
			expectedRequests: 0,
		},
		{
			name: "does not dial canceled context",
			ctx: func(t *testing.T) context.Context {
				t.Helper()
				ctx, cancel := context.WithCancel(t.Context())
				cancel()
				return ctx
			},
			expectedErr:      `Get "$URL": context canceled`,
			expectedDials:    0,
			expectedRequests: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				// Count requests that reach the server (past the dial stage).
				requests := 0
				ts := httptest.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					requests++
					w.WriteHeader(http.StatusOK)
				}))

				// Swap the client context with one that fails on the first dial.
				dials := 0
				client := ts.Client()
				transport := client.Transport.(*http.Transport)
				dialContext := transport.DialContext
				transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
					dials++
					if dials == 1 && tt.dialErr != nil {
						return nil, tt.dialErr
					}
					return dialContext(ctx, network, addr)
				}

				// Execute the request and verify the outcome.
				start := time.Now()
				res, err := httpGet(tt.ctx(t), client, ts.URL, globals.DefaultDevUserAgent)
				if tt.expectedErr != "" {
					expectedErr := strings.ReplaceAll(tt.expectedErr, "$URL", ts.URL)
					require.EqualError(t, err, expectedErr)
				} else {
					require.NoError(t, err)
					defer res.Body.Close()
					require.Equal(t, http.StatusOK, res.StatusCode)
				}

				require.Equal(t, tt.expectedDials, dials)
				require.Equal(t, tt.expectedRequests, requests)
				require.Equal(t, tt.expectedElapsed, time.Since(start))
			})
		})
	}
}

var _ net.Error = netError{}

type netError struct {
	err error
}

func (e netError) Error() string {
	return e.err.Error()
}

func (e netError) Unwrap() error {
	return e.err
}

func (netError) Timeout() bool {
	return false
}

func (netError) Temporary() bool {
	return false
}
