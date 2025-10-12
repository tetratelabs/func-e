// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/process"

	internalapi "github.com/tetratelabs/func-e/internal/api"
)

const adminAddressPathFlag = `--admin-address-path`

// NewAdminClient creates an AdminClient by polling for the admin port at
// adminAddressPath.
func NewAdminClient(ctx context.Context, adminAddressPath string) (internalapi.AdminClient, error) {
	// Block until admin port is available
	port, err := pollAdminAddressPathForPort(ctx, adminAddressPath)
	if err != nil {
		return nil, err
	}

	return &adminClient{port: port}, nil
}

// adminClient checks Envoy readiness via the admin API /ready endpoint.
type adminClient struct {
	port int
}

// Port implements the same method as documented on api.AdminClient
func (c *adminClient) Port() int {
	return c.port
}

// IsReady implements the same method as documented on api.AdminClient
func (c *adminClient) IsReady(ctx context.Context) error {
	body, err := c.Get(ctx, "/ready")
	if err != nil {
		return err
	}
	if body := strings.ToLower(strings.TrimSpace(string(body))); body != "live" {
		return fmt.Errorf("unexpected /ready response body: %q", body)
	}
	return nil
}

// AwaitReady implements the same method as documented on api.AdminClient
func (c *adminClient) AwaitReady(ctx context.Context, tickDuration time.Duration) error {
	ticker := time.NewTicker(tickDuration)
	defer ticker.Stop()

	var lastErr error
	for {
		select {
		case <-ctx.Done():
			// Prioritize the last IsReady error over context error
			if lastErr != nil {
				return lastErr
			}
			return ctx.Err()
		case <-ticker.C:
			if err := c.IsReady(ctx); err == nil {
				return nil
			} else if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return lastErr
			} else {
				lastErr = err
			}
		}
	}
}

type listenersResponse struct {
	ListenerStatuses []listenerStatus `json:"listener_statuses"`
}

type listenerStatus struct {
	Name         string       `json:"name"`
	LocalAddress localAddress `json:"local_address"`
}

type localAddress struct {
	SocketAddress socketAddress `json:"socket_address"`
}

type socketAddress struct {
	PortValue int `json:"port_value"`
}

// NewListenerRequest implements the same method as documented on api.AdminClient
func (c *adminClient) NewListenerRequest(ctx context.Context, name, method, path string, body io.Reader) (*http.Request, error) {
	respBody, err := c.Get(ctx, "/listeners?format=json")
	if err != nil {
		return nil, err
	}

	var lr listenersResponse
	if err := json.Unmarshal(respBody, &lr); err != nil {
		return nil, fmt.Errorf("failed to parse Envoy listeners: %w", err)
	}

	var port int
	for _, ls := range lr.ListenerStatuses {
		if ls.Name == name {
			port = ls.LocalAddress.SocketAddress.PortValue
			break
		}
	}
	if port == 0 {
		return nil, fmt.Errorf("listener %q not found", name)
	}

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	return http.NewRequestWithContext(ctx, method, baseURL+path, body)
}

// Get implements the same method as documented on api.AdminClient
func (c *adminClient) Get(ctx context.Context, path string) ([]byte, error) {
	url := fmt.Sprintf("http://127.0.0.1:%d%s", c.port, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d%s", c.port, path), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error Envoy admin URL %s: %w", url, err)
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error Envoy admin URL %s: status_code=%d,body:%s", url, resp.StatusCode, body)
	}
	return body, nil
}

// pollAdminAddressPathForPort polls for the admin-address.txt file.
// It returns the admin port number or an error if the timeout is reached.
func pollAdminAddressPathForPort(ctx context.Context, adminAddressPath string) (int, error) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	var adminAddr string
	var lastErr error
LOOP:
	for {
		select {
		case <-ctx.Done():
			if lastErr == nil {
				return 0, fmt.Errorf("timeout waiting for Envoy admin address file %s", adminAddressPath)
			}
			return 0, fmt.Errorf("timeout waiting for Envoy admin address file: %w", lastErr)
		case <-ticker.C:
			data, err := os.ReadFile(adminAddressPath)
			if err != nil {
				lastErr = err
				continue
			}

			adminAddr = strings.TrimSpace(string(data))
			if adminAddr == "" {
				lastErr = fmt.Errorf("envoy admin address file %s was empty", adminAddressPath)
				continue
			}
			break LOOP
		}
	}

	// Parse as URL to extract port
	u, err := url.Parse("http://" + adminAddr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse Envoy's admin address: %w", err)
	}

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return 0, fmt.Errorf("failed to parse Envoy's admin port: %w", err)
	}

	return port, nil
}

// extractFlagValue parses a flag value from command line arguments.
func extractFlagValue(flag string, cmdline []string) (string, error) {
	// Join cmdline into a single string and split by spaces to handle sh -c
	// cases (these cases are only used in tests).
	fullCmd := strings.Join(cmdline, " ")
	parts := strings.Fields(fullCmd)

	for i, arg := range parts {
		if arg == flag && i+1 < len(parts) {
			return parts[i+1], nil
		}
	}
	return "", fmt.Errorf("%s not found in command line", flag)
}

// PollEnvoyPidAndAdminAddressPath polls for the Envoy child process and
// extracts its pid and admin address path from its command line.
//
// This polls as the goroutine may be called prior to the Envoy subprocess.
func PollEnvoyPidAndAdminAddressPath(ctx context.Context, funcEPid int) (envoyPid int, adminAddressPath string, err error) {
	funcEProc, err := process.NewProcessWithContext(ctx, int32(funcEPid))
	if err != nil {
		return 0, "", fmt.Errorf("failed to Get func-e process: %w", err)
	}

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	var envoyProc *process.Process
	var lastErr error
LOOP:
	for {
		select {
		case <-ctx.Done():
			if lastErr == nil {
				return 0, "", errors.New("timeout waiting for Envoy process")
			}
			return 0, "", fmt.Errorf("timeout waiting for Envoy process: %w", lastErr)
		case <-ticker.C:
			children, childErr := funcEProc.ChildrenWithContext(ctx)
			if childErr != nil {
				lastErr = childErr
				continue
			}

			if len(children) == 0 {
				lastErr = errors.New("no Envoy process found")
				continue
			}

			// Assume the first child is the Envoy process
			envoyProc = children[0]
			envoyPid = int(envoyProc.Pid)
			break LOOP
		}
	}

	// Get command line args
	envoyCmdline, err := envoyProc.CmdlineSlice()
	if err != nil {
		return 0, "", fmt.Errorf("failed to Get command line of Envoy: %w", err)
	}

	// Extract admin address path
	adminAddressPath, err = extractFlagValue(adminAddressPathFlag, envoyCmdline)
	if err != nil {
		return 0, "", err
	}

	return envoyPid, adminAddressPath, nil
}
