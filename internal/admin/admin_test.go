// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPollEnvoyPidAndAdminAddressPathForPort(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T, path string)
		ctx           func(t *testing.T) context.Context
		expectedPort  int
		expectedError string
	}{
		{
			name: "file appears after delay",
			setup: func(t *testing.T, path string) {
				t.Helper()
				go func() {
					time.Sleep(100 * time.Millisecond)
					_ = os.WriteFile(path, []byte("127.0.0.1:9901\n"), 0o600)
				}()
			},
			ctx:          func(t *testing.T) context.Context { return t.Context() },
			expectedPort: 9901,
		},
		{
			name: "timeout when file never appears",
			setup: func(t *testing.T, path string) {
				t.Helper()
			},
			ctx: func(t *testing.T) context.Context {
				t.Helper()
				ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
				t.Cleanup(cancel)
				return ctx
			},
			expectedError: "timeout waiting for Envoy admin address file",
		},
		{
			name: "extracts port from address with any hostname",
			setup: func(t *testing.T, path string) {
				t.Helper()
				require.NoError(t, os.WriteFile(path, []byte("localhost:9901"), 0o600))
			},
			ctx:          func(t *testing.T) context.Context { return t.Context() },
			expectedPort: 9901,
		},
		{
			name: "invalid address format",
			setup: func(t *testing.T, path string) {
				t.Helper()
				require.NoError(t, os.WriteFile(path, []byte("invalid-address"), 0o600))
			},
			ctx:           func(t *testing.T) context.Context { return t.Context() },
			expectedError: "failed to parse Envoy's admin port: strconv.Atoi: parsing \"\": invalid syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "admin-address.txt")
			tt.setup(t, path)
			actualPort, err := pollAdminAddressPathForPort(tt.ctx(t), path)
			if tt.expectedError != "" {
				// The timeout error includes the file path, others don't
				if tt.name == "timeout when file never appears" {
					expectedErr := fmt.Sprintf("%s: open %s: no such file or directory", tt.expectedError, path)
					require.EqualError(t, err, expectedErr)
				} else {
					require.EqualError(t, err, tt.expectedError)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedPort, actualPort)
			}
		})
	}
}

func setupTestServer(t *testing.T, handler *http.HandlerFunc) *adminClient {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		(*handler)(w, r)
	}))
	t.Cleanup(server.Close)
	u, err := url.Parse(server.URL)
	require.NoError(t, err)
	port, err := strconv.Atoi(u.Port())
	require.NoError(t, err)
	return &adminClient{port: port}
}

func TestAdminClient_get(t *testing.T) {
	var handler http.HandlerFunc
	client := setupTestServer(t, &handler)

	tests := []struct {
		name          string
		handler       http.HandlerFunc
		ctx           func(t *testing.T) context.Context
		path          string
		expected      []byte
		expectedError string
	}{
		{
			name: "returns body on 200 status",
			handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/test", r.URL.Path)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("response body"))
			},
			ctx:      func(t *testing.T) context.Context { return t.Context() },
			path:     "/test",
			expected: []byte("response body"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler = tt.handler
			actual, err := client.Get(tt.ctx(t), tt.path)
			if tt.expectedError != "" {
				expectedErr := fmt.Sprintf("error Envoy admin URL http://127.0.0.1:%d%s: %s", client.port, tt.path, tt.expectedError)
				require.EqualError(t, err, expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, actual)
			}
		})
	}

	t.Run("returns error on non-200 status code", func(t *testing.T) {
		handler = func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("not ready"))
		}
		_, err := client.Get(t.Context(), "/test")
		expectedErr := fmt.Sprintf("error Envoy admin URL http://127.0.0.1:%d/test: status_code=503,body:not ready", client.port)
		require.EqualError(t, err, expectedErr)
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		handler = func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("response"))
		}
		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
		t.Cleanup(cancel)
		_, err := client.Get(ctx, "/test")
		expectedErr := fmt.Sprintf("error Envoy admin URL http://127.0.0.1:%d/test: Get \"http://127.0.0.1:%d/test\": context deadline exceeded", client.port, client.port)
		require.EqualError(t, err, expectedErr)
	})

	t.Run("returns error on connection failure", func(t *testing.T) {
		client := &adminClient{port: 1} // port 1 should not be listening
		_, err := client.Get(t.Context(), "/test")
		require.EqualError(t, err, "error Envoy admin URL http://127.0.0.1:1/test: Get \"http://127.0.0.1:1/test\": dial tcp 127.0.0.1:1: connect: connection refused")
	})
}

