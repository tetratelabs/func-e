// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/admin"
	internalapi "github.com/tetratelabs/func-e/internal/api"
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
			delegate: func(ctx context.Context, _ internalapi.AdminClient) error {
				return nil
			},
		},
		{
			name: "delegate returns error",
			delegate: func(ctx context.Context, _ internalapi.AdminClient) error {
				return fmt.Errorf("delegate failed")
			},
			expectedLog: "delegate failed",
		},
		{
			name: "delegate panics",
			delegate: func(ctx context.Context, _ internalapi.AdminClient) error {
				panic("test panic")
			},
			expectedLog: "startup hook panicked: test panic",
		},
		{
			name: "delegate times out",
			delegate: func(ctx context.Context, _ internalapi.AdminClient) error {
				<-ctx.Done()
				return ctx.Err()
			},
			timeout:     10 * time.Millisecond,
			expectedLog: "context deadline exceeded",
		},
		{
			name: "no timeout set",
			delegate: func(ctx context.Context, _ internalapi.AdminClient) error {
				return fmt.Errorf("no timeout error")
			},
			timeout:     0,
			expectedLog: "no timeout error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logOutput string
			logf := func(format string, args ...interface{}) {
				logOutput = fmt.Sprintf(format, args...)
			}

			hook := &safeStartupHook{
				delegate: tt.delegate,
				logf:     logf,
				timeout:  tt.timeout,
			}

			tempDir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(tempDir, "envoy.pid"), []byte("12345"), 0o600))
			adminAddressPath := filepath.Join(tempDir, "admin-address.txt")
			require.NoError(t, os.WriteFile(adminAddressPath, []byte("127.0.0.1:12345"), 0o600))
			client, err := admin.NewAdminClient(t.Context(), tempDir, adminAddressPath)
			require.NoError(t, err)

			err = hook.Hook(t.Context(), client)
			require.NoError(t, err) // safeStartupHook should never return an error

			if tt.expectedLog != "" {
				require.Contains(t, logOutput, tt.expectedLog)
			} else {
				require.Empty(t, logOutput)
			}
		})
	}
}

func TestCollectConfigDump(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name          string
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name: "successful config dump",
			handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/config_dump", r.URL.Path)
				// require.Contains used here because RawQuery format may vary
				require.Contains(t, r.URL.RawQuery, "include_eds")
				_, _ = w.Write([]byte(`{"configs": [{"@type": "type.googleapis.com/envoy.admin.v3.EndpointsConfigDump"}]}`))
			},
		},
		{
			name: "timeout",
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(4 * time.Second)
			},
			expectedError: `could not read %[1]v: Get "%[1]v": context deadline exceeded`,
		},
		{
			name: "non-200 status",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
			},
			expectedError: "received 503 from %v",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			// Extract port from test server URL
			adminPort := ts.Listener.Addr().(*net.TCPAddr).Port

			ctx := t.Context()
			if tt.name == "timeout" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
				defer cancel()
			}

			require.NoError(t, os.WriteFile(filepath.Join(tempDir, "envoy.pid"), []byte("12345"), 0o600))
			adminAddressPath := filepath.Join(tempDir, "admin-address.txt")
			require.NoError(t, os.WriteFile(adminAddressPath, []byte(fmt.Sprintf("127.0.0.1:%d", adminPort)), 0o600))
			client, err := admin.NewAdminClient(t.Context(), tempDir, adminAddressPath)
			require.NoError(t, err)

			err = collectConfigDump(ctx, ts.Client(), client)
			if tt.expectedError != "" {
				require.Error(t, err)
				url := fmt.Sprintf("http://127.0.0.1:%d/config_dump?include_eds", adminPort)
				expectedErr := fmt.Sprintf(tt.expectedError, url)
				require.EqualError(t, err, expectedErr)
			} else {
				require.NoError(t, err)
				// Verify the file was created with expected content
				configPath := filepath.Join(tempDir, "config_dump.json")
				content, err := os.ReadFile(configPath)
				require.NoError(t, err)
				require.Equal(t, `{"configs": [{"@type": "type.googleapis.com/envoy.admin.v3.EndpointsConfigDump"}]}`, string(content))
			}
		})
	}
}

func TestCopyURLToFile(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name          string
		handler       http.HandlerFunc
		invalidPath   bool
		expectedError string
	}{
		{
			name: "successful copy",
			handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				_, _ = w.Write([]byte("test content"))
			},
		},
		{
			name: "http error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "server error", http.StatusInternalServerError)
			},
			expectedError: "received 500 from %v",
		},
		{
			name:          "invalid file path",
			invalidPath:   true,
			expectedError: `could not open "/invalid\x00path/test.txt": open /invalid` + "\x00" + `path/test.txt: invalid argument`,
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("test content"))
			},
		},
		{
			name: "context cancelled",
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Never respond - wait for context
				<-r.Context().Done()
			},
			expectedError: `could not read %[1]v: Get "%[1]v": context canceled`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			ctx := t.Context()
			if tt.name == "context cancelled" {
				var cancel context.CancelFunc
				// Use immediate cancellation - no timing dependency
				ctx, cancel = context.WithCancel(ctx)
				cancel() // Cancel immediately
			}

			filePath := filepath.Join(tempDir, "test.txt")
			if tt.invalidPath {
				filePath = "/invalid\x00path/test.txt"
			}

			err := copyURLToFile(ctx, ts.Client(), ts.URL, filePath)

			if tt.expectedError != "" {
				require.Error(t, err)
				if tt.invalidPath {
					// Invalid path error doesn't include URL
					require.EqualError(t, err, tt.expectedError)
				} else {
					expectedErr := fmt.Sprintf(tt.expectedError, ts.URL)
					require.EqualError(t, err, expectedErr)
				}
			} else {
				require.NoError(t, err)
				// Verify file content
				content, err := os.ReadFile(filePath)
				require.NoError(t, err)
				require.Equal(t, "test content", string(content))
				// Verify file permissions
				info, err := os.Stat(filePath)
				require.NoError(t, err)
				require.Equal(t, os.FileMode(0o600), info.Mode())
			}
		})
	}
}

func TestCopyURLToFile_InvalidURL(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.txt")

	// Test with invalid URL
	err := copyURLToFile(t.Context(), http.DefaultClient, "://invalid-url", filePath)
	require.Error(t, err)
	require.EqualError(t, err, "could not create request ://invalid-url: parse \"://invalid-url\": missing protocol scheme")
}
