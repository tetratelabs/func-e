// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/admin"
	internalapi "github.com/tetratelabs/func-e/internal/api"
	"github.com/tetratelabs/func-e/internal/test/httptest"
)

func TestSafeStartupHook(t *testing.T) {
	tests := []struct {
		name        string
		delegate    internalapi.StartupHook
		timeout     time.Duration
		expectedLog string
	}{
		{
			name: "successful delegate",
			delegate: func(_ context.Context, _ internalapi.AdminClient, _ string) error {
				return nil
			},
		},
		{
			name: "delegate returns error",
			delegate: func(_ context.Context, _ internalapi.AdminClient, _ string) error {
				return errors.New("delegate failed")
			},
			expectedLog: "delegate failed",
		},
		{
			name: "delegate panics",
			delegate: func(_ context.Context, _ internalapi.AdminClient, _ string) error {
				panic("test panic")
			},
			expectedLog: "startup hook panicked: test panic",
		},
		{
			name: "delegate times out",
			delegate: func(ctx context.Context, _ internalapi.AdminClient, _ string) error {
				<-ctx.Done()
				return ctx.Err()
			},
			timeout:     10 * time.Millisecond,
			expectedLog: "context deadline exceeded",
		},
		{
			name: "no timeout set",
			delegate: func(_ context.Context, _ internalapi.AdminClient, _ string) error {
				return errors.New("no timeout error")
			},
			timeout:     0,
			expectedLog: "no timeout error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logOutput string
			logf := func(format string, args ...any) {
				logOutput = fmt.Sprintf(format, args...)
			}

			hook := &safeStartupHook{
				delegate: tt.delegate,
				logf:     logf,
				timeout:  tt.timeout,
			}

			err := hook.Hook(t.Context(), nil, "test-run-id")
			require.NoError(t, err) // safeStartupHook should never return an error

			if tt.expectedLog != "" {
				require.Contains(t, logOutput, tt.expectedLog)
			} else {
				require.Empty(t, logOutput)
			}
		})
	}
}

func TestSafeStartupHook_TimeoutBoundary(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const timeout = 5 * time.Second

		sawLiveContext := false
		sawDeadline := false
		hook := &safeStartupHook{
			delegate: func(ctx context.Context, _ internalapi.AdminClient, _ string) error {
				time.Sleep(timeout - time.Nanosecond)
				synctest.Wait()
				sawLiveContext = ctx.Err() == nil

				time.Sleep(time.Nanosecond)
				synctest.Wait()
				sawDeadline = errors.Is(ctx.Err(), context.DeadlineExceeded)
				return ctx.Err()
			},
			logf:    func(string, ...any) {},
			timeout: timeout,
		}

		err := hook.Hook(t.Context(), nil, "test-run-id")
		require.NoError(t, err)
		require.True(t, sawLiveContext)
		require.True(t, sawDeadline)
	})
}

func TestCollectConfigDump(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		ctx         func(*testing.T) context.Context
		expectedErr string
	}{
		{
			name: "successful config dump",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"configs": [{"@type": "type.googleapis.com/envoy.admin.v3.EndpointsConfigDump"}]}`))
			},
		},
		{
			name: "timeout",
			handler: func(_ http.ResponseWriter, r *http.Request) {
				<-r.Context().Done()
			},
			ctx: func(t *testing.T) context.Context {
				t.Helper()
				ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
				t.Cleanup(cancel)
				return ctx
			},
			expectedErr: `error Envoy admin URL $URL: Get "$URL": context deadline exceeded`,
		},
		{
			name: "non-200 status",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
			},
			expectedErr: "error Envoy admin URL $URL: status_code=503,body:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			ctx := t.Context()
			if tt.ctx != nil {
				ctx = tt.ctx(t)
			}

			actualPath := ""
			actualQuery := ""
			client, err := admin.NewAdminClientForURL("http://"+admin.ServerAddr, httptest.HTTPClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				actualPath = r.URL.Path
				actualQuery = r.URL.RawQuery
				tt.handler(w, r)
			})))
			require.NoError(t, err)

			err = collectConfigDump(ctx, client, tempDir)
			expectedURL := "http://" + admin.ServerAddr + "/config_dump?include_eds"
			if tt.expectedErr != "" {
				require.EqualError(t, err, strings.ReplaceAll(tt.expectedErr, "$URL", expectedURL))
			} else {
				require.NoError(t, err)
				configPath := filepath.Join(tempDir, "config_dump.json")
				content, err := os.ReadFile(configPath)
				require.NoError(t, err)
				require.JSONEq(t, `{"configs": [{"@type": "type.googleapis.com/envoy.admin.v3.EndpointsConfigDump"}]}`, string(content))
			}
			require.Equal(t, "/config_dump", actualPath)
			require.Equal(t, "include_eds", actualQuery)
		})
	}
}

func TestCopyURLToFile(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		ctx         func(*testing.T) context.Context
		invalidPath bool
		expectedErr string
	}{
		{
			name: "successful copy",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte("test content"))
			},
		},
		{
			name: "http error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				http.Error(w, "server error", http.StatusInternalServerError)
			},
			expectedErr: "received 500 from $URL",
		},
		{
			name:        "invalid file path",
			invalidPath: true,
			expectedErr: `could not open "/invalid\x00path/test.txt": open /invalid` + "\x00" + `path/test.txt: invalid argument`,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte("test content"))
			},
		},
		{
			name: "context canceled",
			ctx: func(t *testing.T) context.Context {
				t.Helper()
				ctx, cancel := context.WithCancel(t.Context())
				cancel()
				return ctx
			},
			handler: func(_ http.ResponseWriter, r *http.Request) {
				<-r.Context().Done()
			},
			expectedErr: `could not read $URL: Get "$URL": context canceled`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			ctx := t.Context()
			if tt.ctx != nil {
				ctx = tt.ctx(t)
			}

			filePath := filepath.Join(tempDir, "test.txt")
			if tt.invalidPath {
				filePath = "/invalid\x00path/test.txt"
			}

			actualMethod := ""
			client := httptest.HTTPClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				actualMethod = r.Method
				tt.handler(w, r)
			}))
			url := "http://" + admin.ServerAddr
			err := copyURLToFile(ctx, client, url, filePath)

			if tt.expectedErr != "" {
				require.Error(t, err)
				expectedErr := strings.ReplaceAll(tt.expectedErr, "$URL", url)
				require.EqualError(t, err, expectedErr)
			} else {
				require.NoError(t, err)
				content, err := os.ReadFile(filePath)
				require.NoError(t, err)
				require.Equal(t, "test content", string(content))
				info, err := os.Stat(filePath)
				require.NoError(t, err)
				require.Equal(t, os.FileMode(0o600), info.Mode())
			}
			if !tt.invalidPath {
				require.Equal(t, http.MethodGet, actualMethod)
			}
		})
	}
}

func TestCopyURLToFile_InvalidURL(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.txt")

	err := copyURLToFile(t.Context(), httptest.HTTPClient(http.NotFoundHandler()), "://invalid-url", filePath)
	require.Error(t, err)
	require.EqualError(t, err, "could not create request ://invalid-url: parse \"://invalid-url\": missing protocol scheme")
}