func TestAdminClient_IsReady(t *testing.T) {
	var handler http.HandlerFunc
	client := setupTestServer(t, &handler)

	tests := []struct {
		name          string
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name: "returns nil when body is live",
			handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/ready", r.URL.Path)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("live"))
			},
		},
		{
			name: "returns error when body is not live",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("something else"))
			},
			expectedError: "unexpected /ready response body: \"something else\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler = tt.handler
			err := client.IsReady(t.Context())
			if tt.expectedError != "" {
				require.EqualError(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAdminClient_NewListenerRequest(t *testing.T) {
	var handler http.HandlerFunc
	client := setupTestServer(t, &handler)

	tests := []struct {
		name           string
		handler        http.HandlerFunc
		listenerName   string
		method         string
		path           string
		expectedPort   int
		expectedMethod string
		expectedError  string
	}{
		{
			name: "creates request when listener exists",
			handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/listeners", r.URL.Path)
				require.Equal(t, "json", r.URL.Query().Get("format"))
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{
                    "listener_statuses": [
                        {"name": "main", "local_address": {"socket_address": {"port_value": 8080}}},
                        {"name": "admin", "local_address": {"socket_address": {"port_value": 9901}}}
                    ]
                }`))
			},
			listenerName:   "main",
			method:         http.MethodGet,
			path:           "/path?query=value#fragment",
			expectedPort:   8080,
			expectedMethod: http.MethodGet,
		},
		{
			name: "returns error when listener not found",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{
                    "listener_statuses": [
                        {"name": "admin", "local_address": {"socket_address": {"port_value": 9901}}}
                    ]
                }`))
			},
			listenerName:  "nonexistent",
			method:        http.MethodGet,
			path:          "/",
			expectedError: "listener \"nonexistent\" not found",
		},
		{
			name: "returns error on invalid JSON response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("not valid json"))
			},
			listenerName:  "main",
			method:        http.MethodPost,
			path:          "/api/data",
			expectedError: "failed to parse Envoy listeners: invalid character 'o' in literal null (expecting 'u')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler = tt.handler
			req, err := client.NewListenerRequest(t.Context(), tt.listenerName, tt.method, tt.path, nil)
			if tt.expectedError != "" {
				require.EqualError(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, "http://127.0.0.1:"+strconv.Itoa(tt.expectedPort)+tt.path, req.URL.String())
				require.Equal(t, tt.expectedMethod, req.Method)
			}
		})
	}
}

func TestExtractFlagValue(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "admin-address.txt")

	tests := []struct {
		name          string
		flag          string
		cmdline       []string
		expected      string
		expectedError string
	}{
		{"valid flag with path", adminAddressPathFlag, []string{"envoy", adminAddressPathFlag, tmpDir}, tmpDir, ""},
		{"flag at end with path", adminAddressPathFlag, []string{"--config", "/etc/envoy.yaml", adminAddressPathFlag, tmpDir}, tmpDir, ""},
		{"flag not present", adminAddressPathFlag, []string{"envoy", "--config", "/etc/envoy.yaml"}, "", adminAddressPathFlag + " not found in command line"},
		{"flag present but no value", adminAddressPathFlag, []string{"envoy", adminAddressPathFlag}, "", adminAddressPathFlag + " not found in command line"},
		{"empty cmdline", adminAddressPathFlag, []string{}, "", adminAddressPathFlag + " not found in command line"},
		{"sh -c wrapped command", adminAddressPathFlag, []string{"sh", "-c", fmt.Sprintf("sleep 30 && echo %s %s", adminAddressPathFlag, tmpDir)}, tmpDir, ""},
		{"sh -c with multiple spaces", adminAddressPathFlag, []string{"sh", "-c", fmt.Sprintf("envoy %s %s --other-flag", adminAddressPathFlag, tmpDir)}, tmpDir, ""},
		{"admin address path flag", adminAddressPathFlag, []string{"envoy", adminAddressPathFlag, tmpFile}, tmpFile, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := extractFlagValue(tt.flag, tt.cmdline)
			if tt.expectedError != "" {
				require.EqualError(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, actual)
			}
		})
	}
}

func TestPollEnvoyPidAndAdminAddressPath(t *testing.T) {
	t.Run("success - finds envoy PID and defaults admin address path", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		t.Cleanup(cancel)

		adminAddressPath := path.Join(t.TempDir(), "admin-address.txt")

		cmdStr := fmt.Sprintf("sleep 30 && echo --admin-address-path %s", adminAddressPath)
		cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
		require.NoError(t, cmd.Start())
		t.Cleanup(func() {
			_ = cmd.Process.Kill()
			_, _ = cmd.Process.Wait()
		})

		time.Sleep(100 * time.Millisecond)

		actualEnvoyPid, actualAdminAddressPath, err := PollEnvoyPidAndAdminAddressPath(t.Context(), os.Getpid())
		require.NoError(t, err)
		require.Equal(t, cmd.Process.Pid, actualEnvoyPid)
		require.Equal(t, adminAddressPath, actualAdminAddressPath)
	})

	t.Run("failure - timeout waiting for Envoy process", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
		t.Cleanup(cancel)

		_, _, err := PollEnvoyPidAndAdminAddressPath(ctx, os.Getpid())
		require.EqualError(t, err, "timeout waiting for Envoy process: no Envoy process found")
	})
}

