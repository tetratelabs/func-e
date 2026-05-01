// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/internal/test/httptest"
)

const readyPath = "/ready"

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func setupTestServer(t *testing.T, handler http.Handler) *adminClient {
	t.Helper()
	client, err := NewAdminClientForURL("http://"+ServerAddr, httptest.HandlerFactory(handler))
	require.NoError(t, err)
	actual, ok := client.(*adminClient)
	require.True(t, ok)
	return actual
}

func TestPollEnvoyPidAndAdminAddressPathForPort(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T, path string)
		ctx          func(t *testing.T) context.Context
		expectedPort int
		expectedErr  string
	}{
		{
			name: "file appears after delay",
			setup: func(t *testing.T, path string) {
				t.Helper()
				go func() {
					time.Sleep(100 * time.Millisecond)
					os.WriteFile(path, []byte("127.0.0.1:9901\n"), 0o600)
				}()
			},
			ctx:          func(t *testing.T) context.Context { t.Helper(); return t.Context() },
			expectedPort: 9901,
		},
		{
			name: "timeout when file never appears",
			setup: func(t *testing.T, _ string) {
				t.Helper()
			},
			ctx: func(t *testing.T) context.Context {
				t.Helper()
				ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
				t.Cleanup(cancel)
				return ctx
			},
			expectedErr: "timeout waiting for Envoy admin address file: open $ADMIN_ADDRESS_PATH: no such file or directory",
		},
		{
			name: "extracts port from address with any hostname",
			setup: func(t *testing.T, path string) {
				t.Helper()
				require.NoError(t, os.WriteFile(path, []byte("localhost:9901"), 0o600))
			},
			ctx:          func(t *testing.T) context.Context { t.Helper(); return t.Context() },
			expectedPort: 9901,
		},
		{
			name: "invalid address format",
			setup: func(t *testing.T, path string) {
				t.Helper()
				require.NoError(t, os.WriteFile(path, []byte("invalid-address"), 0o600))
			},
			ctx:         func(t *testing.T) context.Context { t.Helper(); return t.Context() },
			expectedErr: "failed to parse Envoy's admin port: strconv.Atoi: parsing \"\": invalid syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				adminPath := filepath.Join(t.TempDir(), "admin-address.txt")
				tt.setup(t, adminPath)
				actualPort, err := pollAdminAddressPathForPort(tt.ctx(t), adminPath)
				if tt.expectedErr != "" {
					expectedErr := strings.ReplaceAll(tt.expectedErr, "$ADMIN_ADDRESS_PATH", adminPath)
					require.EqualError(t, err, expectedErr)
				} else {
					require.NoError(t, err)
					require.Equal(t, tt.expectedPort, actualPort)
				}
			})
		})
	}
}

func TestPollAdminAddressPathForPort_PollsOnTickerBoundary(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		adminPath := filepath.Join(t.TempDir(), "admin-address.txt")
		resultCh := make(chan struct {
			port int
			err  error
		}, 1)

		go func() {
			port, err := pollAdminAddressPathForPort(t.Context(), adminPath)
			resultCh <- struct {
				port int
				err  error
			}{port: port, err: err}
		}()

		synctest.Wait()
		select {
		case res := <-resultCh:
			t.Fatalf("poll returned too early: port=%d err=%v", res.port, res.err)
		default:
		}

		time.Sleep(49 * time.Millisecond)
		synctest.Wait()
		select {
		case res := <-resultCh:
			t.Fatalf("poll returned before first tick: port=%d err=%v", res.port, res.err)
		default:
		}

		require.NoError(t, os.WriteFile(adminPath, []byte("127.0.0.1:9901\n"), 0o600))
		synctest.Wait()
		select {
		case res := <-resultCh:
			t.Fatalf("poll observed file before next tick: port=%d err=%v", res.port, res.err)
		default:
		}

		time.Sleep(1 * time.Millisecond)
		synctest.Wait()
		res := <-resultCh
		require.NoError(t, res.err)
		require.Equal(t, 9901, res.port)
	})
}

