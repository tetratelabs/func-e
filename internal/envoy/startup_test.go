// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package envoy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSafeStartupHook(t *testing.T) {
	tests := []struct {
		name     string
		delegate StartupHook
		timeout  time.Duration
		wantLog  string
	}{
		{
			name: "successful delegate",
			delegate: func(ctx context.Context, runDir, adminAddress string) error {
				return nil
			},
		},
		{
			name: "delegate returns error",
			delegate: func(ctx context.Context, runDir, adminAddress string) error {
				return fmt.Errorf("delegate failed")
			},
			wantLog: "delegate failed",
		},
		{
			name: "delegate panics",
			delegate: func(ctx context.Context, runDir, adminAddress string) error {
				panic("test panic")
			},
			wantLog: "startup hook panicked: test panic",
		},
		{
			name: "delegate times out",
			delegate: func(ctx context.Context, runDir, adminAddress string) error {
				<-ctx.Done()
				return ctx.Err()
			},
			timeout: 10 * time.Millisecond,
			wantLog: "context deadline exceeded",
		},
		{
			name: "no timeout set",
			delegate: func(ctx context.Context, runDir, adminAddress string) error {
				return fmt.Errorf("no timeout error")
			},
			timeout: 0,
			wantLog: "no timeout error",
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

			err := hook.Hook(t.Context(), "/tmp", "127.0.0.1:9901")
			require.NoError(t, err) // safeStartupHook should never return an error

			if tt.wantLog != "" {
				require.Contains(t, logOutput, tt.wantLog)
			} else {
				require.Empty(t, logOutput)
			}
		})
	}
}

func TestCollectConfigDump(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr string
	}{
		{
			name: "successful config dump",
			handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/config_dump", r.URL.Path)
				require.Contains(t, r.URL.RawQuery, "include_eds")
				_, _ = w.Write([]byte(`{"configs": [{"@type": "type.googleapis.com/envoy.admin.v3.EndpointsConfigDump"}]}`))
			},
		},
		{
			name: "timeout",
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(4 * time.Second)
			},
			wantErr: "context deadline exceeded",
		},
		{
			name: "non-200 status",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
			},
			wantErr: "received 503",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			// Extract host:port from test server URL
			adminAddress := ts.URL[7:] // strip "http://"

			ctx := t.Context()
			if tt.name == "timeout" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
				defer cancel()
			}
			err := collectConfigDump(ctx, ts.Client(), tempDir, adminAddress)

			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
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
		name        string
		handler     http.HandlerFunc
		invalidPath bool
		wantErr     string
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
			wantErr: "received 500",
		},
		{
			name:        "invalid file path",
			invalidPath: true,
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("test content"))
			},
			wantErr: "could not open",
		},
		{
			name: "context cancelled",
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Never respond
				<-r.Context().Done()
			},
			wantErr: "context canceled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			ctx := t.Context()
			if tt.name == "context cancelled" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				go func() {
					time.Sleep(10 * time.Millisecond)
					cancel()
				}()
			}

			filePath := filepath.Join(tempDir, "test.txt")
			if tt.invalidPath {
				filePath = "/invalid\x00path/test.txt"
			}

			err := copyURLToFile(ctx, ts.Client(), ts.URL, filePath)

			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
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
	require.Contains(t, err.Error(), "could not create request")
}