func TestNewAdminClient(t *testing.T) {
	adminAddressPath := filepath.Join(t.TempDir(), "admin-address.txt")

	tests := []struct {
		name          string
		setup         func(t *testing.T)
		ctx           func(t *testing.T) context.Context
		expectedError string
		expectedPort  int
	}{
		{
			name: "success - polls for admin port",
			setup: func(t *testing.T) {
				t.Helper()
				go func() {
					time.Sleep(100 * time.Millisecond)
					_ = os.WriteFile(adminAddressPath, []byte("127.0.0.1:9901"), 0o600)
				}()
			},
			ctx:          func(t *testing.T) context.Context { return t.Context() },
			expectedPort: 9901,
		},
		{
			name: "returns error when --admin-address-path has invalid content",
			setup: func(t *testing.T) {
				t.Helper()
				require.NoError(t, os.WriteFile(adminAddressPath, []byte("not-a-number"), 0o600))
			},
			ctx:           func(t *testing.T) context.Context { return t.Context() },
			expectedError: "failed to parse Envoy's admin port: strconv.Atoi: parsing \"\": invalid syntax",
		},
		{
			name: "returns error when admin address file never appears",
			ctx: func(t *testing.T) context.Context {
				t.Helper()
				ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
				t.Cleanup(cancel)
				return ctx
			},
			expectedError: "timeout waiting for Envoy admin address file: open " + adminAddressPath + ": no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up from previous test
			_ = os.RemoveAll(adminAddressPath)
			if tt.setup != nil {
				tt.setup(t)
			}
			client, err := NewAdminClient(tt.ctx(t), adminAddressPath)
			if tt.expectedError != "" {
				require.EqualError(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedPort, client.Port())
			}
		})
	}
}

func TestAdminClient_AwaitReady(t *testing.T) {
	var handler http.HandlerFunc
	client := setupTestServer(t, &handler)

	tests := []struct {
		name          string
		handler       func(callCount *int) http.HandlerFunc
		ctx           func(t *testing.T) context.Context
		interval      time.Duration
		expectedError string
		expectedCalls int
	}{
		{
			name: "returns nil when admin becomes ready after polling",
			handler: func(callCount *int) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, "/ready", r.URL.Path)
					*callCount++
					if *callCount < 3 {
						w.WriteHeader(http.StatusServiceUnavailable)
						_, _ = w.Write([]byte("not ready"))
					} else {
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write([]byte("live"))
					}
				}
			},
			ctx:           func(t *testing.T) context.Context { return t.Context() },
			interval:      10 * time.Millisecond,
			expectedCalls: 3,
		},
		{
			name: "returns context error when no IsReady calls made",
			handler: func(callCount *int) http.HandlerFunc {
				return func(w http.ResponseWriter, _ *http.Request) {}
			},
			ctx: func(t *testing.T) context.Context {
				t.Helper()
				ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
				t.Cleanup(cancel)
				return ctx
			},
			interval:      100 * time.Millisecond,
			expectedError: "context deadline exceeded",
		},
		{
			name: "returns immediately when already ready",
			handler: func(callCount *int) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, "/ready", r.URL.Path)
					*callCount++
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("live"))
				}
			},
			ctx:           func(t *testing.T) context.Context { return t.Context() },
			interval:      10 * time.Millisecond,
			expectedCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			handler = tt.handler(&callCount)

			err := client.AwaitReady(tt.ctx(t), tt.interval)
			if tt.expectedError != "" {
				// there's a temp file in the name
				require.EqualError(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
			}
			if tt.expectedCalls > 0 {
				require.Equal(t, tt.expectedCalls, callCount)
			}
		})
	}
}

func TestAdminClient_AwaitReady_ReturnsLastErrorOnTimeout(t *testing.T) {
	var handler http.HandlerFunc
	client := setupTestServer(t, &handler)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	callCount := 1

	handler = func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/ready", r.URL.Path)
		if callCount == 2 {
			cancel()
		}
		callCount++
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("still not ready"))
		w.(http.Flusher).Flush()
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- client.AwaitReady(ctx, 100*time.Millisecond)
	}()

	// The error should be the last IsReady error, not the context cancellation error
	err := <-errChan
	expectedErr := fmt.Sprintf("error Envoy admin URL http://127.0.0.1:%d/ready: status_code=503,body:still not ready", client.port)
	require.EqualError(t, err, expectedErr)
	require.GreaterOrEqual(t, callCount, 2)
}