func TestAdminClient_get(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		ctx         func(t *testing.T) context.Context
		path        string
		expected    []byte
		expectedErr string
	}{
		{
			name: "returns body on 200 status",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("response body"))
			},
			ctx:      func(t *testing.T) context.Context { t.Helper(); return t.Context() },
			path:     "/test",
			expected: []byte("response body"),
		},
		{
			name: "returns error on non-200 status code",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte("not ready"))
			},
			ctx:         func(t *testing.T) context.Context { t.Helper(); return t.Context() },
			path:        "/test",
			expectedErr: `error Envoy admin URL $URL/test: status_code=503,body:not ready`,
		},
		{
			name: "respects context cancellation",
			handler: func(w http.ResponseWriter, r *http.Request) {
				<-r.Context().Done()
				w.WriteHeader(http.StatusOK)
			},
			ctx: func(t *testing.T) context.Context {
				t.Helper()
				ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
				t.Cleanup(cancel)
				return ctx
			},
			path:        "/test",
			expectedErr: `error Envoy admin URL $URL/test: Get "$URL/test": context deadline exceeded`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualMethod := ""
			actualPath := ""
			client := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				actualMethod = r.Method
				actualPath = r.URL.Path
				tt.handler(w, r)
			}))
			actual, err := client.Get(tt.ctx(t), tt.path)
			if tt.expectedErr != "" {
				expectedErr := strings.ReplaceAll(tt.expectedErr, "$URL", client.baseURL)
				require.EqualError(t, err, expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, actual)
			}
			require.Equal(t, http.MethodGet, actualMethod)
			require.Equal(t, tt.path, actualPath)
		})
	}

	t.Run("returns error on connection failure", func(t *testing.T) {
		client := newAdminClient(func() *http.Client {
			return &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return nil, errors.New("connection refused")
			})}
		}, "http://127.0.0.1:1", 1)
		_, err := client.Get(t.Context(), "/test")
		require.EqualError(t, err, "error Envoy admin URL http://127.0.0.1:1/test: Get \"http://127.0.0.1:1/test\": connection refused")
	})
}

func TestAdminClient_IsReady(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		expectedErr string
	}{
		{
			name: "returns nil when body is live",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(live))
			},
		},
		{
			name: "returns error when body is not live",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("something else"))
			},
			expectedErr: "unexpected /ready response body: \"something else\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualPath := ""
			client := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				actualPath = r.URL.Path
				tt.handler(w, r)
			}))
			err := client.IsReady(t.Context())
			if tt.expectedErr != "" {
				require.EqualError(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, readyPath, actualPath)
		})
	}
}

func TestAdminClient_NewListenerRequest(t *testing.T) {
	tests := []struct {
		name           string
		listenerName   string
		method         string
		path           string
		body           string
		expectedPort   int
		expectedMethod string
		expectedErr    string
	}{
		{
			name: "creates request when listener exists",
			body: `{
                    "listener_statuses": [
                        {"name": "main", "local_address": {"socket_address": {"port_value": 8080}}},
                        {"name": "admin", "local_address": {"socket_address": {"port_value": 9901}}}
                    ]
                }`,
			listenerName:   "main",
			method:         http.MethodGet,
			path:           "/path?query=value#fragment",
			expectedPort:   8080,
			expectedMethod: http.MethodGet,
		},
		{
			name: "returns error when listener not found",
			body: `{
                    "listener_statuses": [
                        {"name": "admin", "local_address": {"socket_address": {"port_value": 9901}}}
                    ]
                }`,
			listenerName: "nonexistent",
			method:       http.MethodGet,
			path:         "/",
			expectedErr:  "listener \"nonexistent\" not found",
		},
		{
			name:         "returns error on invalid JSON response",
			body:         "not valid json",
			listenerName: "main",
			method:       http.MethodPost,
			path:         "/api/data",
			expectedErr:  "failed to parse Envoy listeners: invalid character 'o' in literal null (expecting 'u')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualMethod := ""
			actualPath := ""
			actualFormat := ""
			client := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				actualMethod = r.Method
				actualPath = r.URL.Path
				actualFormat = r.URL.Query().Get("format")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.body))
			}))

			req, err := client.NewListenerRequest(t.Context(), tt.listenerName, tt.method, tt.path, http.NoBody)
			if tt.expectedErr != "" {
				require.EqualError(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, "http://127.0.0.1:"+strconv.Itoa(tt.expectedPort)+tt.path, req.URL.String())
				require.Equal(t, tt.expectedMethod, req.Method)
			}
			require.Equal(t, http.MethodGet, actualMethod)
			require.Equal(t, "/listeners", actualPath)
			require.Equal(t, "json", actualFormat)
		})
	}
}

