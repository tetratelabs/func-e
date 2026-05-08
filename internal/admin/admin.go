// Copyright func-e contributors
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/process"

	internalapi "github.com/tetratelabs/func-e/internal/api"
)

const (
	ServerAddr      = "127.0.0.1:9901"
	AddressPathFlag = "--admin-address-path"
	runIDFlag       = "--run-id"
	live            = "live"
	pollInterval    = 50 * time.Millisecond
)

var errMultipleEnvoyProcesses = fmt.Errorf("multiple Envoy processes found; set %s to disambiguate", runIDFlag)

// NewAdminClient creates an AdminClient by polling for the admin port at
// adminAddressPath.
func NewAdminClient(ctx context.Context, client *http.Client, adminAddressPath string) (internalapi.AdminClient, error) {
	// Envoy writes its admin address after startup, so this blocks until the
	// port is available or the caller's context is done.
	port, err := pollAdminAddressPathForPort(ctx, adminAddressPath)
	if err != nil {
		return nil, err
	}
	return newAdminClient(client, fmt.Sprintf("http://127.0.0.1:%d", port), port), nil
}

// NewAdminClientForURL creates an AdminClient for the given base URL and HTTP client factory.
func NewAdminClientForURL(baseURL string, client *http.Client) (internalapi.AdminClient, error) {
	baseURL = strings.TrimRight(baseURL, "/")
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Envoy admin address: %w", err)
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, fmt.Errorf("failed to parse Envoy admin port: %w", err)
	}
	return newAdminClient(client, baseURL, port), nil
}

func newAdminClient(client *http.Client, baseURL string, port int) *adminClient {
	return &adminClient{
		baseURL:    baseURL,
		httpClient: client,
		port:       port,
	}
}

var _ internalapi.AdminClient = (*adminClient)(nil)

// adminClient checks Envoy readiness via the admin API /ready endpoint.
type adminClient struct {
	baseURL    string
	httpClient *http.Client
	port       int
}

// Port implements api.AdminClient.
func (c *adminClient) Port() int {
	return c.port
}

// Do implements api.AdminClient.
func (c *adminClient) Do(req *http.Request) (*http.Response, error) {
	// #nosec G704 -- requests executed through AdminClient target Envoy admin/listener URLs.
	return c.httpClient.Do(req)
}

// IsReady implements api.AdminClient.
func (c *adminClient) IsReady(ctx context.Context) error {
	body, err := c.Get(ctx, "/ready")
	if err != nil {
		return err
	}
	if body := strings.ToLower(strings.TrimSpace(string(body))); body != live {
		return fmt.Errorf("unexpected /ready response body: %q", body)
	}
	return nil
}

// AwaitReady implements api.AdminClient.
func (c *adminClient) AwaitReady(ctx context.Context, tickDuration time.Duration) error {
	ticker := time.NewTicker(tickDuration)
	defer ticker.Stop()

	var lastErr error
	for {
		select {
		case <-ctx.Done():
			// If Envoy answered but never became ready, the last readiness
			// failure is more useful than the polling deadline.
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

// NewListenerRequest implements api.AdminClient.
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

// Get implements api.AdminClient.
func (c *adminClient) Get(ctx context.Context, path string) ([]byte, error) {
	endpoint := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error Envoy admin URL %s: %w", endpoint, err)
	}
	defer resp.Body.Close() //nolint:errcheck // body fully read below

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error Envoy admin URL %s: status_code=%d,body:%s", endpoint, resp.StatusCode, body)
	}
	return body, nil
}

// pollAdminAddressPathForPort polls for the admin-address.txt file.
// It returns the admin port number or an error if the timeout is reached.
func pollAdminAddressPathForPort(ctx context.Context, adminAddressPath string) (int, error) {
	ticker := time.NewTicker(pollInterval)
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
			data, err := os.ReadFile(adminAddressPath) //nolint:gosec // path comes from our own --admin-address-path flag
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

	return parseAdminPort(adminAddr)
}

func parseAdminPort(addr string) (int, error) {
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse Envoy's admin address: %w", err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse Envoy's admin port: %w", err)
	}
	return port, nil
}

// extractAdminAddressPath returns the first match before [internalapi.ArgsIgnoreRest].
func extractAdminAddressPath(cmdline []string) (string, error) {
	for i := range len(cmdline) {
		arg := cmdline[i]
		if arg == internalapi.ArgsIgnoreRest {
			break
		}
		switch {
		case arg == AddressPathFlag && i+1 < len(cmdline) && cmdline[i+1] != "":
			return cmdline[i+1], nil
		case strings.HasPrefix(arg, AddressPathFlag+"="):
			if value := strings.TrimPrefix(arg, AddressPathFlag+"="); value != "" {
				return value, nil
			}
		}
	}

	// Shell wrappers expose the wrapped command as one argv entry. Keep this
	// fallback after the argv-preserving scan so direct args can contain spaces.
	if len(cmdline) >= 3 && cmdline[1] == "-c" {
		fields := strings.Fields(cmdline[2])
		for i := range len(fields) {
			arg := fields[i]
			if arg == internalapi.ArgsIgnoreRest {
				break
			}
			switch {
			case arg == AddressPathFlag && i+1 < len(fields) && fields[i+1] != "":
				return fields[i+1], nil
			case strings.HasPrefix(arg, AddressPathFlag+"="):
				if value := strings.TrimPrefix(arg, AddressPathFlag+"="); value != "" {
					return value, nil
				}
			}
		}
	}

	return "", fmt.Errorf("%s not found in command line", AddressPathFlag)
}