func TestExtractFlagValue(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "admin-address.txt")
	pathWithSpaces := filepath.Join(tmpDir, "admin address.txt")

	tests := []struct {
		name        string
		flag        string
		cmdline     []string
		expected    string
		expectedErr string
	}{
		{"valid flag with path", AddressPathFlag, []string{"envoy", AddressPathFlag, tmpDir}, tmpDir, ""},
		{"valid flag with path containing spaces", AddressPathFlag, []string{"envoy", AddressPathFlag, pathWithSpaces}, pathWithSpaces, ""},
		{"valid equals flag", AddressPathFlag, []string{"envoy", AddressPathFlag + "=" + tmpFile}, tmpFile, ""},
		{"flag at end with path", AddressPathFlag, []string{"--config", "/etc/envoy.yaml", AddressPathFlag, tmpDir}, tmpDir, ""},
		{"flag not present", AddressPathFlag, []string{"envoy", "--config", "/etc/envoy.yaml"}, "", AddressPathFlag + " not found in command line"},
		{"flag present but no value", AddressPathFlag, []string{"envoy", AddressPathFlag}, "", AddressPathFlag + " not found in command line"},
		{"empty cmdline", AddressPathFlag, []string{}, "", AddressPathFlag + " not found in command line"},
		{"sh -c wrapped command", AddressPathFlag, []string{"sh", "-c", fmt.Sprintf("sleep 30 && echo %s %s", AddressPathFlag, tmpDir)}, tmpDir, ""},
		{"sh -c with multiple spaces", AddressPathFlag, []string{"sh", "-c", fmt.Sprintf("envoy %s %s --other-flag", AddressPathFlag, tmpDir)}, tmpDir, ""},
		{"sh -c with equals flag", AddressPathFlag, []string{"sh", "-c", fmt.Sprintf("envoy %s=%s --other-flag", AddressPathFlag, tmpFile)}, tmpFile, ""},
		{"admin address path flag", AddressPathFlag, []string{"envoy", AddressPathFlag, tmpFile}, tmpFile, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := extractFlagValue(tt.flag, tt.cmdline)
			if tt.expectedErr != "" {
				require.EqualError(t, err, tt.expectedErr)
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

		cmdStr := "sleep 30 && echo --admin-address-path " + adminAddressPath
		cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
		require.NoError(t, cmd.Start())
		t.Cleanup(func() {
			cmd.Process.Kill()
			cmd.Process.Wait()
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
	tests := []struct {
		name         string
		setup        func(t *testing.T, adminAddressPath string)
		ctx          func(t *testing.T) context.Context
		expectedErr  string
		expectedPort int
	}{
		{
			name: "success - polls for admin port",
			setup: func(t *testing.T, adminAddressPath string) {
				t.Helper()
				go func() {
					time.Sleep(100 * time.Millisecond)
					os.WriteFile(adminAddressPath, []byte(ServerAddr), 0o600)
				}()
			},
			ctx:          func(t *testing.T) context.Context { t.Helper(); return t.Context() },
			expectedPort: 9901,
		},
		{
			name: "returns error when --admin-address-path has invalid content",
			setup: func(t *testing.T, adminAddressPath string) {
				t.Helper()
				require.NoError(t, os.WriteFile(adminAddressPath, []byte("not-a-number"), 0o600))
			},
			ctx:         func(t *testing.T) context.Context { t.Helper(); return t.Context() },
			expectedErr: "failed to parse Envoy's admin port: strconv.Atoi: parsing \"\": invalid syntax",
		},
		{
			name: "returns error when admin address file never appears",
			ctx: func(t *testing.T) context.Context {
				t.Helper()
				ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
				t.Cleanup(cancel)
				return ctx
			},
			expectedErr: "timeout waiting for Envoy admin address file: open $ADMIN_ADDRESS_PATH: no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				adminAddressPath := filepath.Join(t.TempDir(), "admin-address.txt")
				if tt.setup != nil {
					tt.setup(t, adminAddressPath)
				}
				client, err := NewAdminClient(tt.ctx(t), httptest.HandlerFactory(http.NotFoundHandler()), adminAddressPath)
				if tt.expectedErr != "" {
					expectedErr := strings.ReplaceAll(tt.expectedErr, "$ADMIN_ADDRESS_PATH", adminAddressPath)
					require.EqualError(t, err, expectedErr)
				} else {
					require.NoError(t, err)
					require.Equal(t, tt.expectedPort, client.Port())
				}
			})
		})
	}
}

func TestAdminClient_AwaitReady(t *testing.T) {
	tests := []struct {
		name          string
		body          func(callCount int) string
		statusCode    func(callCount int) int
		ctx           func(t *testing.T) context.Context
		interval      time.Duration
		expectedErr   string
		expectedCalls int
	}{
		{
			name: "returns nil when admin becomes ready after polling",
			body: func(callCount int) string {
				if callCount < 3 {
					return "not ready"
				}
				return live
			},
			statusCode: func(callCount int) int {
				if callCount < 3 {
					return http.StatusServiceUnavailable
				}
				return http.StatusOK
			},
			ctx:           func(t *testing.T) context.Context { t.Helper(); return t.Context() },
			interval:      10 * time.Millisecond,
			expectedCalls: 3,
		},
		{
			name: "returns context error when no IsReady calls made",
			body: func(int) string { return live },
			statusCode: func(int) int {
				return http.StatusOK
			},
			ctx: func(t *testing.T) context.Context {
				t.Helper()
				ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
				t.Cleanup(cancel)
				return ctx
			},
			interval:    100 * time.Millisecond,
			expectedErr: "context deadline exceeded",
		},
		{
			name: "returns immediately when already ready",
			body: func(int) string { return live },
			statusCode: func(int) int {
				return http.StatusOK
			},
			ctx:           func(t *testing.T) context.Context { t.Helper(); return t.Context() },
			interval:      10 * time.Millisecond,
			expectedCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				callCount := 0
				methods := []string{}
				paths := []string{}
				client := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					callCount++
					methods = append(methods, r.Method)
					paths = append(paths, r.URL.Path)
					w.WriteHeader(tt.statusCode(callCount))
					_, _ = w.Write([]byte(tt.body(callCount)))
				}))

				err := client.AwaitReady(tt.ctx(t), tt.interval)
				if tt.expectedErr != "" {
					require.EqualError(t, err, tt.expectedErr)
				} else {
					require.NoError(t, err)
				}
				require.Equal(t, tt.expectedCalls, callCount)
				require.Len(t, methods, callCount)
				require.Len(t, paths, callCount)
				for i := range methods {
					require.Equal(t, http.MethodGet, methods[i])
					require.Equal(t, readyPath, paths[i])
				}
			})
		})
	}
}

func TestAdminClient_AwaitReady_ReturnsLastErrorOnTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		t.Cleanup(cancel)

		callCount := 0
		methods := []string{}
		paths := []string{}
		client := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 2 {
				cancel()
			}
			methods = append(methods, r.Method)
			paths = append(paths, r.URL.Path)
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("still not ready"))
		}))

		err := client.AwaitReady(ctx, 100*time.Millisecond)
		expectedErr := fmt.Sprintf("error Envoy admin URL %s/ready: status_code=503,body:still not ready", client.baseURL)
		require.EqualError(t, err, expectedErr)
		require.Equal(t, 2, callCount)
		require.Len(t, methods, 2)
		require.Len(t, paths, 2)
		for i := range methods {
			require.Equal(t, http.MethodGet, methods[i])
			require.Equal(t, readyPath, paths[i])
		}
	})
}

func TestAdminClient_AwaitReady_FirstPollOnFirstTick(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		callCount := 0
		actualMethod := ""
		actualPath := ""
		client := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			actualMethod = r.Method
			actualPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(live))
		}))

		errCh := make(chan error, 1)
		go func() {
			errCh <- client.AwaitReady(t.Context(), time.Second)
		}()

		synctest.Wait()
		require.Zero(t, callCount)
		select {
		case err := <-errCh:
			t.Fatalf("AwaitReady returned before first tick: %v", err)
		default:
		}

		time.Sleep(time.Second - time.Nanosecond)
		synctest.Wait()
		require.Zero(t, callCount)
		select {
		case err := <-errCh:
			t.Fatalf("AwaitReady returned before first tick: %v", err)
		default:
		}

		time.Sleep(1 * time.Nanosecond)
		synctest.Wait()
		require.Equal(t, 1, callCount)
		require.Equal(t, http.MethodGet, actualMethod)
		require.Equal(t, readyPath, actualPath)
		require.NoError(t, <-errCh)
	})
}