// extractRunID returns the last match, so the func-e-appended value after [internalapi.ArgsIgnoreRest] wins.
func extractRunID(cmdline []string) (string, error) {
	var runID string
	for i := 0; i < len(cmdline); i++ {
		arg := cmdline[i]
		switch {
		case arg == runIDFlag && i+1 < len(cmdline) && cmdline[i+1] != "":
			runID = cmdline[i+1]
			i++
		case strings.HasPrefix(arg, runIDFlag+"="):
			if value := strings.TrimPrefix(arg, runIDFlag+"="); value != "" {
				runID = value
			}
		}
	}
	if runID != "" {
		return runID, nil
	}

	if len(cmdline) >= 3 && cmdline[1] == "-c" {
		fields := strings.Fields(cmdline[2])
		for i := 0; i < len(fields); i++ {
			arg := fields[i]
			switch {
			case arg == runIDFlag && i+1 < len(fields) && fields[i+1] != "":
				runID = fields[i+1]
				i++
			case strings.HasPrefix(arg, runIDFlag+"="):
				if value := strings.TrimPrefix(arg, runIDFlag+"="); value != "" {
					runID = value
				}
			}
		}
	}
	if runID != "" {
		return runID, nil
	}

	return "", fmt.Errorf("%s not found in command line", runIDFlag)
}

type envoyProcessCandidate struct {
	pid     int
	cmdline []string
}

func selectEnvoyProcess(candidates []envoyProcessCandidate, runID string) (envoyPid int, adminAddressPath string, err error) {
	if runID == "" {
		// Without an explicit runID, skip children that lack the func-e marker.
		foundRunID := false
		for _, candidate := range candidates {
			if _, err := extractRunID(candidate.cmdline); err != nil {
				continue
			}
			foundRunID = true

			path, err := extractAdminAddressPath(candidate.cmdline)
			if err != nil {
				continue
			}
			// Keep scanning past the first match to detect ambiguity.
			if adminAddressPath != "" {
				return 0, "", errMultipleEnvoyProcesses
			}
			envoyPid = candidate.pid
			adminAddressPath = path
		}
		if adminAddressPath != "" {
			return envoyPid, adminAddressPath, nil
		}
		if !foundRunID {
			return 0, "", fmt.Errorf("no child with %s", runIDFlag)
		}
		return 0, "", fmt.Errorf("no child with %s", AddressPathFlag)
	}

	for _, candidate := range candidates {
		id, err := extractRunID(candidate.cmdline)
		if err != nil || id != runID {
			continue
		}
		adminAddressPath, err := extractAdminAddressPath(candidate.cmdline)
		if err != nil {
			return 0, "", err
		}
		return candidate.pid, adminAddressPath, nil
	}
	return 0, "", fmt.Errorf("no child with %s %s", runIDFlag, runID)
}

// PollEnvoyPidAndAdminAddressPath polls for a child process tagged with
// runID and extracts its pid and admin address path from its command line.
//
// This polls because the child may not exist yet when the caller starts.
func PollEnvoyPidAndAdminAddressPath(ctx context.Context, funcEPid int, runID string) (envoyPid int, adminAddressPath string, err error) {
	funcEProc, err := process.NewProcessWithContext(ctx, int32(funcEPid)) //nolint:gosec // funcEPid never overflows int32
	if err != nil {
		return 0, "", fmt.Errorf("failed to get func-e process: %w", err)
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	var lastErr error
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

			candidates := make([]envoyProcessCandidate, 0, len(children))
			for _, child := range children {
				cmdline, err := child.CmdlineSliceWithContext(ctx)
				if err != nil {
					continue
				}
				candidates = append(candidates, envoyProcessCandidate{
					pid:     int(child.Pid),
					cmdline: cmdline,
				})
			}

			envoyPid, adminAddressPath, err = selectEnvoyProcess(candidates, runID)
			if err == nil {
				return envoyPid, adminAddressPath, nil
			}
			// Ambiguity won't resolve with more polling; the caller needs a runID.
			if errors.Is(err, errMultipleEnvoyProcesses) {
				return 0, "", err
			}
			lastErr = err
		}
	}
}
